package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PactConfig represents the root pact.json structure
type PactConfig struct {
	Version string        `json:"version"`
	User    string        `json:"user"`
	Modules ModulesConfig `json:"modules"`
	Secrets []string      `json:"secrets"`
}

// ModulesConfig contains all module configurations
type ModulesConfig struct {
	Shell       map[string]ModuleEntry `json:"shell,omitempty"`
	Editor      map[string]EditorEntry `json:"editor,omitempty"`
	Terminal    *TerminalEntry         `json:"terminal,omitempty"`
	Git         map[string]ModuleEntry `json:"git,omitempty"`
	AI          *AIConfig              `json:"ai,omitempty"`
	Tools       *ToolsConfig           `json:"tools,omitempty"`
	Keybindings map[string]ModuleEntry `json:"keybindings,omitempty"`
	Snippets    map[string]ModuleEntry `json:"snippets,omitempty"`
	Fonts       *FontsConfig           `json:"fonts,omitempty"`
	Runtimes    *RuntimesConfig        `json:"runtimes,omitempty"`
}

// ModuleEntry represents a simple source/target mapping
type ModuleEntry struct {
	Source   string      `json:"source"`
	Target   interface{} `json:"target"` // Can be string or map[string]string
	Strategy string      `json:"strategy,omitempty"`
}

// EditorEntry represents editor config with OS-specific targets
type EditorEntry struct {
	Source   string      `json:"source"`
	Target   interface{} `json:"target"` // Can be string or map[string]string
	Strategy string      `json:"strategy,omitempty"`
}

// TerminalEntry represents terminal emulator config
type TerminalEntry struct {
	Emulator string      `json:"emulator"`
	Source   string      `json:"source"`
	Target   interface{} `json:"target"`
	Strategy string      `json:"strategy,omitempty"`
}

// AIConfig represents AI module configuration
type AIConfig struct {
	Providers map[string]ProviderConfig `json:"providers,omitempty"`
	Prompts   map[string]string         `json:"prompts,omitempty"`
	Skills    string                    `json:"skills,omitempty"`
	Agents    map[string]AgentEntry     `json:"agents,omitempty"`
}

// ProviderConfig represents an AI provider configuration
type ProviderConfig struct {
	DefaultModel string   `json:"defaultModel,omitempty"`
	Models       []string `json:"models,omitempty"`
}

// AgentEntry represents an agent config file mapping
type AgentEntry struct {
	Source   string `json:"source"`
	Target   string `json:"target,omitempty"`
	Strategy string `json:"strategy,omitempty"`
}

// ToolsConfig represents tools module configuration
type ToolsConfig struct {
	Configs  map[string]ModuleEntry `json:"configs,omitempty"`
	Packages *PackagesConfig        `json:"packages,omitempty"`
}

// PackagesConfig represents package lists by manager
type PackagesConfig struct {
	Brew  []string `json:"brew,omitempty"`
	NPM   []string `json:"npm,omitempty"`
	Cargo []string `json:"cargo,omitempty"`
	Go    []string `json:"go,omitempty"`
}

// FontsConfig represents fonts to install
type FontsConfig struct {
	Install []string `json:"install,omitempty"`
}

// RuntimesConfig represents runtime versions
type RuntimesConfig struct {
	Node    string `json:"node,omitempty"`
	Python  string `json:"python,omitempty"`
	Go      string `json:"go,omitempty"`
	Manager string `json:"manager,omitempty"`
}

// SyncItem represents a single item to sync
type SyncItem struct {
	Module   string
	Name     string
	Source   string
	Target   string
	Strategy string
	IsDir    bool
}

// GetPactDir returns the pact directory path
// It searches for .pact/ in current directory and walks up the tree (like git)
// Falls back to ~/.pact/ for backwards compatibility
func GetPactDir() (string, error) {
	// First, look for .pact in current directory and walk up
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	dir := cwd
	for {
		pactDir := filepath.Join(dir, ".pact")
		if info, err := os.Stat(pactDir); err == nil && info.IsDir() {
			return pactDir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, no .pact found
			break
		}
		dir = parent
	}

	// Fallback to ~/.pact for backwards compatibility
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".pact"), nil
}

// GetLocalPactDir returns .pact/ in the current working directory
// Used by init to create the local .pact folder
func GetLocalPactDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, ".pact"), nil
}

// FindPactDir searches for .pact/ starting from current directory
// Returns empty string if not found (does not fall back to ~/.pact)
func FindPactDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := cwd
	for {
		pactDir := filepath.Join(dir, ".pact")
		if info, err := os.Stat(pactDir); err == nil && info.IsDir() {
			return pactDir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

// GetConfigPath returns the path to pact.json
func GetConfigPath() (string, error) {
	pactDir, err := GetPactDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(pactDir, "pact.json"), nil
}

// Load reads and parses pact.json from ~/.pact/
func Load() (*PactConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pact.json: %w", err)
	}

	var config PactConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse pact.json: %w", err)
	}

	return &config, nil
}

// Exists checks if pact.json exists
func Exists() bool {
	configPath, err := GetConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(configPath)
	return err == nil
}

// GetCurrentOS returns the current OS name matching pact.json keys
func GetCurrentOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return runtime.GOOS
	}
}

// ExpandPath expands ~ to home directory
func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

