package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cloudboy-jh/pact/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update pact to the latest version",
	Long: `Update the pact CLI to the latest available version.

This command will detect your platform and update method, then update pact.

Supported methods:
  - Homebrew (macOS/Linux with brew)
  - Scoop (Windows with scoop)
  - Direct download (fallback for all platforms)

Examples:
  pact update              # Update to the latest version
  pact update --version    # Check current version`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle --version flag
		if versionFlag {
			fmt.Println(ui.RenderLogoWithVersion())
			return
		}

		fmt.Println(ui.RenderLogo())
		fmt.Println("Checking for updates...")

		// Detect update method
		method := detectUpdateMethod()

		switch method {
		case "homebrew":
			updateViaHomebrew()
		case "scoop":
			updateViaScoop()
		default:
			updateViaDirectDownload()
		}
	},
}

func detectUpdateMethod() string {
	// Check if installed via Homebrew
	if runtime.GOOS != "windows" {
		if _, err := exec.LookPath("brew"); err == nil {
			// Check if pact is installed via brew
			cmd := exec.Command("brew", "list", "pact")
			if err := cmd.Run(); err == nil {
				return "homebrew"
			}
		}
	}

	// Check if installed via Scoop
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("scoop"); err == nil {
			// Check if pact is in scoop's apps directory
			home, _ := os.UserHomeDir()
			scoopPactPath := filepath.Join(home, "scoop", "apps", "pact", "current", "pact.exe")
			if _, err := os.Stat(scoopPactPath); err == nil {
				return "scoop"
			}
		}
	}

	return "direct"
}

func updateViaHomebrew() {
	fmt.Println("Updating via Homebrew...")

	// Update homebrew tap
	cmd := exec.Command("brew", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to update Homebrew: %v\n", err)
	}

	// Upgrade pact
	cmd = exec.Command("brew", "upgrade", "pact")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: Failed to update pact via Homebrew: %v\n", err)
		fmt.Println("\nYou can try updating manually:")
		fmt.Println("  brew update && brew upgrade pact")
		os.Exit(1)
	}

	fmt.Println("\n✓ Pact updated successfully!")

	// Show new version
	cmd = exec.Command("pact", "--version")
	output, _ := cmd.Output()
	fmt.Printf("New version: %s", strings.TrimSpace(string(output)))
}

func updateViaScoop() {
	fmt.Println("Updating via Scoop...")

	// Update scoop bucket
	cmd := exec.Command("scoop", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to update Scoop: %v\n", err)
	}

	// Update pact
	cmd = exec.Command("scoop", "update", "pact")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: Failed to update pact via Scoop: %v\n", err)
		fmt.Println("\nYou can try updating manually:")
		fmt.Println("  scoop update && scoop update pact")
		os.Exit(1)
	}

	fmt.Println("\n✓ Pact updated successfully!")

	// Show new version
	cmd = exec.Command("pact", "--version")
	output, _ := cmd.Output()
	fmt.Printf("New version: %s", strings.TrimSpace(string(output)))
}

