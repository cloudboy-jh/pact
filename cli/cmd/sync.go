package cmd

import (
	"fmt"
	"os"

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
	Long:  `Pull latest changes from GitHub and apply all module configs (or a specific module).`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists() {
			fmt.Println("Pact is not initialized. Run 'pact init' first.")
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
		if err := git.Pull(token); err != nil {
			fmt.Printf("Warning: Could not pull: %v\n", err)
		} else {
			fmt.Println("✓ Pulled latest changes")
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Sync
		var results []sync.Result
		if len(args) > 0 {
			// Sync specific module
			module := args[0]
			fmt.Printf("Syncing %s module...\n", module)
			results, err = sync.SyncModule(cfg, module)
		} else {
			// Sync all
			fmt.Println("Syncing all modules...")
			results, err = sync.SyncAll(cfg)
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Convert to UI results and render
		uiResults := make([]ui.SyncResult, len(results))
		for i, r := range results {
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