// ResolveTarget resolves the target path for the current OS
func ResolveTarget(target interface{}) (string, error) {
	switch t := target.(type) {
	case string:
		return ExpandPath(t)
	case map[string]interface{}:
		currentOS := GetCurrentOS()
		if path, ok := t[currentOS]; ok {
			if pathStr, ok := path.(string); ok {
				return ExpandPath(pathStr)
			}
		}
		return "", fmt.Errorf("no target configured for %s", currentOS)
	default:
		return "", fmt.Errorf("invalid target type: %T", target)
	}
}

// GetSyncItems returns all items that need to be synced for the current OS
func (c *PactConfig) GetSyncItems() ([]SyncItem, error) {
	pactDir, err := GetPactDir()
	if err != nil {
		return nil, err
	}

	var items []SyncItem
	currentOS := GetCurrentOS()

	// Shell module
	if c.Modules.Shell != nil {
		if entry, ok := c.Modules.Shell[currentOS]; ok {
			target, err := ResolveTarget(entry.Target)
			if err == nil {
				items = append(items, SyncItem{
					Module:   "shell",
					Name:     currentOS,
					Source:   filepath.Join(pactDir, entry.Source),
					Target:   target,
					Strategy: entry.Strategy,
				})
			}
		}
	}

	// Editor module
	if c.Modules.Editor != nil {
		for name, entry := range c.Modules.Editor {
			target, err := ResolveTarget(entry.Target)
			if err != nil {
				continue // Skip if no target for current OS
			}
			source := filepath.Join(pactDir, entry.Source)
			info, statErr := os.Stat(source)
			isDir := statErr == nil && info.IsDir()

			items = append(items, SyncItem{
				Module:   "editor",
				Name:     name,
				Source:   source,
				Target:   target,
				Strategy: entry.Strategy,
				IsDir:    isDir,
			})
		}
	}

	// Terminal module
	if c.Modules.Terminal != nil {
		target, err := ResolveTarget(c.Modules.Terminal.Target)
		if err == nil {
			items = append(items, SyncItem{
				Module:   "terminal",
				Name:     c.Modules.Terminal.Emulator,
				Source:   filepath.Join(pactDir, c.Modules.Terminal.Source),
				Target:   target,
				Strategy: c.Modules.Terminal.Strategy,
			})
		}
	}

	// Git module
	if c.Modules.Git != nil {
		for name, entry := range c.Modules.Git {
			target, err := ResolveTarget(entry.Target)
			if err != nil {
				continue
			}
			items = append(items, SyncItem{
				Module:   "git",
				Name:     name,
				Source:   filepath.Join(pactDir, entry.Source),
				Target:   target,
				Strategy: entry.Strategy,
			})
		}
	}

	// AI agents
	if c.Modules.AI != nil && c.Modules.AI.Agents != nil {
		for name, agent := range c.Modules.AI.Agents {
			if agent.Target == "" {
				continue // No target means it's project-local only
			}
			target, err := ExpandPath(agent.Target)
			if err != nil {
				continue
			}
			items = append(items, SyncItem{
				Module:   "ai",
				Name:     name,
				Source:   filepath.Join(pactDir, agent.Source),
				Target:   target,
				Strategy: agent.Strategy,
			})
		}
	}

	// Tools configs
	if c.Modules.Tools != nil && c.Modules.Tools.Configs != nil {
		for name, entry := range c.Modules.Tools.Configs {
			target, err := ResolveTarget(entry.Target)
			if err != nil {
				continue
			}
			items = append(items, SyncItem{
				Module:   "tools",
				Name:     name,
				Source:   filepath.Join(pactDir, entry.Source),
				Target:   target,
				Strategy: entry.Strategy,
			})
		}
	}

	// Keybindings
	if c.Modules.Keybindings != nil {
		for name, entry := range c.Modules.Keybindings {
			target, err := ResolveTarget(entry.Target)
			if err != nil {
				continue
			}
			items = append(items, SyncItem{
				Module:   "keybindings",
				Name:     name,
				Source:   filepath.Join(pactDir, entry.Source),
				Target:   target,
				Strategy: entry.Strategy,
			})
		}
	}

	// Snippets
	if c.Modules.Snippets != nil {
		for name, entry := range c.Modules.Snippets {
			target, err := ResolveTarget(entry.Target)
			if err != nil {
				continue
			}
			source := filepath.Join(pactDir, entry.Source)
			info, statErr := os.Stat(source)
			isDir := statErr == nil && info.IsDir()

			items = append(items, SyncItem{
				Module:   "snippets",
				Name:     name,
				Source:   source,
				Target:   target,
				Strategy: entry.Strategy,
				IsDir:    isDir,
			})
		}
	}

	return items, nil
}

// CountModuleFiles counts files for a module
func (c *PactConfig) CountModuleFiles(module string) int {
	pactDir, err := GetPactDir()
	if err != nil {
		return 0
	}

	count := 0
	items, _ := c.GetSyncItems()
	for _, item := range items {
		if item.Module == module {
			source := item.Source
			if item.IsDir {
				count += countFilesInDir(source)
			} else {
				if _, err := os.Stat(source); err == nil {
					count++
				}
			}
		}
	}

	// Special handling for AI prompts and skills
	if module == "ai" && c.Modules.AI != nil {
		if c.Modules.AI.Prompts != nil {
			count += len(c.Modules.AI.Prompts)
		}
		if c.Modules.AI.Skills != "" {
			skillsDir := filepath.Join(pactDir, c.Modules.AI.Skills)
			count += countFilesInDir(skillsDir)
		}
	}

	return count
}

func countFilesInDir(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}
