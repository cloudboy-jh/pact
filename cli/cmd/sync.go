package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cloudboy-jh/pact/internal/apply"
	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/git"
	"github.com/cloudboy-jh/pact/internal/keyring"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [module]",
	Short: "Sync and apply configs",
	Long: `Pull latest changes from GitHub and apply module configs.

Without arguments, shows an interactive picker to select modules.
With a module name, syncs that specific module directly.

Examples:
  pact sync              # Interactive module picker
  pact sync shell        # Install shell tools, configure prompt
  pact sync cli          # Install CLI tools (bun, node, lazygit, etc.)
  pact sync git          # Configure git (user, email, default branch)
  pact sync editor       # Setup editor preferences
  pact sync all          # Apply everything`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists() {
			fmt.Println("Pact is not initialized. Run 'pact init' first.")
			os.Exit(1)
		}

		// Get pact directory
		pactDir, err := config.GetPactDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
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
		if err := git.Pull(token, pactDir); err != nil {
			fmt.Printf("Warning: Could not pull: %v\n", err)
		} else {
			fmt.Println("✓ Pulled latest changes")
		}
		fmt.Println()

		// Load config
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Get available modules from config
		modules := cfg.GetModules()
		if len(modules) == 0 {
			fmt.Println("No modules found in pact.json")
			return
		}

		var modulesToSync []string

		if len(args) > 0 {
			arg := strings.ToLower(args[0])
			if arg == "all" {
				modulesToSync = modules
			} else {
				modulesToSync = []string{args[0]}
			}
		} else {
			// Interactive mode - show picker
			modulesToSync = promptModuleSelection(cfg, modules)
			if len(modulesToSync) == 0 {
				fmt.Println("No modules selected. Cancelled.")
				return
			}
		}

		// Apply selected modules
		fmt.Println()
		var allResults []apply.Result

		for _, moduleName := range modulesToSync {
			fmt.Printf("Applying %s...\n", moduleName)
			results, err := apply.ApplyModule(cfg, moduleName)
			if err != nil {
				fmt.Printf("  Error applying %s: %v\n", moduleName, err)
				continue
			}
			allResults = append(allResults, results...)
		}

		// Render results
		fmt.Println()
		renderApplyResults(allResults)
	},
}

func promptModuleSelection(cfg *config.PactConfig, modules []string) []string {
	fmt.Printf("Found %d modules in pact.json:\n\n", len(modules))

	// Display modules with numbers and details
	for i, mod := range modules {
		details := getModulePreview(cfg, mod)
		fmt.Printf("  [%d] %-12s %s\n", i+1, mod, details)
	}

	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  Enter numbers separated by commas (e.g., 1,3,5)")
	fmt.Println("  'a' or 'all' to sync all modules")
	fmt.Println("  'q' or 'quit' to cancel")
	fmt.Println()
	fmt.Print("Select modules: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" || input == "q" || input == "quit" {
		return nil
	}

	if input == "a" || input == "all" {
		return modules
	}

	// Parse comma-separated numbers
	var selected []string
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		num, err := strconv.Atoi(part)
		if err != nil {
			fmt.Printf("Warning: '%s' is not a valid number, skipping\n", part)
			continue
		}
		if num < 1 || num > len(modules) {
			fmt.Printf("Warning: %d is out of range, skipping\n", num)
			continue
		}
		selected = append(selected, modules[num-1])
	}

	return selected
}

func getModulePreview(cfg *config.PactConfig, module string) string {
	var parts []string

	switch module {
	case "shell":
		if tool := cfg.GetString("shell.prompt.tool"); tool != "" {
			parts = append(parts, tool)
		}
		if tools := cfg.GetStringSlice("shell.tools"); len(tools) > 0 {
			parts = append(parts, strings.Join(tools, ", "))
		}
	case "cli":
		if tools := cfg.GetStringSlice("cli.tools"); len(tools) > 0 {
			if len(tools) > 4 {
				parts = append(parts, strings.Join(tools[:4], ", ")+"...")
			} else {
				parts = append(parts, strings.Join(tools, ", "))
			}
		}
	case "git":
		if user := cfg.GetString("git.user"); user != "" {
			parts = append(parts, user)
		}
	case "editor":
		if def := cfg.GetString("editor.default"); def != "" {
			parts = append(parts, def)
		}
	case "terminal":
		if font := cfg.GetString("terminal.font"); font != "" {
			parts = append(parts, font)
		}
	case "llm":
		if providers := cfg.GetStringSlice("llm.providers"); len(providers) > 0 {
			parts = append(parts, strings.Join(providers, ", "))
		}
	}

	if len(parts) > 0 {
		return "(" + strings.Join(parts, ", ") + ")"
	}
	return ""
}

