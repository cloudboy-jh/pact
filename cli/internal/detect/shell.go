package detect

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// DetectShell detects shell configuration
func DetectShell() ShellDetected {
	result := ShellDetected{
		Tools: []string{},
	}

	// Detect shell type
	result.Type = detectShellType()

	// Detect prompt tool
	result.Prompt = detectPromptTool()

	// Detect shell tools
	for _, tool := range knownShellTools {
		if isToolInstalled(tool) {
			result.Tools = append(result.Tools, tool)
		}
	}

	return result
}

// detectShellType determines the current shell
func detectShellType() string {
	shell := os.Getenv("SHELL")

	if runtime.GOOS == "windows" {
		// Check for PowerShell
		if os.Getenv("PSModulePath") != "" {
			return "powershell"
		}
		return "cmd"
	}

	if strings.Contains(shell, "zsh") {
		return "zsh"
	} else if strings.Contains(shell, "bash") {
		return "bash"
	} else if strings.Contains(shell, "fish") {
		return "fish"
	}

	return filepath.Base(shell)
}

// detectPromptTool checks for installed prompt tools and their config
func detectPromptTool() *PromptInfo {
	// Check oh-my-posh first
	if isToolInstalled("oh-my-posh") {
		info := &PromptInfo{Tool: "oh-my-posh"}
		info.Theme, info.Source = parseOhMyPoshConfig()
		return info
	}

	// Check starship
	if isToolInstalled("starship") {
		return &PromptInfo{Tool: "starship"}
	}

	return nil
}

// parseOhMyPoshConfig tries to find oh-my-posh theme from shell config
func parseOhMyPoshConfig() (theme, source string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ""
	}

	// Shell config files to check
	var shellConfigs []string
	switch runtime.GOOS {
	case "darwin", "linux":
		shellConfigs = []string{
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".config/fish/config.fish"),
		}
	case "windows":
		shellConfigs = []string{
			filepath.Join(home, "Documents/PowerShell/Microsoft.PowerShell_profile.ps1"),
			filepath.Join(home, "Documents/WindowsPowerShell/Microsoft.PowerShell_profile.ps1"),
		}
	}

	// Regex to find oh-my-posh init with config
	// Matches: oh-my-posh init zsh --config '/path/to/theme.omp.json'
	// Or: eval "$(oh-my-posh init zsh --config '/path/to/theme.omp.json')"
	configRegex := regexp.MustCompile(`oh-my-posh.*--config\s+['"]?([^'"]+)['"]?`)

	for _, configPath := range shellConfigs {
		content, err := os.ReadFile(configPath)
		if err != nil {
			continue
		}

		matches := configRegex.FindStringSubmatch(string(content))
		if len(matches) >= 2 {
			themePath := matches[1]
			// Extract theme name from path
			baseName := filepath.Base(themePath)
			// Remove .omp.json extension
			theme = strings.TrimSuffix(baseName, ".omp.json")

			// Check if it's a URL
			if strings.HasPrefix(themePath, "http") {
				source = themePath
			}

			return theme, source
		}
	}

	return "", ""
}
