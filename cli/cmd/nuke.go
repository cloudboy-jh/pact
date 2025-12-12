package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/keyring"
	"github.com/cloudboy-jh/pact/internal/sync"
	"github.com/spf13/cobra"
)

var nukeForce bool

var nukeCmd = &cobra.Command{
	Use:   "nuke",
	Short: "Remove pact completely",
	Long:  `Remove all symlinks, delete .pact/, and remove stored token.`,
	Run: func(cmd *cobra.Command, args []string) {
		pactDir := config.FindPactDir()
		if pactDir == "" {
			fmt.Println("Pact is not initialized. Nothing to remove.")
			return
		}

		// Confirm unless --force
		if !nukeForce {
			fmt.Println("This will:")
			fmt.Println("  - Remove all symlinks created by pact")
			fmt.Printf("  - Delete %s directory\n", pactDir)
			fmt.Println("  - Remove stored GitHub token from keychain")
			fmt.Println()
			fmt.Print("Are you sure? [y/N] ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return
			}
		}

		// Remove symlinks first
		cfg, err := config.Load()
		if err == nil {
			fmt.Println("Removing symlinks...")
			results, _ := sync.RemoveAllSymlinks(cfg)
			removed := 0
			for _, r := range results {
				if r.Success {
					removed++
				}
			}
			fmt.Printf("  ✓ Removed %d symlinks\n", removed)
		}

		// Delete .pact directory
		fmt.Printf("Deleting %s...\n", pactDir)
		if err := os.RemoveAll(pactDir); err != nil {
			fmt.Printf("  ✗ Error removing %s: %v\n", pactDir, err)
		} else {
			fmt.Printf("  ✓ Deleted %s\n", pactDir)
		}

		// Remove token from keychain
		fmt.Println("Removing token from keychain...")
		if err := keyring.DeleteToken(); err != nil {
			// Ignore error if token doesn't exist
			fmt.Println("  ○ No token found or already removed")
		} else {
			fmt.Println("  ✓ Removed token from keychain")
		}

		fmt.Println()
		fmt.Println("Pact has been completely removed.")
	},
}

func init() {
	nukeCmd.Flags().BoolVarP(&nukeForce, "force", "f", false, "Skip confirmation")
}