func renderApplyResults(results []apply.Result) {
	if len(results) == 0 {
		fmt.Println("No actions taken.")
		return
	}

	successCount := 0
	skipCount := 0
	failCount := 0

	// Group by category
	installs := []apply.Result{}
	configs := []apply.Result{}
	files := []apply.Result{}
	fonts := []apply.Result{}
	extensions := []apply.Result{}
	apps := []apply.Result{}

	for _, r := range results {
		switch r.Category {
		case "install":
			installs = append(installs, r)
		case "configure":
			configs = append(configs, r)
		case "file":
			files = append(files, r)
		case "font":
			fonts = append(fonts, r)
		case "extension":
			extensions = append(extensions, r)
		case "app":
			apps = append(apps, r)
		}
	}

	// Render installs
	if len(installs) > 0 {
		fmt.Println("Installations:")
		for _, r := range installs {
			icon, status := getResultDisplay(r)
			fmt.Printf("  %s %-20s %s\n", icon, r.Name, status)
			if r.Success {
				if r.Skipped {
					skipCount++
				} else {
					successCount++
				}
			} else {
				failCount++
			}
		}
		fmt.Println()
	}

	// Render configs
	if len(configs) > 0 {
		fmt.Println("Configuration:")
		for _, r := range configs {
			icon, status := getResultDisplay(r)
			name := fmt.Sprintf("%s.%s", r.Module, r.Name)
			fmt.Printf("  %s %-20s %s\n", icon, name, status)
			if r.Success {
				if r.Skipped {
					skipCount++
				} else {
					successCount++
				}
			} else {
				failCount++
			}
		}
		fmt.Println()
	}

	// Render files
	if len(files) > 0 {
		fmt.Println("Files:")
		for _, r := range files {
			icon, status := getResultDisplay(r)
			name := fmt.Sprintf("%s/%s", r.Module, r.Name)
			fmt.Printf("  %s %-20s %s\n", icon, name, status)
			if r.Success {
				if r.Skipped {
					skipCount++
				} else {
					successCount++
				}
			} else {
				failCount++
			}
		}
		fmt.Println()
	}

	// Render fonts
	if len(fonts) > 0 {
		fmt.Println("Fonts:")
		for _, r := range fonts {
			icon, status := getResultDisplay(r)
			fmt.Printf("  %s %-20s %s\n", icon, r.Name, status)
			if r.Success {
				if r.Skipped {
					skipCount++
				} else {
					successCount++
				}
			} else {
				failCount++
			}
		}
		fmt.Println()
	}

	// Render extensions
	if len(extensions) > 0 {
		fmt.Println("Extensions:")
		for _, r := range extensions {
			icon, status := getResultDisplay(r)
			fmt.Printf("  %s %-20s %s\n", icon, r.Name, status)
			if r.Success {
				if r.Skipped {
					skipCount++
				} else {
					successCount++
				}
			} else {
				failCount++
			}
		}
		fmt.Println()
	}

	// Render apps
	if len(apps) > 0 {
		fmt.Println("Apps:")
		for _, r := range apps {
			icon, status := getResultDisplay(r)
			fmt.Printf("  %s %-20s %s\n", icon, r.Name, status)
			if r.Success {
				if r.Skipped {
					skipCount++
				} else {
					successCount++
				}
			} else {
				failCount++
			}
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("Done: %d applied, %d skipped, %d failed\n", successCount, skipCount, failCount)
}

func getResultDisplay(r apply.Result) (string, string) {
	if r.Error != nil {
		return "✗", r.Error.Error()
	}
	if r.Skipped {
		return "○", r.Message
	}
	if r.Success {
		return "✓", r.Message
	}
	return "?", "unknown"
}
