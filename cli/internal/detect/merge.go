package detect

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cloudboy-jh/pact/internal/config"
)

// ImportSelection represents what the user wants to import
type ImportSelection struct {
	CLITools     []string     // Tools to add to cli.tools
	CLICustom    []string     // Tools to add to cli.custom
	ShellPrompt  *PromptInfo  // Prompt config to set
	ShellTools   []string     // Tools to add to shell.tools
	Git          *GitDetected // Git settings to import
	Editor       string       // Default editor to set
	LLMProviders []string     // Providers to add
	LLMRuntime   string       // Local runtime (ollama)
	LLMModels    []string     // Models to add
	LLMAgents    []string     // Coding agents to add
	Secrets      []string     // Secrets to add to secrets array
	ConfigFiles  []ConfigFile // Config files to copy
}

// Merge applies the import selection to pact.json
func Merge(selection ImportSelection, pactDir string) error {
	configPath := filepath.Join(pactDir, "pact.json")

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Merge CLI tools
	if len(selection.CLITools) > 0 || len(selection.CLICustom) > 0 {
		cli := getOrCreateMap(raw, "cli")

		if len(selection.CLITools) > 0 {
			existing := getStringSlice(cli, "tools")
			cli["tools"] = mergeStringSlices(existing, selection.CLITools)
		}

		if len(selection.CLICustom) > 0 {
			existing := getStringSlice(cli, "custom")
			cli["custom"] = mergeStringSlices(existing, selection.CLICustom)
		}
	}

	// Merge shell config
	if selection.ShellPrompt != nil || len(selection.ShellTools) > 0 {
		shell := getOrCreateMap(raw, "shell")

		if selection.ShellPrompt != nil {
			prompt := make(map[string]any)
			prompt["tool"] = selection.ShellPrompt.Tool
			if selection.ShellPrompt.Theme != "" {
				prompt["theme"] = selection.ShellPrompt.Theme
			}
			if selection.ShellPrompt.Source != "" {
				prompt["source"] = selection.ShellPrompt.Source
			}
			shell["prompt"] = prompt
		}

		if len(selection.ShellTools) > 0 {
			existing := getStringSlice(shell, "tools")
			shell["tools"] = mergeStringSlices(existing, selection.ShellTools)
		}
	}

	// Merge git config
	if selection.Git != nil {
		git := getOrCreateMap(raw, "git")

		if selection.Git.User != "" {
			git["user"] = selection.Git.User
		}
		if selection.Git.Email != "" {
			git["email"] = selection.Git.Email
		}
		if selection.Git.DefaultBranch != "" {
			git["defaultBranch"] = selection.Git.DefaultBranch
		}
		if selection.Git.LFS {
			git["lfs"] = true
		}
	}

	// Merge editor config
	if selection.Editor != "" {
		editor := getOrCreateMap(raw, "editor")
		editor["default"] = selection.Editor
	}

	// Merge LLM config
	if len(selection.LLMProviders) > 0 || selection.LLMRuntime != "" || len(selection.LLMModels) > 0 || len(selection.LLMAgents) > 0 {
		llm := getOrCreateMap(raw, "llm")

		if len(selection.LLMProviders) > 0 {
			existing := getStringSlice(llm, "providers")
			llm["providers"] = mergeStringSlices(existing, selection.LLMProviders)
		}

		if selection.LLMRuntime != "" || len(selection.LLMModels) > 0 {
			local := getOrCreateMap(llm, "local")
			if selection.LLMRuntime != "" {
				local["runtime"] = selection.LLMRuntime
			}
			if len(selection.LLMModels) > 0 {
				existing := getStringSlice(local, "models")
				local["models"] = mergeStringSlices(existing, selection.LLMModels)
			}
		}

		if len(selection.LLMAgents) > 0 {
			coding := getOrCreateMap(llm, "coding")
			existing := getStringSlice(coding, "agents")
			coding["agents"] = mergeStringSlices(existing, selection.LLMAgents)
		}
	}

	// Merge secrets
	if len(selection.Secrets) > 0 {
		existing := getStringSlice(raw, "secrets")
		raw["secrets"] = mergeStringSlices(existing, selection.Secrets)
	}

	// Copy config files
	for _, cf := range selection.ConfigFiles {
		if err := CopyConfigFile(cf, pactDir); err != nil {
			// Log but continue
			continue
		}
	}

	// Write updated config
	output, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, output, 0644)
}

