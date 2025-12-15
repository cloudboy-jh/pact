package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PactConfig represents a flexible pact.json - any structure is valid
type PactConfig struct {
	Raw map[string]any // The raw parsed JSON
}

// SyncItem represents a single item to sync (for files that have source/target)
type SyncItem struct {
	Module   string
	Name     string
	Source   string
	Target   string
	Strategy string
	IsDir    bool
}

// ModuleInfo represents information about a module for display
type ModuleInfo struct {
	Name      string
	FileCount int
	Items     []string
}

// GetPactDir returns the pact directory path
// It searches for .pact/ in current directory and walks up the tree (like git)
// Falls back to ~/.pact/ for backwards compatibility
func GetPactDir() (string, error) {
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

// Load reads and parses pact.json flexibly
func Load() (*PactConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pact.json: %w", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse pact.json: %w", err)
	}

	return &PactConfig{Raw: raw}, nil
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

// GetCurrentOS returns the current OS name
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

// Get returns a value from the config by dot-separated path
// e.g., Get("shell.prompt.tool") or Get("name")
func (c *PactConfig) Get(path string) any {
	parts := strings.Split(path, ".")
	var current any = c.Raw

	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			return nil
		}
	}
	return current
}

// GetString returns a string value from the config
func (c *PactConfig) GetString(path string) string {
	val := c.Get(path)
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetStringSlice returns a string slice from the config
func (c *PactConfig) GetStringSlice(path string) []string {
	val := c.Get(path)
	if arr, ok := val.([]any); ok {
		var result []string
		for _, v := range arr {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// GetMap returns a map from the config
func (c *PactConfig) GetMap(path string) map[string]any {
	val := c.Get(path)
	if m, ok := val.(map[string]any); ok {
		return m
	}
	return nil
}

// HasKey checks if a key exists in the config
func (c *PactConfig) HasKey(path string) bool {
	return c.Get(path) != nil
}

// GetTopLevelKeys returns all top-level keys in the config
func (c *PactConfig) GetTopLevelKeys() []string {
	var keys []string
	for k := range c.Raw {
		keys = append(keys, k)
	}
	return keys
}

// GetModules returns all top-level keys that look like modules (objects, not primitives)
func (c *PactConfig) GetModules() []string {
	var modules []string
	skip := map[string]bool{"name": true, "version": true, "secrets": true}

	for k, v := range c.Raw {
		if skip[k] {
			continue
		}
		if _, ok := v.(map[string]any); ok {
			modules = append(modules, k)
		}
	}
	return modules
}

// GetSecrets returns the secrets array if it exists
func (c *PactConfig) GetSecrets() []string {
	return c.GetStringSlice("secrets")
}

// GetSyncItems finds all items with source/target for syncing
// Looks for "files" keys anywhere in the config tree
func (c *PactConfig) GetSyncItems() ([]SyncItem, error) {
	pactDir, err := GetPactDir()
	if err != nil {
		return nil, err
	}

	var items []SyncItem
	c.findFilesRecursive(c.Raw, "", pactDir, &items)
	return items, nil
}

// findFilesRecursive walks the config tree looking for "files" objects
func (c *PactConfig) findFilesRecursive(node any, module string, pactDir string, items *[]SyncItem) {
	m, ok := node.(map[string]any)
	if !ok {
		return
	}

	// Check if this node has a "files" key
	if files, ok := m["files"].(map[string]any); ok {
		for name, fileEntry := range files {
			if entry, ok := fileEntry.(map[string]any); ok {
				item := c.parseFileEntry(module, name, entry, pactDir)
				if item != nil {
					*items = append(*items, *item)
				}
			}
		}
	}

	// Recurse into child objects
	for key, val := range m {
		if key == "files" {
			continue
		}
		if childMap, ok := val.(map[string]any); ok {
			nextModule := key
			if module != "" {
				nextModule = module // Keep the top-level module name
			}
			c.findFilesRecursive(childMap, nextModule, pactDir, items)
		}
	}
}

// parseFileEntry parses a file entry with source/target
func (c *PactConfig) parseFileEntry(module, name string, entry map[string]any, pactDir string) *SyncItem {
	source, ok := entry["source"].(string)
	if !ok {
		return nil
	}

	target, err := c.resolveTarget(entry["target"])
	if err != nil {
		return nil
	}

	strategy, _ := entry["strategy"].(string)

	sourcePath := filepath.Join(pactDir, source)
	info, statErr := os.Stat(sourcePath)
	isDir := statErr == nil && info.IsDir()

	return &SyncItem{
		Module:   module,
		Name:     name,
		Source:   sourcePath,
		Target:   target,
		Strategy: strategy,
		IsDir:    isDir,
	}
}

// resolveTarget resolves the target path for the current OS
func (c *PactConfig) resolveTarget(target any) (string, error) {
	switch t := target.(type) {
	case string:
		return ExpandPath(t)
	case map[string]any:
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

// GetAvailableModules returns modules that have files configured for syncing
func (c *PactConfig) GetAvailableModules() []ModuleInfo {
	items, _ := c.GetSyncItems()

	// Group by module
	moduleMap := make(map[string][]string)
	for _, item := range items {
		moduleMap[item.Module] = append(moduleMap[item.Module], item.Name)
	}

	var modules []ModuleInfo
	for name, fileNames := range moduleMap {
		modules = append(modules, ModuleInfo{
			Name:      name,
			FileCount: len(fileNames),
			Items:     fileNames,
		})
	}
	return modules
}

// CountModuleFiles counts files for a module
func (c *PactConfig) CountModuleFiles(module string) int {
	items, _ := c.GetSyncItems()
	count := 0
	for _, item := range items {
		if item.Module == module {
			if item.IsDir {
				count += countFilesInDir(item.Source)
			} else {
				if _, err := os.Stat(item.Source); err == nil {
					count++
				}
			}
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

// GetSyncItemsForModule returns sync items for a specific module only
func (c *PactConfig) GetSyncItemsForModule(moduleName string) ([]SyncItem, error) {
	allItems, err := c.GetSyncItems()
	if err != nil {
		return nil, err
	}

	var items []SyncItem
	for _, item := range allItems {
		if item.Module == moduleName {
			items = append(items, item)
		}
	}
	return items, nil
}
