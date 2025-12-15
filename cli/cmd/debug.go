package cmd

import (
	"fmt"

	"github.com/cloudboy-jh/pact/internal/apply"
	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug config loading",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		font := cfg.GetString("terminal.font")
		fmt.Printf("terminal.font = %q\n", font)

		// Test applying terminal
		results, _ := apply.ApplyModule(cfg, "terminal")
		fmt.Printf("Results: %d\n", len(results))
		for _, r := range results {
			fmt.Printf("  %s/%s: success=%v skipped=%v msg=%s err=%v\n",
				r.Category, r.Name, r.Success, r.Skipped, r.Message, r.Error)
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
