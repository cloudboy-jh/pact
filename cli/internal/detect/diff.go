package detect

import (
	"github.com/cloudboy-jh/pact/internal/config"
)

// DiffResult shows differences for a module
type DiffResult struct {
	Module    string     `json:"module"`
	LocalOnly []DiffItem `json:"localOnly"` // Detected but not in pact.json
	PactOnly  []DiffItem `json:"pactOnly"`  // In pact.json but not detected
	Synced    []DiffItem `json:"synced"`    // Present in both
}

// DiffItem represents a single item in the diff
type DiffItem struct {
	Name  string `json:"name"`
	Type  string `json:"type"` // "tool", "config", "secret", "setting"
	Value any    `json:"value,omitempty"`
}

// Compare compares detected config against existing pact.json
func Compare(detected *DetectedConfig, cfg *config.PactConfig) []DiffResult {
	var results []DiffResult

	// Compare CLI tools
	if cliDiff := compareCLI(detected.CLI, cfg); len(cliDiff.LocalOnly) > 0 || len(cliDiff.PactOnly) > 0 || len(cliDiff.Synced) > 0 {
		results = append(results, cliDiff)
	}

	// Compare shell
	if shellDiff := compareShell(detected.Shell, cfg); len(shellDiff.LocalOnly) > 0 || len(shellDiff.PactOnly) > 0 || len(shellDiff.Synced) > 0 {
		results = append(results, shellDiff)
	}

	// Compare git
	if gitDiff := compareGit(detected.Git, cfg); len(gitDiff.LocalOnly) > 0 || len(gitDiff.PactOnly) > 0 || len(gitDiff.Synced) > 0 {
		results = append(results, gitDiff)
	}

	// Compare editor
	if editorDiff := compareEditor(detected.Editor, cfg); len(editorDiff.LocalOnly) > 0 || len(editorDiff.PactOnly) > 0 || len(editorDiff.Synced) > 0 {
		results = append(results, editorDiff)
	}

	// Compare LLM
	if llmDiff := compareLLM(detected.LLM, cfg); len(llmDiff.LocalOnly) > 0 || len(llmDiff.PactOnly) > 0 || len(llmDiff.Synced) > 0 {
		results = append(results, llmDiff)
	}

	// Compare secrets
	if secretsDiff := compareSecrets(detected.Secrets, cfg); len(secretsDiff.LocalOnly) > 0 || len(secretsDiff.PactOnly) > 0 || len(secretsDiff.Synced) > 0 {
		results = append(results, secretsDiff)
	}

	// Compare config files
	if configDiff := compareConfigFiles(detected.ConfigFiles, cfg); len(configDiff.LocalOnly) > 0 || len(configDiff.PactOnly) > 0 || len(configDiff.Synced) > 0 {
		results = append(results, configDiff)
	}

	return results
}

func compareCLI(detected CLIDetected, cfg *config.PactConfig) DiffResult {
	result := DiffResult{Module: "cli"}

	pactTools := cfg.GetStringSlice("cli.tools")
	pactCustom := cfg.GetStringSlice("cli.custom")

	pactToolsSet := toSet(pactTools)
	pactCustomSet := toSet(pactCustom)

	// Check detected tools
	for _, tool := range detected.Tools {
		if pactToolsSet[tool] {
			result.Synced = append(result.Synced, DiffItem{Name: tool, Type: "tool"})
		} else {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: tool, Type: "tool"})
		}
	}

	// Check detected custom tools
	for _, tool := range detected.Custom {
		if pactCustomSet[tool] {
			result.Synced = append(result.Synced, DiffItem{Name: tool, Type: "custom"})
		} else {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: tool, Type: "custom"})
		}
	}

	// Check pact tools not detected locally
	detectedToolsSet := toSet(detected.Tools)
	detectedCustomSet := toSet(detected.Custom)

	for _, tool := range pactTools {
		if !detectedToolsSet[tool] {
			result.PactOnly = append(result.PactOnly, DiffItem{Name: tool, Type: "tool"})
		}
	}

	for _, tool := range pactCustom {
		if !detectedCustomSet[tool] {
			result.PactOnly = append(result.PactOnly, DiffItem{Name: tool, Type: "custom"})
		}
	}

	return result
}

