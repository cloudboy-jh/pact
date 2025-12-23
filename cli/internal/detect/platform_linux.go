//go:build linux

package detect

import (
	"os/exec"
	"strings"
)

// GetAptPackages returns installed apt packages
func GetAptPackages() []string {
	cmd := exec.Command("dpkg-query", "-W", "-f=${Package}\n")
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

// GetDnfPackages returns installed dnf packages
func GetDnfPackages() []string {
	cmd := exec.Command("dnf", "list", "installed", "-q")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var packages []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			// Package name is first field, may include arch
			name := strings.Split(fields[0], ".")[0]
			packages = append(packages, name)
		}
	}
	return packages
}

// GetPacmanPackages returns installed pacman packages
func GetPacmanPackages() []string {
	cmd := exec.Command("pacman", "-Qq")
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

// GetInstalledApps returns desktop applications
// This is a stub for future implementation
func GetInstalledApps() []string {
	// Future: scan /usr/share/applications and ~/.local/share/applications
	return nil
}

// GetDefaultTerminal returns the default terminal emulator
// This is a stub for future implementation
func GetDefaultTerminal() string {
	return ""
}

// GetTerminalFont returns the configured terminal font
// This is a stub for future implementation
func GetTerminalFont() string {
	return ""
}