// BuildSelectionFromDiffs creates an ImportSelection from user-selected diff items
func BuildSelectionFromDiffs(selected map[string][]DiffItem, detected *DetectedConfig) ImportSelection {
	selection := ImportSelection{}

	// CLI items
	if items, ok := selected["cli"]; ok {
		for _, item := range items {
			switch item.Type {
			case "tool":
				selection.CLITools = append(selection.CLITools, item.Name)
			case "custom":
				selection.CLICustom = append(selection.CLICustom, item.Name)
			}
		}
	}

	// Shell items
	if items, ok := selected["shell"]; ok {
		for _, item := range items {
			switch item.Type {
			case "prompt":
				if detected.Shell.Prompt != nil && detected.Shell.Prompt.Tool == item.Name {
					selection.ShellPrompt = detected.Shell.Prompt
				}
			case "tool":
				selection.ShellTools = append(selection.ShellTools, item.Name)
			}
		}
	}

	// Git items
	if items, ok := selected["git"]; ok {
		selection.Git = &GitDetected{}
		for _, item := range items {
			switch item.Name {
			case "user":
				if v, ok := item.Value.(string); ok {
					selection.Git.User = v
				}
			case "email":
				if v, ok := item.Value.(string); ok {
					selection.Git.Email = v
				}
			case "defaultBranch":
				if v, ok := item.Value.(string); ok {
					selection.Git.DefaultBranch = v
				}
			case "lfs":
				selection.Git.LFS = true
			}
		}
	}

	// Editor items
	if items, ok := selected["editor"]; ok {
		for _, item := range items {
			if item.Type == "editor" {
				selection.Editor = item.Name
				break
			}
		}
	}

	// LLM items
	if items, ok := selected["llm"]; ok {
		for _, item := range items {
			switch item.Type {
			case "provider":
				selection.LLMProviders = append(selection.LLMProviders, item.Name)
			case "runtime":
				selection.LLMRuntime = item.Name
			case "model":
				selection.LLMModels = append(selection.LLMModels, item.Name)
			case "agent":
				selection.LLMAgents = append(selection.LLMAgents, item.Name)
			}
		}
	}

	// Secrets
	if items, ok := selected["secrets"]; ok {
		for _, item := range items {
			selection.Secrets = append(selection.Secrets, item.Name)
		}
	}

	// Config files
	if items, ok := selected["files"]; ok {
		for _, item := range items {
			// Find the matching config file from detected
			for _, cf := range detected.ConfigFiles {
				if cf.Name == item.Name {
					selection.ConfigFiles = append(selection.ConfigFiles, cf)
					break
				}
			}
		}
	}

	return selection
}

// CreateDefaultPactJSON creates a new pact.json from detected config
func CreateDefaultPactJSON(detected *DetectedConfig, username string, pactDir string) error {
	pactJSON := map[string]any{
		"name":    username,
		"version": "1.0.0",
	}

	// Add CLI tools
	if len(detected.CLI.Tools) > 0 || len(detected.CLI.Custom) > 0 {
		cli := make(map[string]any)
		if len(detected.CLI.Tools) > 0 {
			cli["tools"] = detected.CLI.Tools
		}
		if len(detected.CLI.Custom) > 0 {
			cli["custom"] = detected.CLI.Custom
		}
		pactJSON["cli"] = cli
	}

	// Add shell config
	if detected.Shell.Prompt != nil || len(detected.Shell.Tools) > 0 {
		shell := make(map[string]any)
		if detected.Shell.Prompt != nil {
			prompt := map[string]any{"tool": detected.Shell.Prompt.Tool}
			if detected.Shell.Prompt.Theme != "" {
				prompt["theme"] = detected.Shell.Prompt.Theme
			}
			if detected.Shell.Prompt.Source != "" {
				prompt["source"] = detected.Shell.Prompt.Source
			}
			shell["prompt"] = prompt
		}
		if len(detected.Shell.Tools) > 0 {
			shell["tools"] = detected.Shell.Tools
		}
		pactJSON["shell"] = shell
	}

	// Add git config
	if detected.Git.User != "" || detected.Git.Email != "" {
		git := make(map[string]any)
		if detected.Git.User != "" {
			git["user"] = detected.Git.User
		}
		if detected.Git.Email != "" {
			git["email"] = detected.Git.Email
		}
		if detected.Git.DefaultBranch != "" {
			git["defaultBranch"] = detected.Git.DefaultBranch
		}
		if detected.Git.LFS {
			git["lfs"] = true
		}
		pactJSON["git"] = git
	}

	// Add editor config
	if detected.Editor.Default != "" {
		pactJSON["editor"] = map[string]any{
			"default": detected.Editor.Default,
		}
	}

	// Add LLM config
	if len(detected.LLM.Providers) > 0 || detected.LLM.Local != nil {
		llm := make(map[string]any)
		if len(detected.LLM.Providers) > 0 {
			llm["providers"] = detected.LLM.Providers
		}
		if detected.LLM.Local != nil {
			local := map[string]any{"runtime": detected.LLM.Local.Runtime}
			if len(detected.LLM.Local.Models) > 0 {
				local["models"] = detected.LLM.Local.Models
			}
			llm["local"] = local
		}
		if detected.LLM.Coding != nil && len(detected.LLM.Coding.Agents) > 0 {
			llm["coding"] = map[string]any{"agents": detected.LLM.Coding.Agents}
		}
		pactJSON["llm"] = llm
	}

	// Add secrets (just the names, not values)
	var secretNames []string
	for _, s := range detected.Secrets {
		secretNames = append(secretNames, s.Name)
	}
	if len(secretNames) > 0 {
		pactJSON["secrets"] = secretNames
	}

	// Write to file
	output, err := json.MarshalIndent(pactJSON, "", "  ")
	if err != nil {
		return err
	}

	configPath := filepath.Join(pactDir, "pact.json")
	return os.WriteFile(configPath, output, 0644)
}

// Helper functions

func getOrCreateMap(parent map[string]any, key string) map[string]any {
	if v, ok := parent[key].(map[string]any); ok {
		return v
	}
	m := make(map[string]any)
	parent[key] = m
	return m
}

func getStringSlice(m map[string]any, key string) []string {
	if v, ok := m[key].([]any); ok {
		var result []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

func mergeStringSlices(existing, new []string) []any {
	seen := make(map[string]bool)
	var result []any

	for _, s := range existing {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	for _, s := range new {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}

// ValidateSelection checks if the selection is valid for import
func ValidateSelection(selection ImportSelection, pactDir string) error {
	// Check if pact directory exists
	if _, err := os.Stat(pactDir); os.IsNotExist(err) {
		return err
	}

	// Check if config files exist before trying to copy
	for _, cf := range selection.ConfigFiles {
		if _, err := os.Stat(cf.SourcePath); os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

// GetExistingConfig loads the existing pact config if available
func GetExistingConfig() (*config.PactConfig, error) {
	if !config.Exists() {
		return nil, nil
	}
	return config.Load()
}
