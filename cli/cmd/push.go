package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/git"
	"github.com/cloudboy-jh/pact/internal/keyring"
	"github.com/spf13/cobra"
)

var (
	pushMessage string
	pushForce   bool
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local changes to GitHub",
	Long:  `Commit and push all local changes in ~/.pact/ to GitHub.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists() {
			fmt.Println("Pact is not initialized. Run 'pact init' first.")
			os.Exit(1)
		}

		// Get token
		token, err := keyring.GetToken()
		if err != nil {
			fmt.Println("Not authenticated. Run 'pact init' to authenticate.")
			os.Exit(1)
		}

		// Check for changes
		hasChanges, err := git.HasChanges()
		if err != nil {
			fmt.Printf("Error checking for changes: %v\n", err)
			os.Exit(1)
		}

		if !hasChanges {
			fmt.Println("No changes to push.")
			return
		}

		// Get commit message
		message := pushMessage
		if message == "" {
			fmt.Print("Commit message: ")
			reader := bufio.NewReader(os.Stdin)
			message, _ = reader.ReadString('\n')
			message = strings.TrimSpace(message)
		}

		if message == "" {
			message = "Update pact configuration"
		}

		// Push
		fmt.Println("Pushing changes...")
		if err := git.Push(token, message); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Changes pushed to GitHub")
	},
}

func init() {
	pushCmd.Flags().StringVarP(&pushMessage, "message", "m", "", "Commit message")
	pushCmd.Flags().BoolVar(&pushForce, "force", false, "Force push (overwrite remote)")
}
