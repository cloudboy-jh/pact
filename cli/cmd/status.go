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
		fmt.Println(ui.RenderStatus(cfg))
		return
	}

	// Set terminal to raw mode for single key input
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to non-interactive mode
		fmt.Println(ui.RenderStatus(cfg))
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Render status (convert \n to \r\n for raw mode)
	renderStatus(cfg)

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
				openEditor()
				return
			case 'r', 'R':
				// Refresh
				cfg, _ = config.Load()
				renderStatus(cfg)
			}
		}
	}
}

func renderStatus(cfg *config.PactConfig) {
	// Clear screen
	fmt.Print("\033[H\033[2J")
	// Move cursor to top-left
	fmt.Print("\033[1;1H")

	// Get status and convert newlines for raw mode
	status := ui.RenderStatus(cfg)
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

func openEditor() {
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
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

// waitForKey waits for user to press any key (used in non-raw mode)
func waitForKey() {
	reader := bufio.NewReader(os.Stdin)
	reader.ReadByte()
}
