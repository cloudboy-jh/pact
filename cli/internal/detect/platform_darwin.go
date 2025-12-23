//go:build darwin

package detect

import (
	"os/exec"
	"strings"
)

// GetBrewPackages returns installed Homebrew packages
func GetBrewPackages() []string {
	cmd := exec.Command("brew", "list", "--formula", "-1")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var packages []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			packages = append(packages, line)
		}
	}
	return packages
}

// GetBrewCasks returns installed Homebrew casks
func GetBrewCasks() []string {
	cmd := exec.Command("brew", "list", "--cask", "-1")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var casks []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			casks = append(casks, line)
		}
	}
	return casks
}

// GetInstalledApps returns apps from /Applications
// This is a stub for future implementation
func GetInstalledApps() []string {
	// Future: scan /Applications and ~/Applications
	return nil
}

// GetDefaultTerminal returns the default terminal app
// This is a stub for future implementation
func GetDefaultTerminal() string {
	// Could check defaults read com.apple.LaunchServices/com.apple.launchservices.secure
	// or look for iTerm2, Alacritty, Kitty, Ghostty, etc.
	return ""
}

// GetTerminalFont returns the font configured in the default terminal
// This is a stub for future implementation
func GetTerminalFont() string {
	// Would need to read Terminal.app or iTerm2 preferences
	return ""
}
