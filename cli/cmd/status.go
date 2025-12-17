package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	webEditorURL    = "https://pact-dev.com"
	editorConfigKey = "editor.default"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show pact status",
	Long:  `Display the current status of all modules and secrets.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists() {
			fmt.Println("Pact is not initialized. Run 'pact init' first.")
			os.Exit(1)
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		runInteractiveStatus(cfg)
	},
}

func runInteractiveStatus(cfg *config.PactConfig) {
	// Check if we're in a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// Non-interactive mode
		fmt.Println(ui.RenderStatus(cfg, 0, 0))
		return
	}

	// Get terminal dimensions
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24 // Fallback defaults
	}

	// Set terminal to raw mode for single key input
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to non-interactive mode
		fmt.Println(ui.RenderStatus(cfg, 0, 0))
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	scrollOffset := 0

	// Render status (convert \n to \r\n for raw mode)
	renderStatus(cfg, scrollOffset, height)

	// Read single keys
	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		if n > 0 {
			key := buf[0]

			// Check for arrow keys (escape sequences)
			if n >= 3 && buf[0] == 27 && buf[1] == 91 {
				switch buf[2] {
				case 65: // Up arrow
					if scrollOffset > 0 {
						scrollOffset--
						renderStatus(cfg, scrollOffset, height)
					}
					continue
				case 66: // Down arrow
					maxScroll := ui.GetMaxScroll(cfg, height)
					if scrollOffset < maxScroll {
						scrollOffset++
						renderStatus(cfg, scrollOffset, height)
					}
					continue
				}
			}

			switch key {
			case 'q', 'Q', 3: // q, Q, or Ctrl+C
				// Clear and exit
				fmt.Print("\033[H\033[2J")
				return
			case 's', 'S':
				// Restore terminal, run sync, then return
				term.Restore(int(os.Stdin.Fd()), oldState)
				fmt.Print("\033[H\033[2J")
				runSync()
				return
			case 'e', 'E':
				// Open editor
				term.Restore(int(os.Stdin.Fd()), oldState)
				fmt.Print("\033[H\033[2J")
				openEditor(cfg, width, height)
				return
			case 'r', 'R':
				// Refresh
				cfg, _ = config.Load()
				scrollOffset = 0
				renderStatus(cfg, scrollOffset, height)
			case 'j', 'J': // Vim-style down
				maxScroll := ui.GetMaxScroll(cfg, height)
				if scrollOffset < maxScroll {
					scrollOffset++
					renderStatus(cfg, scrollOffset, height)
				}
			case 'k', 'K': // Vim-style up
				if scrollOffset > 0 {
					scrollOffset--
					renderStatus(cfg, scrollOffset, height)
				}
			}
		}
	}
}

func renderStatus(cfg *config.PactConfig, scrollOffset int, termHeight int) {
	// Clear screen
	fmt.Print("\033[H\033[2J")
	// Move cursor to top-left
	fmt.Print("\033[1;1H")

	// Get status and convert newlines for raw mode
	status := ui.RenderStatus(cfg, scrollOffset, termHeight)
	lines := strings.Split(status, "\n")
	for i, line := range lines {
		fmt.Print(line)
		if i < len(lines)-1 {
			fmt.Print("\r\n")
		}
	}
}

func runSync() {
	syncCmd.Run(syncCmd, []string{})
}

func openEditor(cfg *config.PactConfig, termWidth int, termHeight int) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Check if editor preference is configured in pact.json
	editorPref := cfg.GetString(editorConfigKey)

	// If not configured, prompt the user
	if editorPref == "" {
		editorPref = promptEditorChoice()
		if editorPref == "" {
			return // User cancelled
		}
	}

	// Handle "web" choice
	if editorPref == "web" {
		openWebEditor()
		return
	}

	// Handle "local" or specific editor
	openLocalEditor(configPath, editorPref)
}

func promptEditorChoice() string {
	fmt.Println("No editor configured. How would you like to edit your config?")
	fmt.Println()
	fmt.Println("  [1] Web Editor (pact-dev.com)")
	fmt.Println("  [2] Local Editor")
	fmt.Println("  [q] Cancel")
	fmt.Println()
	fmt.Print("Choose [1/2/q]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}

	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "1", "web", "w":
		fmt.Println()
		fmt.Println("Tip: Set \"editor\": { \"default\": \"web\" } in pact.json to skip this prompt.")
		return "web"
	case "2", "local", "l":
		fmt.Println()
		fmt.Println("Tip: Set \"editor\": { \"default\": \"local\" } in pact.json to skip this prompt.")
		return "local"
	case "q", "quit", "":
		return ""
	default:
		fmt.Printf("Unknown option: %s\n", input)
		return ""
	}
}

func openWebEditor() {
	fmt.Println("Opening web editor...")
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", webEditorURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", webEditorURL)
	default:
		cmd = exec.Command("xdg-open", webEditorURL)
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error opening web editor: %v\n", err)
		fmt.Printf("Please visit %s manually.\n", webEditorURL)
	}
}

func openLocalEditor(configPath string, editorPref string) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}

	var cmd *exec.Cmd

	// If user specified a specific editor name (not just "local"), try to use it
	if editorPref != "local" && editorPref != "" {
		if _, err := exec.LookPath(editorPref); err == nil {
			cmd = exec.Command(editorPref, configPath)
		}
	}

	if cmd == nil && editor != "" {
		// Use user's preferred editor from environment
		cmd = exec.Command(editor, configPath)
	}

	if cmd == nil {
		// Platform-specific defaults
		switch runtime.GOOS {
		case "darwin":
			// Use open -W to wait for the app to close, -t to open in default text editor
			cmd = exec.Command("open", "-W", "-t", configPath)
		case "windows":
			cmd = exec.Command("notepad", configPath)
		default:
			// Try common editors
			if _, err := exec.LookPath("nano"); err == nil {
				cmd = exec.Command("nano", configPath)
			} else if _, err := exec.LookPath("vim"); err == nil {
				cmd = exec.Command("vim", configPath)
			} else if _, err := exec.LookPath("vi"); err == nil {
				cmd = exec.Command("vi", configPath)
			} else {
				cmd = exec.Command("xdg-open", configPath)
			}
		}
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error opening editor: %v\n", err)
	}
}

// waitForKey waits for user to press any key (used in non-raw mode)
func waitForKey() {
	reader := bufio.NewReader(os.Stdin)
	reader.ReadByte()
}