func updateViaDirectDownload() {
	fmt.Println("Updating via direct download...")

	// Get latest version from GitHub
	latestVersion, err := getLatestVersion()
	if err != nil {
		fmt.Printf("Error: Failed to check for updates: %v\n", err)
		os.Exit(1)
	}

	currentVersion := strings.TrimPrefix(ui.Version, "v")
	latestVersion = strings.TrimPrefix(latestVersion, "v")

	if currentVersion == latestVersion {
		fmt.Printf("\n✓ You already have the latest version: %s\n", currentVersion)
		return
	}

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Latest version: %s\n", latestVersion)
	fmt.Println()

	// Determine download URL
	goos := runtime.GOOS
	arch := runtime.GOARCH

	// Map arch names
	if arch == "amd64" {
		arch = "amd64"
	}

	var filename string
	if goos == "windows" {
		filename = fmt.Sprintf("pact_%s_windows_%s.zip", latestVersion, arch)
	} else {
		filename = fmt.Sprintf("pact_%s_%s_%s.tar.gz", latestVersion, goos, arch)
	}

	url := fmt.Sprintf("https://github.com/cloudboy-jh/pact/releases/download/v%s/%s", latestVersion, filename)

	fmt.Printf("Downloading %s...\n", filename)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pact-update")
	if err != nil {
		fmt.Printf("Error: Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// Download file
	tmpFile := filepath.Join(tmpDir, filename)
	if err := downloadFile(url, tmpFile); err != nil {
		fmt.Printf("Error: Failed to download update: %v\n", err)
		fmt.Println("\nYou can manually download from:")
		fmt.Printf("  %s\n", url)
		os.Exit(1)
	}

	// Extract
	fmt.Println("Extracting...")
	if err := extractArchive(tmpFile, tmpDir); err != nil {
		fmt.Printf("Error: Failed to extract update: %v\n", err)
		os.Exit(1)
	}

	// Find the binary
	binaryName := "pact"
	if runtime.GOOS == "windows" {
		binaryName = "pact.exe"
	}
	newBinary := filepath.Join(tmpDir, binaryName)

	// Find current binary location
	currentBinary, err := os.Executable()
	if err != nil {
		fmt.Printf("Error: Failed to find current binary: %v\n", err)
		os.Exit(1)
	}

	// Resolve symlinks
	currentBinary, err = filepath.EvalSymlinks(currentBinary)
	if err != nil {
		fmt.Printf("Error: Failed to resolve binary path: %v\n", err)
		os.Exit(1)
	}

	// Replace binary
	fmt.Println("Installing...")
	if err := replaceBinary(newBinary, currentBinary); err != nil {
		fmt.Printf("Error: Failed to install update: %v\n", err)
		fmt.Println("\nYou may need to run with elevated permissions.")
		fmt.Println("Or manually download from:")
		fmt.Printf("  %s\n", url)
		os.Exit(1)
	}

	fmt.Println("\n✓ Pact updated successfully!")
	fmt.Printf("New version: %s\n", latestVersion)
	fmt.Println("\nPlease restart your terminal for changes to take effect.")
}

func getLatestVersion() (string, error) {
	// Use GitHub API to get latest release
	cmd := exec.Command("curl", "-fsSL", "https://api.github.com/repos/cloudboy-jh/pact/releases/latest")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse tag_name from JSON
	outputStr := string(output)
	const tagPrefix = `"tag_name": "`
	start := strings.Index(outputStr, tagPrefix)
	if start == -1 {
		return "", fmt.Errorf("could not find tag_name in response")
	}
	start += len(tagPrefix)
	end := strings.Index(outputStr[start:], `"`)
	if end == -1 {
		return "", fmt.Errorf("could not parse tag_name")
	}

	return outputStr[start : start+end], nil
}

func downloadFile(url, filepath string) error {
	cmd := exec.Command("curl", "-fsSL", "-o", filepath, url)
	return cmd.Run()
}

func extractArchive(archivePath, destDir string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		cmd := exec.Command("unzip", "-q", "-o", archivePath, "-d", destDir)
		return cmd.Run()
	}
	cmd := exec.Command("tar", "-xzf", archivePath, "-C", destDir)
	return cmd.Run()
}

func replaceBinary(src, dst string) error {
	// On Windows, we can't overwrite a running executable
	// So we rename the old one, move the new one, and schedule deletion of old
	if runtime.GOOS == "windows" {
		oldDst := dst + ".old"

		// Remove any existing .old file
		os.Remove(oldDst)

		// Rename current binary to .old
		if err := os.Rename(dst, oldDst); err != nil {
			return fmt.Errorf("failed to rename old binary: %w", err)
		}

		// Move new binary into place
		if err := os.Rename(src, dst); err != nil {
			// Try to restore old binary
			os.Rename(oldDst, dst)
			return fmt.Errorf("failed to move new binary: %w", err)
		}

		// Schedule deletion of old binary on next reboot (Windows specific)
		// For now, we'll just leave it and it can be cleaned up manually
		fmt.Println("Note: Old binary saved as pact.exe.old - you can delete it manually.")

		return nil
	}

	// On Unix-like systems, we can overwrite the binary directly
	return os.Rename(src, dst)
}

func init() {
	updateCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version information")
}
