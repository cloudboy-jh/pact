package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

const defaultWebURL = "https://pact-dev.com"

func getWebURL() string {
	if url := os.Getenv("PACT_WEB_URL"); url != "" {
		return url
	}
	return defaultWebURL
}

func getEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	// Fallback based on OS
	if _, err := exec.LookPath("vim"); err == nil {
		return "vim"
	}
	if _, err := exec.LookPath("nano"); err == nil {
		return "nano"
	}
	return "vi"
}

func openInEditor(path string) error {
	editor := getEditor()
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var editCmd = &cobra.Command{
	Use:   "edit [path]",
	Short: "Edit pact files locally or open web editor",
	Long: `Edit pact configuration files.

Without arguments, opens .pact/pact.json in your $EDITOR.
With a path, opens that file/directory relative to .pact/.

Use 'pact edit web' to open the web editor in your browser.

Examples:
  pact edit              # Edit pact.json in $EDITOR
  pact edit shell        # Edit shell directory
  pact edit shell/zshrc  # Edit specific file
  pact edit web          # Open web editor in browser`,
	Run: func(cmd *cobra.Command, args []string) {
		// No args = open pact.json in editor
		if len(args) == 0 {
			pactDir, err := config.GetPactDir()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			configPath := filepath.Join(pactDir, "pact.json")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				fmt.Println("Pact not initialized. Run 'pact init' first.")
				os.Exit(1)
			}

			fmt.Printf("Opening %s in %s...\n", configPath, getEditor())
			if err := openInEditor(configPath); err != nil {
				fmt.Printf("Error opening editor: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// "web" subcommand = open browser
		if args[0] == "web" {
			webURL := getWebURL()
			fmt.Printf("Opening %s...\n", webURL)
			if err := browser.OpenURL(webURL); err != nil {
				fmt.Printf("Error opening browser: %v\n", err)
				fmt.Printf("Please visit %s manually.\n", webURL)
			}
			return
		}

		// Otherwise, open the specified path in editor
		pactDir, err := config.GetPactDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		targetPath := filepath.Join(pactDir, args[0])
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			fmt.Printf("Path not found: %s\n", targetPath)
			fmt.Println("Available modules: shell, editor, terminal, git, ai, tools, keybindings, snippets, fonts")
			os.Exit(1)
		}

		fmt.Printf("Opening %s in %s...\n", targetPath, getEditor())
		if err := openInEditor(targetPath); err != nil {
			fmt.Printf("Error opening editor: %v\n", err)
			os.Exit(1)
		}
	},
}
