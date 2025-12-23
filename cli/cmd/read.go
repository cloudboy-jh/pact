package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/pact/internal/auth"
	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/detect"
	"github.com/cloudboy-jh/pact/internal/git"
	"github.com/cloudboy-jh/pact/internal/keyring"
	"github.com/cloudboy-jh/pact/internal/ui"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var (
	flagDiff   bool
	flagJSON   bool
	flagYes    bool
	flagDryRun bool
)

var readCmd = &cobra.Command{
	Use:   "read [modules...]",
	Short: "Scan local environment and import to pact",
	Long: `Scan your development environment for installed tools, 
configurations, and settings. Optionally import them into pact.json.

This is the reverse of 'pact sync' - instead of applying pact.json to your
machine, it detects what's on your machine and imports it into pact.json.

Examples:
  pact read                  # Interactive scan and import
  pact read cli shell        # Only scan specific modules
  pact read --diff           # Show what differs from pact.json
  pact read --json           # Output as JSON (no prompts)
  pact read -y               # Import everything without prompts
  pact read --dry-run        # Preview without modifying anything`,
	Run: runRead,
}

func init() {
	readCmd.Flags().BoolVar(&flagDiff, "diff", false, "Only show differences from pact.json")
	readCmd.Flags().BoolVar(&flagJSON, "json", false, "Output detected config as JSON")
	readCmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Import all detected items without prompting")
	readCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Preview changes without modifying anything")

	rootCmd.AddCommand(readCmd)
}

func runRead(cmd *cobra.Command, args []string) {
	// Check if pact is initialized
	if !config.Exists() {
		if !promptGitHubConnect() {
			return
		}
	}

	fmt.Println()
	fmt.Println("Scanning your development environment...")
	fmt.Println()

	// Scan environment
	opts := detect.ScanOptions{
		Modules:      args,
		IncludeFiles: true,
	}
	detected := detect.Scan(opts)

	// If --json flag, output JSON and exit
	if flagJSON {
		output, err := json.MarshalIndent(detected, "", "  ")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
		return
	}

	// Load existing config for comparison
	var existingCfg *config.PactConfig
	if config.Exists() {
		var err error
		existingCfg, err = config.Load()
		if err != nil {
			fmt.Printf("Warning: Could not load existing pact.json: %v\n", err)
		}
	}

	// Get secrets from existing config for comparison
	var existingSecrets []string
	if existingCfg != nil {
		existingSecrets = existingCfg.GetSecrets()
	}

	// Re-scan secrets with existing secrets for comparison
	detected.Secrets = detect.DetectSecrets(existingSecrets)

	// Update keychain status for secrets
	for i := range detected.Secrets {
		detected.Secrets[i].InKeychain = keyring.HasSecret(detected.Secrets[i].Name)
	}

	// Compare with existing config
	var diffs []detect.DiffResult
	if existingCfg != nil {
		diffs = detect.Compare(detected, existingCfg)
	} else {
		// No existing config - everything is "local only"
		diffs = createAllLocalDiffs(detected)
	}

	// Render the diff
	renderDiffs(diffs, existingCfg != nil)

	// If --diff flag, just show diffs and exit
	if flagDiff {
		return
	}

	// Count new items
	newCount := detect.CountNewItems(diffs)
	if newCount == 0 {
		fmt.Println("\nNo new items to import.")
		return
	}

	fmt.Printf("\nFound %d item(s) that can be imported.\n", newCount)

	// If --dry-run, show what would be imported and exit
	if flagDryRun {
		fmt.Println("\n[Dry run] Would import the above items.")
		return
	}

	// If --yes flag, import all
	if flagYes {
		importAll(detected, diffs)
		return
	}

	// Run interactive TUI picker
	p := tea.NewProgram(initialReadModel(detected, diffs))
	result, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Process selection
	if m, ok := result.(readModel); ok && !m.cancelled {
		applySelection(m.selected, detected)
	}
}

