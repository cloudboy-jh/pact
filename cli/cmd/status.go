package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
	// Set terminal to raw mode for single key input
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to non-interactive mode
		fmt.Println(ui.RenderStatus(cfg))
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Clear screen and render status
	clearScreen()
	fmt.Print(ui.RenderStatus(cfg))

	// Read single keys
	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		if n > 0 {
			key := buf[0]

			switch key {
			case 'q', 'Q', 3: // q, Q, or Ctrl+C
				clearScreen()
				return
			case 's', 'S':
				// Restore terminal, run sync, then return
				term.Restore(int(os.Stdin.Fd()), oldState)
				clearScreen()
				runSync()
				return
			case 'e', 'E':
				// Open editor/web
				term.Restore(int(os.Stdin.Fd()), oldState)
				clearScreen()
				openEditor(cfg)
				return
			case 'r', 'R':
				// Refresh
				clearScreen()
				cfg, _ = config.Load()
				fmt.Print(ui.RenderStatus(cfg))
			}
		}
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func runSync() {
	// Execute sync command
	syncCmd.Run(syncCmd, []string{})
}

func openEditor(cfg *config.PactConfig) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Try to open with default editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Platform-specific defaults
		switch runtime.GOOS {
		case "darwin":
			editor = "open"
		case "windows":
			editor = "notepad"
		default:
			editor = "xdg-open"
		}
	}

	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error opening editor: %v\n", err)
	}
}
