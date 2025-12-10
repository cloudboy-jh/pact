package sync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudboy-jh/pact/internal/config"
)

// Result represents the result of syncing a single item
type Result struct {
	Module  string
	Name    string
	Success bool
	Error   error
	Skipped bool
	Message string
}

// SyncAll syncs all items from pact.json
func SyncAll(cfg *config.PactConfig) ([]Result, error) {
	items, err := cfg.GetSyncItems()
	if err != nil {
		return nil, err
	}

	var results []Result
	for _, item := range items {
		result := syncItem(item)
		results = append(results, result)
	}

	return results, nil
}

// SyncModule syncs only items from a specific module
func SyncModule(cfg *config.PactConfig, module string) ([]Result, error) {
	items, err := cfg.GetSyncItems()
	if err != nil {
		return nil, err
	}

	var results []Result
	found := false
	for _, item := range items {
		if item.Module == module {
			found = true
			result := syncItem(item)
			results = append(results, result)
		}
	}

	if !found {
		return nil, fmt.Errorf("module '%s' not found or not configured for this OS", module)
	}

	return results, nil
}

func syncItem(item config.SyncItem) Result {
	result := Result{
		Module: item.Module,
		Name:   item.Name,
	}

	// Check if source exists
	sourceInfo, err := os.Stat(item.Source)
	if err != nil {
		result.Error = fmt.Errorf("source not found: %s", item.Source)
		return result
	}

	// Determine strategy (default to symlink)
	strategy := item.Strategy
	if strategy == "" {
		strategy = "symlink"
	}

	// Ensure target parent directory exists
	targetDir := filepath.Dir(item.Target)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		result.Error = fmt.Errorf("failed to create target directory: %w", err)
		return result
	}

	// Remove existing target if it exists
	if _, err := os.Lstat(item.Target); err == nil {
		if err := os.RemoveAll(item.Target); err != nil {
			result.Error = fmt.Errorf("failed to remove existing target: %w", err)
			return result
		}
	}

	switch strategy {
	case "symlink":
		if err := createSymlink(item.Source, item.Target, sourceInfo.IsDir()); err != nil {
			result.Error = err
			return result
		}
		result.Message = fmt.Sprintf("symlinked %s -> %s", item.Target, item.Source)
	case "copy":
		if sourceInfo.IsDir() {
			if err := copyDir(item.Source, item.Target); err != nil {
				result.Error = err
				return result
			}
		} else {
			if err := copyFile(item.Source, item.Target); err != nil {
				result.Error = err
				return result
			}
		}
		result.Message = fmt.Sprintf("copied %s -> %s", item.Source, item.Target)
	default:
		result.Error = fmt.Errorf("unknown strategy: %s", strategy)
		return result
	}

	result.Success = true
	return result
}

func createSymlink(source, target string, isDir bool) error {
	// Convert to absolute path for symlink
	absSource, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := os.Symlink(absSource, target); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

func copyFile(source, target string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy permissions
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	return os.Chmod(target, sourceInfo.Mode())
}

func copyDir(source, target string) error {
	// Create target directory
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(target, sourceInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		targetPath := filepath.Join(target, entry.Name())

		if entry.IsDir() {
			if err := copyDir(sourcePath, targetPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// RemoveAllSymlinks removes all symlinks created by pact
func RemoveAllSymlinks(cfg *config.PactConfig) ([]Result, error) {
	items, err := cfg.GetSyncItems()
	if err != nil {
		return nil, err
	}

	var results []Result
	for _, item := range items {
		result := Result{
			Module: item.Module,
			Name:   item.Name,
		}

		// Check if target exists and is a symlink
		info, err := os.Lstat(item.Target)
		if err != nil {
			result.Skipped = true
			result.Message = "target does not exist"
			results = append(results, result)
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(item.Target); err != nil {
				result.Error = fmt.Errorf("failed to remove symlink: %w", err)
			} else {
				result.Success = true
				result.Message = fmt.Sprintf("removed symlink %s", item.Target)
			}
		} else {
			result.Skipped = true
			result.Message = "target is not a symlink (was it copied?)"
		}

		results = append(results, result)
	}

	return results, nil
}