// promptGitHubConnect prompts user to connect GitHub and initialize pact
func promptGitHubConnect() bool {
	fmt.Println(ui.RenderLogo())
	fmt.Println()
	fmt.Println("Pact is not initialized.")
	fmt.Print("Would you like to connect GitHub and create your pact repo? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "" && response != "y" && response != "yes" {
		fmt.Println("Cancelled.")
		return false
	}

	// Run the init flow
	return runInitFlow()
}

// runInitFlow runs the GitHub auth and repo setup (extracted from init.go)
func runInitFlow() bool {
	// Check if we already have a token
	if keyring.HasToken() {
		fmt.Println("Found existing GitHub token. Verifying...")
		token, _ := keyring.GetToken()
		user, err := auth.GetUser(token)
		if err == nil {
			fmt.Printf("Authenticated as %s\n", user.Login)
			return setupPactRepo(token, user.Login)
		}
		fmt.Println("Token expired or invalid. Re-authenticating...")
		keyring.DeleteToken()
	}

	// Start device flow
	fmt.Println("Authenticating with GitHub...")
	fmt.Println()

	deviceCode, err := auth.RequestDeviceCode()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	fmt.Printf("Please visit: %s\n", deviceCode.VerificationURI)
	fmt.Printf("And enter code: %s\n", deviceCode.UserCode)
	fmt.Println()
	fmt.Println("Waiting for authorization...")

	// Try to open browser
	browser.OpenURL(deviceCode.VerificationURI)

	// Poll for token
	token, err := auth.PollForToken(deviceCode.DeviceCode, deviceCode.Interval)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	// Get user info
	user, err := auth.GetUser(token)
	if err != nil {
		fmt.Printf("Error getting user info: %v\n", err)
		return false
	}

	fmt.Printf("\n✓ Authenticated as %s\n", user.Login)

	// Store token
	if err := keyring.SetToken(token); err != nil {
		fmt.Printf("Warning: Could not store token in keychain: %v\n", err)
	}

	return setupPactRepo(token, user.Login)
}

// setupPactRepo creates the repo and clones it
func setupPactRepo(token, username string) bool {
	// Check if repo exists
	fmt.Printf("Checking for %s/my-pact repo...\n", username)
	exists, err := auth.RepoExists(token, username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	if !exists {
		fmt.Println("Repo not found. Creating...")
		if err := auth.CreateRepo(token); err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}
		fmt.Println("✓ Created my-pact repo")
		time.Sleep(2 * time.Second)
	}

	// Get local pact directory
	pactDir, err := config.GetLocalPactDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	// Clone repo
	fmt.Println("Cloning to ./.pact/...")
	if err := git.Clone(token, username, pactDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}
	fmt.Println("✓ Cloned repo to ./.pact/")

	return true
}

