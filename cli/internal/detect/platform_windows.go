//go:build windows

package detect

import (
	"os/exec"
	"strings"
)

// GetWingetPackages returns installed winget packages
func GetWingetPackages() []string {
	cmd := exec.Command("winget", "list", "--disable-interactivity")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var packages []string
	lines := strings.Split(string(output), "\n")
	// Skip header lines
	for i, line := range lines {
		if i < 2 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			packages = append(packages, fields[0])
		}
	}
	return packages
}

// GetChocoPackages returns installed Chocolatey packages
func GetChocoPackages() []string {
	cmd := exec.Command("choco", "list", "--local-only", "-r")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var packages []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		// Format: package|version
		parts := strings.Split(line, "|")
		if len(parts) > 0 && parts[0] != "" {
			packages = append(packages, parts[0])
		}
	}
	return packages
}

// GetScoopPackages returns installed Scoop packages
func GetScoopPackages() []string {
	cmd := exec.Command("scoop", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var packages []string
	lines := strings.Split(string(output), "\n")
	// Skip header lines
	for i, line := range lines {
		if i < 2 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			packages = append(packages, fields[0])
		}
	}
	return packages
}

// GetInstalledApps returns installed Windows applications
// This is a stub for future implementation
func GetInstalledApps() []string {
	// Future: query registry or use Get-AppxPackage
	return nil
}

// GetDefaultTerminal returns the default terminal
// This is a stub for future implementation
func GetDefaultTerminal() string {
	// Could be Windows Terminal, PowerShell, cmd, etc.
	return ""
}

// GetTerminalFont returns the configured terminal font
// This is a stub for future implementation
func GetTerminalFont() string {
	// Would need to read Windows Terminal settings.json
	return ""
}
