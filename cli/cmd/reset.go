package cmd

import (
	"fmt"
	"os"

	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/sync"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Remove all symlinks",
	Long:  `Remove all symlinks created by pact. Keeps .pact/ intact.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists() {
			fmt.Println("Pact is not initialized.")
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Removing symlinks...")
		results, err := sync.RemoveAllSymlinks(cfg)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		removed := 0
		skipped := 0
		for _, r := range results {
			if r.Success {
				fmt.Printf("  ✓ Removed %s\n", r.Message)
				removed++
			} else if r.Skipped {
				skipped++
			} else if r.Error != nil {
				fmt.Printf("  ✗ %s/%s: %v\n", r.Module, r.Name, r.Error)
			}
		}

		fmt.Printf("\n%d removed, %d skipped\n", removed, skipped)
		fmt.Println(".pact/ directory kept intact. Run 'pact nuke' to remove it.")
	},
}