func compareShell(detected ShellDetected, cfg *config.PactConfig) DiffResult {
	result := DiffResult{Module: "shell"}

	pactPromptTool := cfg.GetString("shell.prompt.tool")
	pactShellTools := cfg.GetStringSlice("shell.tools")
	pactShellToolsSet := toSet(pactShellTools)

	// Compare prompt
	if detected.Prompt != nil {
		if detected.Prompt.Tool == pactPromptTool {
			result.Synced = append(result.Synced, DiffItem{Name: detected.Prompt.Tool, Type: "prompt", Value: detected.Prompt.Theme})
		} else {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: detected.Prompt.Tool, Type: "prompt", Value: detected.Prompt.Theme})
		}
	}
	if pactPromptTool != "" && (detected.Prompt == nil || detected.Prompt.Tool != pactPromptTool) {
		result.PactOnly = append(result.PactOnly, DiffItem{Name: pactPromptTool, Type: "prompt"})
	}

	// Compare shell tools
	detectedShellToolsSet := toSet(detected.Tools)
	for _, tool := range detected.Tools {
		if pactShellToolsSet[tool] {
			result.Synced = append(result.Synced, DiffItem{Name: tool, Type: "tool"})
		} else {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: tool, Type: "tool"})
		}
	}

	for _, tool := range pactShellTools {
		if !detectedShellToolsSet[tool] {
			result.PactOnly = append(result.PactOnly, DiffItem{Name: tool, Type: "tool"})
		}
	}

	return result
}

func compareGit(detected GitDetected, cfg *config.PactConfig) DiffResult {
	result := DiffResult{Module: "git"}

	pactUser := cfg.GetString("git.user")
	pactEmail := cfg.GetString("git.email")
	pactBranch := cfg.GetString("git.defaultBranch")
	pactLFS := cfg.Get("git.lfs") == true

	// User
	if detected.User != "" {
		if detected.User == pactUser {
			result.Synced = append(result.Synced, DiffItem{Name: "user", Type: "setting", Value: detected.User})
		} else if pactUser == "" {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: "user", Type: "setting", Value: detected.User})
		} else {
			// Different values - show as local (they can choose to overwrite)
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: "user", Type: "setting", Value: detected.User})
		}
	} else if pactUser != "" {
		result.PactOnly = append(result.PactOnly, DiffItem{Name: "user", Type: "setting", Value: pactUser})
	}

	// Email
	if detected.Email != "" {
		if detected.Email == pactEmail {
			result.Synced = append(result.Synced, DiffItem{Name: "email", Type: "setting", Value: detected.Email})
		} else if pactEmail == "" {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: "email", Type: "setting", Value: detected.Email})
		} else {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: "email", Type: "setting", Value: detected.Email})
		}
	} else if pactEmail != "" {
		result.PactOnly = append(result.PactOnly, DiffItem{Name: "email", Type: "setting", Value: pactEmail})
	}

	// Default branch
	if detected.DefaultBranch != "" {
		if detected.DefaultBranch == pactBranch {
			result.Synced = append(result.Synced, DiffItem{Name: "defaultBranch", Type: "setting", Value: detected.DefaultBranch})
		} else if pactBranch == "" {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: "defaultBranch", Type: "setting", Value: detected.DefaultBranch})
		} else {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: "defaultBranch", Type: "setting", Value: detected.DefaultBranch})
		}
	} else if pactBranch != "" {
		result.PactOnly = append(result.PactOnly, DiffItem{Name: "defaultBranch", Type: "setting", Value: pactBranch})
	}

	// LFS
	if detected.LFS {
		if pactLFS {
			result.Synced = append(result.Synced, DiffItem{Name: "lfs", Type: "setting", Value: true})
		} else {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: "lfs", Type: "setting", Value: true})
		}
	} else if pactLFS {
		result.PactOnly = append(result.PactOnly, DiffItem{Name: "lfs", Type: "setting", Value: true})
	}

	return result
}

func compareEditor(detected EditorDetected, cfg *config.PactConfig) DiffResult {
	result := DiffResult{Module: "editor"}

	pactDefault := cfg.GetString("editor.default")

	if detected.Default != "" {
		if detected.Default == pactDefault {
			result.Synced = append(result.Synced, DiffItem{Name: detected.Default, Type: "editor"})
		} else if pactDefault == "" {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: detected.Default, Type: "editor"})
		} else {
			// Different default editor
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: detected.Default, Type: "editor"})
		}
	} else if pactDefault != "" {
		result.PactOnly = append(result.PactOnly, DiffItem{Name: pactDefault, Type: "editor"})
	}

	// Other editors are just informational - add as local only if not default
	for _, editor := range detected.Others {
		if editor != pactDefault {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: editor, Type: "editor-other"})
		}
	}

	return result
}

