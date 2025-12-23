package detect

import (
	"os/exec"
	"strings"
)

// DetectGit detects git configuration
func DetectGit() GitDetected {
	result := GitDetected{}

	// Get git config values
	result.User = getGitConfig("user.name")
	result.Email = getGitConfig("user.email")
	result.DefaultBranch = getGitConfig("init.defaultBranch")

	// Check for Git LFS
	result.LFS = isGitLFSInstalled()

	return result
}

// getGitConfig retrieves a git config value
func getGitConfig(key string) string {
	cmd := exec.Command("git", "config", "--global", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// isGitLFSInstalled checks if Git LFS is installed and configured
func isGitLFSInstalled() bool {
	// Check if git-lfs command exists
	if !isToolInstalled("git-lfs") {
		return false
	}

	// Check if LFS is installed in git config
	cmd := exec.Command("git", "lfs", "version")
	err := cmd.Run()
	return err == nil
}