// createAllLocalDiffs creates diffs where everything is local-only (for new pact.json)
func createAllLocalDiffs(detected *detect.DetectedConfig) []detect.DiffResult {
	var diffs []detect.DiffResult

	// CLI
	if len(detected.CLI.Tools) > 0 || len(detected.CLI.Custom) > 0 {
		diff := detect.DiffResult{Module: "cli"}
		for _, t := range detected.CLI.Tools {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: t, Type: "tool"})
		}
		for _, t := range detected.CLI.Custom {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: t, Type: "custom"})
		}
		diffs = append(diffs, diff)
	}

	// Shell
	if detected.Shell.Prompt != nil || len(detected.Shell.Tools) > 0 {
		diff := detect.DiffResult{Module: "shell"}
		if detected.Shell.Prompt != nil {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{
				Name:  detected.Shell.Prompt.Tool,
				Type:  "prompt",
				Value: detected.Shell.Prompt.Theme,
			})
		}
		for _, t := range detected.Shell.Tools {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: t, Type: "tool"})
		}
		diffs = append(diffs, diff)
	}

	// Git
	if detected.Git.User != "" || detected.Git.Email != "" {
		diff := detect.DiffResult{Module: "git"}
		if detected.Git.User != "" {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: "user", Type: "setting", Value: detected.Git.User})
		}
		if detected.Git.Email != "" {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: "email", Type: "setting", Value: detected.Git.Email})
		}
		if detected.Git.DefaultBranch != "" {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: "defaultBranch", Type: "setting", Value: detected.Git.DefaultBranch})
		}
		if detected.Git.LFS {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: "lfs", Type: "setting", Value: true})
		}
		diffs = append(diffs, diff)
	}

	// Editor
	if detected.Editor.Default != "" {
		diff := detect.DiffResult{Module: "editor"}
		diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: detected.Editor.Default, Type: "editor"})
		diffs = append(diffs, diff)
	}

	// LLM
	if len(detected.LLM.Providers) > 0 || detected.LLM.Local != nil {
		diff := detect.DiffResult{Module: "llm"}
		for _, p := range detected.LLM.Providers {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: p, Type: "provider"})
		}
		if detected.LLM.Local != nil {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: detected.LLM.Local.Runtime, Type: "runtime"})
			for _, m := range detected.LLM.Local.Models {
				diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: m, Type: "model"})
			}
		}
		if detected.LLM.Coding != nil {
			for _, a := range detected.LLM.Coding.Agents {
				diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: a, Type: "agent"})
			}
		}
		diffs = append(diffs, diff)
	}

	// Secrets
	if len(detected.Secrets) > 0 {
		diff := detect.DiffResult{Module: "secrets"}
		for _, s := range detected.Secrets {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: s.Name, Type: "secret"})
		}
		diffs = append(diffs, diff)
	}

	// Config files
	if len(detected.ConfigFiles) > 0 {
		diff := detect.DiffResult{Module: "files"}
		for _, cf := range detected.ConfigFiles {
			diff.LocalOnly = append(diff.LocalOnly, detect.DiffItem{Name: cf.Name, Type: "config", Value: cf.SourcePath})
		}
		diffs = append(diffs, diff)
	}

	return diffs
}

// Styles for rendering
var (
	moduleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#34d399"))

	syncedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#71717a"))

	localOnlyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fbbf24"))

	pactOnlyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f87171"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#52525b"))
)

// renderDiffs displays the diff results
func renderDiffs(diffs []detect.DiffResult, hasExisting bool) {
	fmt.Println("Detected Configuration:")
	fmt.Println(strings.Repeat("─", 60))

	for _, diff := range diffs {
		fmt.Println()
		fmt.Println(moduleStyle.Render("  " + diff.Module))

		// Show synced items
		for _, item := range diff.Synced {
			value := formatValue(item.Value)
			fmt.Printf("    %s %s %s\n",
				syncedStyle.Render("●"),
				item.Name,
				dimStyle.Render(value+" ✓"))
		}

		// Show local-only items
		for _, item := range diff.LocalOnly {
			value := formatValue(item.Value)
			label := "NEW"
			if hasExisting {
				label = "LOCAL ONLY"
			}
			fmt.Printf("    %s %s %s\n",
				localOnlyStyle.Render("○"),
				item.Name,
				localOnlyStyle.Render("← "+label+" "+value))
		}

		// Show pact-only items
		for _, item := range diff.PactOnly {
			value := formatValue(item.Value)
			fmt.Printf("    %s %s %s\n",
				pactOnlyStyle.Render("✗"),
				item.Name,
				pactOnlyStyle.Render("← PACT ONLY (not installed) "+value))
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()
	fmt.Printf("Legend: %s synced  %s can import  %s missing locally\n",
		syncedStyle.Render("●"),
		localOnlyStyle.Render("○"),
		pactOnlyStyle.Render("✗"))
}

func formatValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		if val != "" {
			return "(" + val + ")"
		}
	case bool:
		if val {
			return "(enabled)"
		}
	}
	return ""
}

// importAll imports all detected items without prompting
func importAll(detected *detect.DetectedConfig, diffs []detect.DiffResult) {
	// Build selection from all local-only items
	selected := make(map[string][]detect.DiffItem)
	for _, diff := range diffs {
		if len(diff.LocalOnly) > 0 {
			selected[diff.Module] = diff.LocalOnly
		}
	}

	applySelection(selected, detected)
}

