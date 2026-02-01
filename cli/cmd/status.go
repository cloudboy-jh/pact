package cmd

import (
	"fmt"
	"os"
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
		fmt.Println(ui.RenderStatus(cfg, 0, 0))
		return
	}

	// Get terminal dimensions
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		height = 24 // Fallback default
	}

	// Validate terminal height - use sane minimum
	if height < 10 {
		height = 24
	}

	// Disable mouse reporting BEFORE entering raw mode
	fmt.Print("\033[?1000l\033[?1002l\033[?1006l\033[?1015l")

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

	// Read single keys - use larger buffer for mouse/escape sequences
	buf := make([]byte, 32)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		if n > 0 {
			key := buf[0]

			// Skip escape sequences that aren't arrow keys (e.g., mouse events)
			// Mouse events typically look like: ESC [ M ... or ESC [ < ...
			if key == 27 && n >= 3 {
				// Check if it's an arrow key (ESC [ A/B/C/D)
				if buf[1] == 91 {
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
					case 67, 68: // Right/Left arrow - ignore
						continue
					default:
						// Any other escape sequence (mouse, etc.) - ignore
						continue
					}
				}
				// Other escape sequences - ignore
				continue
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
				// Drain any pending input first
				drainInput()
				// Show edit menu inline (below status)
				showEditMenuInline()
				// Read choice while still in raw mode
				choiceBuf := make([]byte, 1)
				_, err := os.Stdin.Read(choiceBuf)
				if err != nil {
					// Error reading - just re-render
					renderStatus(cfg, scrollOffset, height)
					continue
				}

				switch choiceBuf[0] {
				case 'l', 'L':
					term.Restore(int(os.Stdin.Fd()), oldState)
					fmt.Print("\r\n")
					editCmd.Run(editCmd, []string{})
					return
				case 'w', 'W':
					term.Restore(int(os.Stdin.Fd()), oldState)
					fmt.Print("\r\n")
					editCmd.Run(editCmd, []string{"web"})
					return
				case 'q', 'Q', 3: // q, Q, or Ctrl+C
					// Cancel - re-render status
					renderStatus(cfg, scrollOffset, height)
				default:
					// Any other key - cancel and re-render status
					renderStatus(cfg, scrollOffset, height)
				}
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
			case 27: // Lone ESC key - ignore
				continue
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

// drainInput clears any pending input from stdin
func drainInput() {
	// Set stdin to non-blocking temporarily to drain any buffered input
	buf := make([]byte, 256)
	for {
		n, _ := os.Stdin.Read(buf)
		if n == 0 {
			break
		}
	}
}

func showEditMenuInline() {
	// Print menu below current content (raw mode, so use \r\n)
	fmt.Print("\r\n")
	fmt.Print("\r\nEdit config:\r\n")
	fmt.Print("\r\n")
	fmt.Print("  [l] Local editor\r\n")
	fmt.Print("  [w] Web editor (pact-dev.com)\r\n")
	fmt.Print("  [q] Cancel\r\n")
	fmt.Print("\r\n")
	fmt.Print("Choose: ")
}