func compareLLM(detected LLMDetected, cfg *config.PactConfig) DiffResult {
	result := DiffResult{Module: "llm"}

	pactProviders := cfg.GetStringSlice("llm.providers")
	pactProvidersSet := toSet(pactProviders)

	// Providers
	detectedProvidersSet := toSet(detected.Providers)
	for _, provider := range detected.Providers {
		if pactProvidersSet[provider] {
			result.Synced = append(result.Synced, DiffItem{Name: provider, Type: "provider"})
		} else {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: provider, Type: "provider"})
		}
	}

	for _, provider := range pactProviders {
		if !detectedProvidersSet[provider] {
			result.PactOnly = append(result.PactOnly, DiffItem{Name: provider, Type: "provider"})
		}
	}

	// Local LLM
	if detected.Local != nil {
		pactRuntime := cfg.GetString("llm.local.runtime")
		pactModels := cfg.GetStringSlice("llm.local.models")
		pactModelsSet := toSet(pactModels)

		if detected.Local.Runtime == pactRuntime {
			result.Synced = append(result.Synced, DiffItem{Name: detected.Local.Runtime, Type: "runtime"})
		} else if pactRuntime == "" {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: detected.Local.Runtime, Type: "runtime"})
		}

		for _, model := range detected.Local.Models {
			if pactModelsSet[model] {
				result.Synced = append(result.Synced, DiffItem{Name: model, Type: "model"})
			} else {
				result.LocalOnly = append(result.LocalOnly, DiffItem{Name: model, Type: "model"})
			}
		}
	}

	// Coding agents
	if detected.Coding != nil {
		pactAgents := cfg.GetStringSlice("llm.coding.agents")
		pactAgentsSet := toSet(pactAgents)

		for _, agent := range detected.Coding.Agents {
			if pactAgentsSet[agent] {
				result.Synced = append(result.Synced, DiffItem{Name: agent, Type: "agent"})
			} else {
				result.LocalOnly = append(result.LocalOnly, DiffItem{Name: agent, Type: "agent"})
			}
		}
	}

	return result
}

func compareSecrets(detected []SecretDetected, cfg *config.PactConfig) DiffResult {
	result := DiffResult{Module: "secrets"}

	pactSecrets := cfg.GetSecrets()
	pactSecretsSet := toSet(pactSecrets)

	detectedSecretsSet := make(map[string]bool)
	for _, s := range detected {
		detectedSecretsSet[s.Name] = true
		if pactSecretsSet[s.Name] {
			result.Synced = append(result.Synced, DiffItem{Name: s.Name, Type: "secret"})
		} else {
			result.LocalOnly = append(result.LocalOnly, DiffItem{Name: s.Name, Type: "secret"})
		}
	}

	for _, secret := range pactSecrets {
		if !detectedSecretsSet[secret] {
			result.PactOnly = append(result.PactOnly, DiffItem{Name: secret, Type: "secret"})
		}
	}

	return result
}

func compareConfigFiles(detected []ConfigFile, _ *config.PactConfig) DiffResult {
	result := DiffResult{Module: "files"}

	// For config files, we just show what's available locally
	// There's no direct mapping in pact.json to compare against
	for _, cf := range detected {
		if cf.Exists {
			result.LocalOnly = append(result.LocalOnly, DiffItem{
				Name:  cf.Name,
				Type:  "config",
				Value: cf.SourcePath,
			})
		}
	}

	return result
}

// toSet converts a string slice to a set (map)
func toSet(items []string) map[string]bool {
	set := make(map[string]bool)
	for _, item := range items {
		set[item] = true
	}
	return set
}

// CountNewItems counts items that are local-only across all diffs
func CountNewItems(diffs []DiffResult) int {
	count := 0
	for _, d := range diffs {
		count += len(d.LocalOnly)
	}
	return count
}

// CountMissingItems counts items that are pact-only (not installed locally)
func CountMissingItems(diffs []DiffResult) int {
	count := 0
	for _, d := range diffs {
		count += len(d.PactOnly)
	}
	return count
}
