package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/cloudboy-jh/pact/internal/auth"
	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/git"
	"github.com/cloudboy-jh/pact/internal/keyring"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var fromUser string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize pact in current directory",
	Long:  `Authenticate with GitHub and clone your pact repo to ./.pact/ in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if already initialized in this directory tree
		if config.FindPactDir() != "" {
			fmt.Printf("Pact is already initialized at %s\n", config.FindPactDir())
			fmt.Println("Run 'pact nuke' first if you want to start fresh.")
			return
		}

		// Check if we already have a token
		if keyring.HasToken() {
			fmt.Println("Found existing GitHub token. Verifying...")
			token, _ := keyring.GetToken()
			user, err := auth.GetUser(token)
			if err == nil {
				fmt.Printf("Authenticated as %s\n", user.Login)
				if err := setupRepo(token, user.Login); err != nil {
					fmt.Printf("Error: %v\n", err)
					os.Exit(1)
				}
				return
			}
			fmt.Println("Token expired or invalid. Re-authenticating...")
			keyring.DeleteToken()
		}

		// Start device flow
		fmt.Println("Authenticating with GitHub...")
		fmt.Println()

		deviceCode, err := auth.RequestDeviceCode()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Please visit: %s\n", deviceCode.VerificationURI)
		fmt.Printf("And enter code: %s\n", deviceCode.UserCode)
		fmt.Println()
		fmt.Println("Waiting for authorization...")

		// Try to open browser
		browser.OpenURL(deviceCode.VerificationURI)

		// Poll for token
		token, err := auth.PollForToken(deviceCode.DeviceCode, deviceCode.Interval)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Get user info
		user, err := auth.GetUser(token)
		if err != nil {
			fmt.Printf("Error getting user info: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✓ Authenticated as %s\n", user.Login)

		// Store token
		if err := keyring.SetToken(token); err != nil {
			fmt.Printf("Warning: Could not store token in keychain: %v\n", err)
			fmt.Println("You may need to re-authenticate on next run.")
		}

		// Setup repo
		if err := setupRepo(token, user.Login); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	initCmd.Flags().StringVar(&fromUser, "from", "", "Fork pact from another user")
}

func setupRepo(token, username string) error {
	targetUser := username
	if fromUser != "" {
		targetUser = fromUser
		// TODO: Implement fork functionality
		fmt.Printf("Forking from %s is not yet implemented\n", fromUser)
		return nil
	}

	// Check if repo exists
	fmt.Printf("Checking for %s/my-pact repo...\n", targetUser)
	exists, err := auth.RepoExists(token, targetUser)
	if err != nil {
		return fmt.Errorf("failed to check repo: %w", err)
	}

	if !exists {
		fmt.Println("Repo not found. Creating...")
		if err := auth.CreateRepo(token); err != nil {
			return fmt.Errorf("failed to create repo: %w", err)
		}
		fmt.Println("✓ Created my-pact repo")

		// Wait a moment for GitHub to initialize the repo
		time.Sleep(2 * time.Second)
	}

	// Get local pact directory (current working directory)
	pactDir, err := config.GetLocalPactDir()
	if err != nil {
		return fmt.Errorf("failed to get pact directory: %w", err)
	}

	// Clone repo to ./.pact/
	fmt.Println("Cloning to ./.pact/...")
	if err := git.Clone(token, targetUser, pactDir); err != nil {
		return fmt.Errorf("failed to clone: %w", err)
	}

	fmt.Println("✓ Cloned repo to ./.pact/")

	// Check if pact.json exists, if not create a default one
	if !config.Exists() {
		fmt.Println("Creating default pact.json...")
		if err := createDefaultConfig(username); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
		fmt.Println("✓ Created pact.json")
	}

	fmt.Println()
	fmt.Println("Pact initialized! Run 'pact' to see status or 'pact sync' to apply configs.")

	return nil
}

func createDefaultConfig(username string) error {
	pactDir, err := config.GetPactDir()
	if err != nil {
		return err
	}

	defaultConfig := fmt.Sprintf(`{
  "version": "1.0.0",
  "user": "%s",
  "modules": {
    "shell": {},
    "editor": {},
    "git": {},
    "ai": {
      "providers": {},
      "prompts": {},
      "agents": {}
    },
    "tools": {
      "configs": {}
    }
  },
  "secrets": []
}
`, username)

	configPath := pactDir + "/pact.json"
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}
