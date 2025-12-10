package cmd

import (
	"fmt"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

const webURL = "http://localhost:5173" // TODO: Update to pact.dev when deployed

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open pact editor in browser",
	Long:  `Opens the pact web editor in your default browser.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Opening %s...\n", webURL)
		if err := browser.OpenURL(webURL); err != nil {
			fmt.Printf("Error opening browser: %v\n", err)
			fmt.Printf("Please visit %s manually.\n", webURL)
		}
	},
}
