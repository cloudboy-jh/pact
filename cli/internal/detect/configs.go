package detect

import (
	"os"
	"path/filepath"
	"runtime"
)

// configLocation defines where to look for a config file
type configLocation struct {
	name       string   // Display name
	module     string   // Which module it belongs to
	paths      []string // Paths to check (first found wins)
	destSubdir string   // Destination subdirectory in .pact/
	isDir      bool     // Whether this is a directory
}

// getConfigLocations returns OS-appropriate config locations
func getConfigLocations() []configLocation {
	home, _ := os.UserHomeDir()

	locations := []configLocation{
		// Shell configs
		{
			name:       "zshrc",
			module:     "shell",
			paths:      []string{filepath.Join(home, ".zshrc")},
			destSubdir: "shell",
		},
		{
			name:       "bashrc",
			module:     "shell",
			paths:      []string{filepath.Join(home, ".bashrc")},
			destSubdir: "shell",
		},
		{
			name:       "profile",
			module:     "shell",
			paths:      []string{filepath.Join(home, ".profile"), filepath.Join(home, ".zprofile")},
			destSubdir: "shell",
		},

		// Git configs
		{
			name:       "gitconfig",
			module:     "git",
			paths:      []string{filepath.Join(home, ".gitconfig")},
			destSubdir: "git",
		},
		{
			name:       "gitignore_global",
			module:     "git",
			paths:      []string{filepath.Join(home, ".gitignore_global"), filepath.Join(home, ".gitignore")},
			destSubdir: "git",
		},

		// Tool configs
		{
			name:       "lazygit",
			module:     "tools",
			paths:      []string{filepath.Join(home, ".config/lazygit/config.yml")},
			destSubdir: "tools",
		},
		{
			name:       "starship",
			module:     "tools",
			paths:      []string{filepath.Join(home, ".config/starship.toml")},
			destSubdir: "tools",
		},
	}

	// Editor configs - platform specific
	switch runtime.GOOS {
	case "darwin":
		locations = append(locations,
			configLocation{
				name:       "nvim",
				module:     "editor",
				paths:      []string{filepath.Join(home, ".config/nvim")},
				destSubdir: "editor",
				isDir:      true,
			},
			configLocation{
				name:       "vscode-settings",
				module:     "editor",
				paths:      []string{filepath.Join(home, "Library/Application Support/Code/User/settings.json")},
				destSubdir: "editor/vscode",
			},
			configLocation{
				name:       "vscode-keybindings",
				module:     "editor",
				paths:      []string{filepath.Join(home, "Library/Application Support/Code/User/keybindings.json")},
				destSubdir: "editor/vscode",
			},
			configLocation{
				name:       "cursor-settings",
				module:     "editor",
				paths:      []string{filepath.Join(home, "Library/Application Support/Cursor/User/settings.json")},
				destSubdir: "editor/cursor",
			},
			configLocation{
				name:       "zed-settings",
				module:     "editor",
				paths:      []string{filepath.Join(home, ".config/zed/settings.json")},
				destSubdir: "editor/zed",
			},
		)
	case "linux":
		locations = append(locations,
			configLocation{
				name:       "nvim",
				module:     "editor",
				paths:      []string{filepath.Join(home, ".config/nvim")},
				destSubdir: "editor",
				isDir:      true,
			},
			configLocation{
				name:       "vscode-settings",
				module:     "editor",
				paths:      []string{filepath.Join(home, ".config/Code/User/settings.json")},
				destSubdir: "editor/vscode",
			},
			configLocation{
				name:       "vscode-keybindings",
				module:     "editor",
				paths:      []string{filepath.Join(home, ".config/Code/User/keybindings.json")},
				destSubdir: "editor/vscode",
			},
		)
	case "windows":
		locations = append(locations,
			configLocation{
				name:   "powershell-profile",
				module: "shell",
				paths: []string{
					filepath.Join(home, "Documents/PowerShell/Microsoft.PowerShell_profile.ps1"),
					filepath.Join(home, "Documents/WindowsPowerShell/Microsoft.PowerShell_profile.ps1"),
				},
				destSubdir: "shell",
			},
			configLocation{
				name:       "nvim",
				module:     "editor",
				paths:      []string{filepath.Join(home, "AppData/Local/nvim")},
				destSubdir: "editor",
				isDir:      true,
			},
			configLocation{
				name:       "vscode-settings",
				module:     "editor",
				paths:      []string{filepath.Join(home, "AppData/Roaming/Code/User/settings.json")},
				destSubdir: "editor/vscode",
			},
			configLocation{
				name:       "vscode-keybindings",
				module:     "editor",
				paths:      []string{filepath.Join(home, "AppData/Roaming/Code/User/keybindings.json")},
				destSubdir: "editor/vscode",
			},
		)
	}

	return locations
}

// DiscoverConfigFiles finds config files on the system
func DiscoverConfigFiles() []ConfigFile {
	var found []ConfigFile

	for _, loc := range getConfigLocations() {
		for _, p := range loc.paths {
			info, err := os.Stat(p)
			if err != nil {
				continue
			}

			// Check if it's expected type (file vs dir)
			if loc.isDir && !info.IsDir() {
				continue
			}
			if !loc.isDir && info.IsDir() {
				continue
			}

			found = append(found, ConfigFile{
				Name:       loc.name,
				SourcePath: p,
				DestPath:   filepath.Join(loc.destSubdir, loc.name),
				Module:     loc.module,
				Exists:     true,
				IsDir:      loc.isDir,
			})
			break // Found one, stop checking alternatives
		}
	}

	return found
}

// CopyConfigFile copies a config file to the pact directory
func CopyConfigFile(cf ConfigFile, pactDir string) error {
	destPath := filepath.Join(pactDir, cf.DestPath)

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	if cf.IsDir {
		return copyDir(cf.SourcePath, destPath)
	}
	return copyFile(cf.SourcePath, destPath)
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Get source file permissions
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, content, info.Mode())
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Create destination directory
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