// applySelection applies the user's selection
func applySelection(selected map[string][]detect.DiffItem, detected *detect.DetectedConfig) {
	if len(selected) == 0 {
		fmt.Println("Nothing selected to import.")
		return
	}

	pactDir, err := config.GetPactDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Build import selection
	selection := detect.BuildSelectionFromDiffs(selected, detected)

	// Check if pact.json exists
	if !config.Exists() {
		// Get username from git config or token
		username := detected.Git.User
		if username == "" {
			if keyring.HasToken() {
				token, _ := keyring.GetToken()
				if user, err := auth.GetUser(token); err == nil {
					username = user.Login
				}
			}
		}
		if username == "" {
			username = "user"
		}

		// Create new pact.json
		fmt.Println("\nCreating pact.json...")
		if err := detect.CreateDefaultPactJSON(detected, username, pactDir); err != nil {
			fmt.Printf("Error creating pact.json: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Merge into existing
		fmt.Println("\nUpdating pact.json...")
		if err := detect.Merge(selection, pactDir); err != nil {
			fmt.Printf("Error updating pact.json: %v\n", err)
			os.Exit(1)
		}
	}

	// Store secrets in keychain
	secretsStored := 0
	for _, s := range selection.Secrets {
		// Get the value from environment
		if val := os.Getenv(s); val != "" {
			if err := keyring.SetSecret(s, val); err == nil {
				secretsStored++
			}
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("✓ Updated pact.json")
	if len(selection.ConfigFiles) > 0 {
		fmt.Printf("✓ Copied %d config file(s) to .pact/\n", len(selection.ConfigFiles))
	}
	if secretsStored > 0 {
		fmt.Printf("✓ Added %d secret(s) to keychain\n", secretsStored)
	}
	fmt.Println()
	fmt.Println("Run 'pact push' to sync changes to GitHub")
}

// ============================================================================
// TUI Model for hierarchical selection
// ============================================================================

type readModel struct {
	stage     int // 0 = module selection, 1 = item selection
	diffs     []detect.DiffResult
	detected  *detect.DetectedConfig
	cursor    int
	selected  map[string][]detect.DiffItem
	moduleIdx int // Current module being edited (for stage 1)
	cancelled bool
	quitting  bool
}

type readKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Toggle key.Binding
	Enter  key.Binding
	Back   key.Binding
	All    key.Binding
	Quit   key.Binding
}

var readKeys = readKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Back: key.NewBinding(
		key.WithKeys("b", "esc"),
		key.WithHelp("b/esc", "back"),
	),
	All: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "all"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func initialReadModel(detected *detect.DetectedConfig, diffs []detect.DiffResult) readModel {
	// Filter to only modules with local-only items
	var filteredDiffs []detect.DiffResult
	for _, d := range diffs {
		if len(d.LocalOnly) > 0 {
			filteredDiffs = append(filteredDiffs, d)
		}
	}

	// Pre-select all modules
	selected := make(map[string][]detect.DiffItem)
	for _, d := range filteredDiffs {
		selected[d.Module] = d.LocalOnly
	}

	return readModel{
		stage:    0,
		diffs:    filteredDiffs,
		detected: detected,
		cursor:   0,
		selected: selected,
	}
}

func (m readModel) Init() tea.Cmd {
	return nil
}

func (m readModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, readKeys.Quit):
			m.cancelled = true
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, readKeys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, readKeys.Down):
			maxIdx := m.getMaxIndex()
			if m.cursor < maxIdx {
				m.cursor++
			}

		case key.Matches(msg, readKeys.Toggle):
			m.toggleCurrent()

		case key.Matches(msg, readKeys.All):
			m.toggleAll()

		case key.Matches(msg, readKeys.Enter):
			if m.stage == 0 {
				// Move to item selection for first selected module
				for i, d := range m.diffs {
					if _, ok := m.selected[d.Module]; ok {
						m.moduleIdx = i
						m.stage = 1
						m.cursor = 0
						return m, nil
					}
				}
				// No modules selected - finish
				m.quitting = true
				return m, tea.Quit
			} else {
				// Find next module to edit
				for i := m.moduleIdx + 1; i < len(m.diffs); i++ {
					if _, ok := m.selected[m.diffs[i].Module]; ok {
						m.moduleIdx = i
						m.cursor = 0
						return m, nil
					}
				}
				// No more modules - finish
				m.quitting = true
				return m, tea.Quit
			}

		case key.Matches(msg, readKeys.Back):
			if m.stage == 1 {
				m.stage = 0
				m.cursor = 0
			} else {
				m.cancelled = true
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m readModel) getMaxIndex() int {
	if m.stage == 0 {
		return len(m.diffs) - 1
	}
	return len(m.diffs[m.moduleIdx].LocalOnly) - 1
}

func (m *readModel) toggleCurrent() {
	if m.stage == 0 {
		// Toggle entire module
		module := m.diffs[m.cursor].Module
		if _, ok := m.selected[module]; ok {
			delete(m.selected, module)
		} else {
			m.selected[module] = m.diffs[m.cursor].LocalOnly
		}
	} else {
		// Toggle individual item
		module := m.diffs[m.moduleIdx].Module
		item := m.diffs[m.moduleIdx].LocalOnly[m.cursor]

		items := m.selected[module]
		found := -1
		for i, it := range items {
			if it.Name == item.Name && it.Type == item.Type {
				found = i
				break
			}
		}

		if found >= 0 {
			// Remove item
			items = append(items[:found], items[found+1:]...)
			if len(items) == 0 {
				delete(m.selected, module)
			} else {
				m.selected[module] = items
			}
		} else {
			// Add item
			m.selected[module] = append(m.selected[module], item)
		}
	}
}

func (m *readModel) toggleAll() {
	if m.stage == 0 {
		// Check if all are selected
		allSelected := true
		for _, d := range m.diffs {
			if _, ok := m.selected[d.Module]; !ok {
				allSelected = false
				break
			}
		}

		if allSelected {
			// Deselect all
			m.selected = make(map[string][]detect.DiffItem)
		} else {
			// Select all
			for _, d := range m.diffs {
				m.selected[d.Module] = d.LocalOnly
			}
		}
	} else {
		// Toggle all items in current module
		module := m.diffs[m.moduleIdx].Module
		allItems := m.diffs[m.moduleIdx].LocalOnly

		if len(m.selected[module]) == len(allItems) {
			delete(m.selected, module)
		} else {
			m.selected[module] = allItems
		}
	}
}

func (m readModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	if m.stage == 0 {
		b.WriteString("\nSelect modules to import:\n\n")

		for i, d := range m.diffs {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}

			checkbox := "[ ]"
			if _, ok := m.selected[d.Module]; ok {
				checkbox = "[x]"
			}

			count := len(d.LocalOnly)
			b.WriteString(fmt.Sprintf("%s%s %s (%d new)\n", cursor, checkbox, d.Module, count))
		}

		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  ↑/↓: navigate  space: toggle  enter: continue  a: all  q: quit"))
	} else {
		module := m.diffs[m.moduleIdx].Module
		b.WriteString(fmt.Sprintf("\nImporting from: %s\n\n", moduleStyle.Render(module)))

		items := m.diffs[m.moduleIdx].LocalOnly
		selectedItems := m.selected[module]
		selectedSet := make(map[string]bool)
		for _, it := range selectedItems {
			selectedSet[it.Name+":"+it.Type] = true
		}

		for i, item := range items {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}

			checkbox := "[ ]"
			if selectedSet[item.Name+":"+item.Type] {
				checkbox = "[x]"
			}

			value := formatValue(item.Value)
			b.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor, checkbox, item.Name, dimStyle.Render(value)))
		}

		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  ↑/↓: navigate  space: toggle  enter: confirm  b: back  a: all"))
	}

	return b.String()
}
