package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/git"
	"github.com/cloudboy-jh/pact/internal/keyring"
	"github.com/cloudboy-jh/pact/internal/sync"
	"github.com/cloudboy-jh/pact/internal/ui"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [module]",
	Short: "Sync configs from GitHub",
	Long: `Pull latest changes from GitHub and apply module configs.

Without arguments, shows an interactive picker to select modules.
With a module name, syncs that specific module directly.

Examples:
  pact sync              # Interactive module picker
  pact sync shell        # Sync only shell module
  pact sync editor       # Sync only editor module
  pact sync theme        # Sync only theme module`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists() {
			fmt.Println("Pact is not initialized. Run 'pact init' first.")
			os.Exit(1)
		}

		// Get pact directory
		pactDir, err := config.GetPactDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Get token for pull
		token, err := keyring.GetToken()
		if err != nil {
			fmt.Println("Not authenticated. Run 'pact init' to authenticate.")
			os.Exit(1)
		}

		// Pull latest changes
		fmt.Println("Pulling latest changes...")
		if err := git.Pull(token, pactDir); err != nil {
			fmt.Printf("Warning: Could not pull: %v\n", err)
		} else {
			fmt.Println("✓ Pulled latest changes")
		}
		fmt.Println()

		// Load config
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Get available modules
		modules := cfg.GetAvailableModules()
		if len(modules) == 0 {
			fmt.Println("No modules configured in pact.json")
			return
		}

		var modulesToSync []string

		if len(args) > 0 {
			// Specific module requested - sync directly without prompt
			modulesToSync = []string{args[0]}
		} else {
			// Interactive mode - show picker
			modulesToSync = promptModuleSelection(modules)
			if len(modulesToSync) == 0 {
				fmt.Println("No modules selected. Cancelled.")
				return
			}
		}

		// Sync selected modules
		fmt.Println()
		var allResults []sync.Result
		for _, moduleName := range modulesToSync {
			fmt.Printf("Syncing %s...\n", moduleName)
			results, err := sync.SyncModule(cfg, moduleName)
			if err != nil {
				fmt.Printf("  Error syncing %s: %v\n", moduleName, err)
				continue
			}
			allResults = append(allResults, results...)
		}

		// Convert to UI results and render
		uiResults := make([]ui.SyncResult, len(allResults))
		for i, r := range allResults {
			uiResults[i] = ui.SyncResult{
				Module:  r.Module,
				Name:    r.Name,
				Success: r.Success,
				Skipped: r.Skipped,
				Error:   r.Error,
			}
		}

		fmt.Println()
		fmt.Println(ui.RenderSyncResults(uiResults))
	},
}

func promptModuleSelection(modules []config.ModuleInfo) []string {
	fmt.Printf("Found %d modules in pact.json:\n\n", len(modules))

	// Display modules with numbers
	for i, mod := range modules {
		itemsStr := ""
		if len(mod.Items) > 0 {
			itemsStr = fmt.Sprintf("(%s)", strings.Join(mod.Items, ", "))
		}
		fileStr := "file"
		if mod.FileCount != 1 {
			fileStr = "files"
		}
		fmt.Printf("  [%d] %-12s %d %s %s\n", i+1, mod.Name, mod.FileCount, fileStr, itemsStr)
	}

	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  Enter numbers separated by commas (e.g., 1,3,5)")
	fmt.Println("  'a' or 'all' to sync all modules")
	fmt.Println("  'q' or 'quit' to cancel")
	fmt.Println()
	fmt.Print("Select modules: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" || input == "q" || input == "quit" {
		return nil
	}

	if input == "a" || input == "all" {
		var all []string
		for _, mod := range modules {
			all = append(all, mod.Name)
		}
		return all
	}

	// Parse comma-separated numbers
	var selected []string
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		num, err := strconv.Atoi(part)
		if err != nil {
			fmt.Printf("Warning: '%s' is not a valid number, skipping\n", part)
			continue
		}
		if num < 1 || num > len(modules) {
			fmt.Printf("Warning: %d is out of range, skipping\n", num)
			continue
		}
		selected = append(selected, modules[num-1].Name)
	}

	return selected
}
