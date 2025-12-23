package detect

import (
	"os/exec"
)

// Known CLI tools to scan for
// Note: Shell-specific tools (zoxide, fzf, oh-my-posh, starship) are handled in shell.go
var knownCLITools = []string{
	// Runtimes
	"node", "bun", "deno", "go", "cargo", "python3", "ruby",
	// Package managers
	"npm", "yarn", "pnpm", "pip", "gem",
	// Git tools
	"git", "gh", "lazygit", "tig",
	// Docker/K8s
	"docker", "kubectl", "helm",
	// Search/navigation
	"ripgrep", "rg", "fd", "bat", "eza", "exa",
	// Utilities
	"jq", "yq", "curl", "wget", "httpie",
	// Build tools
	"make", "cmake", "ninja",
	// Cloud
	"aws", "gcloud", "az",
}

// Known shell tools that need init in shell config
var knownShellTools = []string{
	"zoxide", "fzf", "direnv", "nvm", "rbenv", "pyenv",
}

// Known prompt tools (used in shell.go)
// Exported for potential future use
var KnownPromptTools = []string{
	"oh-my-posh", "starship",
}

// Custom tools from GitHub releases
var knownCustomTools = []string{
	"pact", "churn", "annotr",
}

// DetectCLITools scans for installed CLI tools
func DetectCLITools() CLIDetected {
	result := CLIDetected{
		Tools:  []string{},
		Custom: []string{},
	}

	for _, tool := range knownCLITools {
		if isToolInstalled(tool) {
			// Avoid duplicates (rg is ripgrep)
			if tool == "rg" {
				continue // We already check for ripgrep
			}
			result.Tools = append(result.Tools, tool)
		}
	}

	for _, tool := range knownCustomTools {
		if isToolInstalled(tool) {
			result.Custom = append(result.Custom, tool)
		}
	}

	return result
}

// isToolInstalled checks if a tool is available in PATH
func isToolInstalled(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}
