package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/keyring"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets",
	Long:  `Manage secrets stored in your OS keychain.`,
}

var secretSetCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Set a secret",
	Long:  `Store a secret in the OS keychain.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		fmt.Printf("Enter value for %s: ", name)

		// Read password without echo
		password, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // newline after password input

		if err != nil {
			// Fallback to regular input if term.ReadPassword fails
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			password = []byte(strings.TrimSpace(input))
		}

		value := strings.TrimSpace(string(password))
		if value == "" {
			fmt.Println("Error: Value cannot be empty")
			os.Exit(1)
		}

		if err := keyring.SetSecret(name, value); err != nil {
			fmt.Printf("Error storing secret: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Secret '%s' stored in keychain\n", name)
	},
}

var secretListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets status",
	Long:  `Show which secrets are set in the keychain.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		if len(cfg.Secrets) == 0 {
			fmt.Println("No secrets configured in pact.json")
			return
		}

		fmt.Println("Secrets:")
		for _, name := range cfg.Secrets {
			if keyring.HasSecret(name) {
				fmt.Printf("  ● %s (set)\n", name)
			} else {
				fmt.Printf("  ○ %s (not set)\n", name)
			}
		}
	},
}

var secretRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a secret",
	Long:  `Remove a secret from the OS keychain.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if !keyring.HasSecret(name) {
			fmt.Printf("Secret '%s' is not set\n", name)
			return
		}

		if err := keyring.DeleteSecret(name); err != nil {
			fmt.Printf("Error removing secret: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Secret '%s' removed from keychain\n", name)
	},
}

func init() {
	secretCmd.AddCommand(secretSetCmd)
	secretCmd.AddCommand(secretListCmd)
	secretCmd.AddCommand(secretRemoveCmd)
}
