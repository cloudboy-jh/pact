package apply

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cloudboy-jh/pact/internal/config"
)

// Result represents the result of applying a config item
type Result struct {
	Category string // "install", "configure", "file", "font", "extension", "app"
	Module   string
	Name     string
	Success  bool
	Skipped  bool
	Message  string
	Error    error
}

// Apply applies the entire pact configuration
func Apply(cfg *config.PactConfig) ([]Result, error) {
	var results []Result

	// 1. Install CLI tools
	toolResults := applyCliTools(cfg)
	results = append(results, toolResults...)

	// 2. Setup shell (prompt, tools, config injection)
	shellResults := applyShell(cfg)
	results = append(results, shellResults...)

	// 3. Setup git config
	gitResults := applyGit(cfg)
	results = append(results, gitResults...)

	// 4. Setup editor + extensions
	editorResults := applyEditor(cfg)
	results = append(results, editorResults...)

	// 5. Setup terminal + fonts
	terminalResults := applyTerminal(cfg)
	results = append(results, terminalResults...)

	// 6. Install apps
	appResults := applyApps(cfg)
	results = append(results, appResults...)

	// 7. Apply any file syncs
	fileResults := applyFiles(cfg)
	results = append(results, fileResults...)

	return results, nil
}

// ApplyModule applies a specific module
func ApplyModule(cfg *config.PactConfig, module string) ([]Result, error) {
	switch module {
	case "cli":
		return applyCliTools(cfg), nil
	case "shell":
		return applyShell(cfg), nil
	case "git":
		return applyGit(cfg), nil
	case "editor":
		return applyEditor(cfg), nil
	case "terminal":
		return applyTerminal(cfg), nil
	case "llm":
		return applyLLM(cfg), nil
	case "apps":
		return applyApps(cfg), nil
	default:
		// Try to apply files for this module
		return applyModuleFiles(cfg, module), nil
	}
}

// =============================================================================
// CLI Tools
// =============================================================================

func applyCliTools(cfg *config.PactConfig) []Result {
	var results []Result

	// Standard tools from package manager
	tools := cfg.GetStringSlice("cli.tools")
	if len(tools) > 0 {
		pm := detectPackageManager()
		if pm == "" {
			results = append(results, Result{
				Category: "install",
				Module:   "cli",
				Name:     "package-manager",
				Error:    fmt.Errorf("no supported package manager found (brew, apt, winget)"),
			})
		} else {
			for _, tool := range tools {
				result := installTool(pm, tool)
				results = append(results, result)
			}
		}
	}

	// Custom tools from GitHub releases
	customTools := cfg.GetStringSlice("cli.custom")
	for _, tool := range customTools {
		result := installCustomTool(cfg, tool)
		results = append(results, result)
	}

	return results
}

// installCustomTool installs a tool from GitHub releases
func installCustomTool(cfg *config.PactConfig, tool string) Result {
	result := Result{
		Category: "install",
		Module:   "cli",
		Name:     tool,
	}

	// Check if already installed
	if isToolInstalled(tool) {
		result.Success = true
		result.Skipped = true
		result.Message = "already installed"
		return result
	}

	// Map tool names to GitHub repos
	repoMap := map[string]string{
		"pact":   "cloudboy-jh/pact",
		"churn":  "cloudboy-jh/churn",
		"annotr": "cloudboy-jh/annotr",
	}

	repo, ok := repoMap[tool]
	if !ok {
		// Try to install via package manager as fallback
		pm := detectPackageManager()
		if pm != "" {
			return installTool(pm, tool)
		}
		result.Error = fmt.Errorf("unknown custom tool and no package manager available")
		return result
	}

	// Get latest release from GitHub
	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(releaseURL)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch release info: %w", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		result.Error = fmt.Errorf("no releases found for %s", repo)
		return result
	}

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		result.Error = fmt.Errorf("failed to parse release info: %w", err)
		return result
	}

	// Find the right asset for this OS/arch
	osName := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, osName) && (strings.Contains(name, arch) || strings.Contains(name, "amd64") || strings.Contains(name, "x64")) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		result.Error = fmt.Errorf("no compatible release found for %s/%s", osName, arch)
		return result
	}

	// Download and install
	tmpFile := filepath.Join(os.TempDir(), tool+"-download")
	if err := downloadFile(downloadURL, tmpFile); err != nil {
		result.Error = err
		return result
	}
	defer os.Remove(tmpFile)

	// Determine install location
	installDir := "/usr/local/bin"
	if runtime.GOOS == "windows" {
		home, _ := os.UserHomeDir()
		installDir = filepath.Join(home, "bin")
		os.MkdirAll(installDir, 0755)
	}

	installPath := filepath.Join(installDir, tool)
	if runtime.GOOS == "windows" {
		installPath += ".exe"
	}

	// Handle tar.gz or zip
	if strings.HasSuffix(downloadURL, ".tar.gz") || strings.HasSuffix(downloadURL, ".tgz") {
		if err := extractTarGz(tmpFile, installDir, tool); err != nil {
			result.Error = err
			return result
		}
	} else if strings.HasSuffix(downloadURL, ".zip") {
		if err := extractZip(tmpFile, installDir, tool); err != nil {
			result.Error = err
			return result
		}
	} else {
		// Direct binary
		if err := copyFile(tmpFile, installPath); err != nil {
			result.Error = err
			return result
		}
		os.Chmod(installPath, 0755)
	}

	result.Success = true
	result.Message = fmt.Sprintf("installed from %s", repo)
	return result
}

// =============================================================================
// Shell
// =============================================================================

func applyShell(cfg *config.PactConfig) []Result {
	var results []Result

	// Install prompt tool
	promptTool := cfg.GetString("shell.prompt.tool")
	if promptTool != "" {
		pm := detectPackageManager()
		if pm != "" {
			result := installTool(pm, promptTool)
			results = append(results, result)
		}

		// Download theme
		themeSource := cfg.GetString("shell.prompt.source")
		themeName := cfg.GetString("shell.prompt.theme")
		if themeSource != "" && themeName != "" {
			result := downloadPromptTheme(promptTool, themeName, themeSource)
			results = append(results, result)
		}

		// Inject shell config
		result := injectShellConfig(cfg, promptTool, themeName)
		results = append(results, result)
	}

	// Install shell tools
	shellTools := cfg.GetStringSlice("shell.tools")
	if len(shellTools) > 0 {
		pm := detectPackageManager()
		if pm != "" {
			for _, tool := range shellTools {
				result := installTool(pm, tool)
				results = append(results, result)

				// Inject tool init into shell config
				initResult := injectToolInit(tool)
				if initResult.Message != "" {
					results = append(results, initResult)
				}
			}
		}
	}

	return results
}

// injectShellConfig adds prompt initialization to shell config
func injectShellConfig(cfg *config.PactConfig, promptTool, themeName string) Result {
	result := Result{
		Category: "configure",
		Module:   "shell",
		Name:     "shell-config",
	}

	home, _ := os.UserHomeDir()
	var shellConfig string
	var initLine string

	switch runtime.GOOS {
	case "darwin", "linux":
		// Detect shell
		shell := os.Getenv("SHELL")
		if strings.Contains(shell, "zsh") {
			shellConfig = filepath.Join(home, ".zshrc")
		} else if strings.Contains(shell, "bash") {
			shellConfig = filepath.Join(home, ".bashrc")
		} else {
			shellConfig = filepath.Join(home, ".zshrc") // default
		}

		switch promptTool {
		case "oh-my-posh":
			themePath := filepath.Join(home, ".config/oh-my-posh/themes", themeName+".omp.json")
			initLine = fmt.Sprintf(`eval "$(oh-my-posh init %s --config '%s')"`, filepath.Base(shell), themePath)
		case "starship":
			initLine = `eval "$(starship init zsh)"`
		}

	case "windows":
		shellConfig = filepath.Join(home, "Documents/PowerShell/Microsoft.PowerShell_profile.ps1")
		os.MkdirAll(filepath.Dir(shellConfig), 0755)

		switch promptTool {
		case "oh-my-posh":
			themePath := filepath.Join(home, "AppData/Local/Programs/oh-my-posh/themes", themeName+".omp.json")
			initLine = fmt.Sprintf(`oh-my-posh init pwsh --config '%s' | Invoke-Expression`, themePath)
		case "starship":
			initLine = `Invoke-Expression (&starship init powershell)`
		}
	}

	if initLine == "" {
		result.Skipped = true
		result.Success = true
		result.Message = "no init line for this prompt tool"
		return result
	}

	// Check if already in config
	existing, _ := os.ReadFile(shellConfig)
	if strings.Contains(string(existing), promptTool) {
		result.Success = true
		result.Skipped = true
		result.Message = "already configured"
		return result
	}

	// Append to shell config
	f, err := os.OpenFile(shellConfig, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		result.Error = err
		return result
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("\n# Pact: %s\n%s\n", promptTool, initLine))
	if err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.Message = fmt.Sprintf("added to %s", filepath.Base(shellConfig))
	return result
}

// injectToolInit adds tool initialization to shell config
func injectToolInit(tool string) Result {
	result := Result{
		Category: "configure",
		Module:   "shell",
		Name:     tool + "-init",
	}

	home, _ := os.UserHomeDir()
	var shellConfig string
	var initLine string

	shell := os.Getenv("SHELL")
	shellName := "zsh"
	if strings.Contains(shell, "bash") {
		shellName = "bash"
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		if shellName == "zsh" {
			shellConfig = filepath.Join(home, ".zshrc")
		} else {
			shellConfig = filepath.Join(home, ".bashrc")
		}

		switch tool {
		case "zoxide":
			initLine = fmt.Sprintf(`eval "$(zoxide init %s)"`, shellName)
		case "fzf":
			initLine = `[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh`
		case "direnv":
			initLine = fmt.Sprintf(`eval "$(direnv hook %s)"`, shellName)
		default:
			return result // No init needed
		}

	case "windows":
		shellConfig = filepath.Join(home, "Documents/PowerShell/Microsoft.PowerShell_profile.ps1")

		switch tool {
		case "zoxide":
			initLine = `Invoke-Expression (& { (zoxide init powershell | Out-String) })`
		default:
			return result
		}
	}

	if initLine == "" {
		return result
	}

	// Check if already in config
	existing, _ := os.ReadFile(shellConfig)
	if strings.Contains(string(existing), tool) {
		result.Success = true
		result.Skipped = true
		result.Message = "already configured"
		return result
	}

	f, err := os.OpenFile(shellConfig, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		result.Error = err
		return result
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("\n# Pact: %s\n%s\n", tool, initLine))
	if err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.Message = fmt.Sprintf("added to %s", filepath.Base(shellConfig))
	return result
}

// =============================================================================
// Git
// =============================================================================

func applyGit(cfg *config.PactConfig) []Result {
	var results []Result

	user := cfg.GetString("git.user")
	email := cfg.GetString("git.email")
	defaultBranch := cfg.GetString("git.defaultBranch")

	if user != "" {
		if err := runGitConfig("user.name", user); err != nil {
			results = append(results, Result{
				Category: "configure",
				Module:   "git",
				Name:     "user.name",
				Error:    err,
			})
		} else {
			results = append(results, Result{
				Category: "configure",
				Module:   "git",
				Name:     "user.name",
				Success:  true,
				Message:  user,
			})
		}
	}

	if email != "" {
		if err := runGitConfig("user.email", email); err != nil {
			results = append(results, Result{
				Category: "configure",
				Module:   "git",
				Name:     "user.email",
				Error:    err,
			})
		} else {
			results = append(results, Result{
				Category: "configure",
				Module:   "git",
				Name:     "user.email",
				Success:  true,
				Message:  email,
			})
		}
	}

	if defaultBranch != "" {
		if err := runGitConfig("init.defaultBranch", defaultBranch); err != nil {
			results = append(results, Result{
				Category: "configure",
				Module:   "git",
				Name:     "init.defaultBranch",
				Error:    err,
			})
		} else {
			results = append(results, Result{
				Category: "configure",
				Module:   "git",
				Name:     "init.defaultBranch",
				Success:  true,
				Message:  defaultBranch,
			})
		}
	}

	// Git LFS
	if cfg.Get("git.lfs") == true {
		if err := exec.Command("git", "lfs", "install").Run(); err != nil {
			pm := detectPackageManager()
			if pm != "" {
				installTool(pm, "git-lfs")
				exec.Command("git", "lfs", "install").Run()
			}
		}
		results = append(results, Result{
			Category: "configure",
			Module:   "git",
			Name:     "lfs",
			Success:  true,
			Message:  "enabled",
		})
	}

	return results
}

// =============================================================================
// Editor
// =============================================================================

func applyEditor(cfg *config.PactConfig) []Result {
	var results []Result

	defaultEditor := cfg.GetString("editor.default")

	// Install editor if possible
	if defaultEditor != "" {
		result := installEditor(defaultEditor)
		results = append(results, result)
	}

	// Install extensions
	extensions := cfg.GetStringSlice("editor.extensions")
	if len(extensions) > 0 {
		for _, ext := range extensions {
			result := installExtension(defaultEditor, ext)
			results = append(results, result)
		}
	}

	// Also check for vscode/cursor specific extensions
	vscodeExts := cfg.GetStringSlice("editor.vscode.extensions")
	for _, ext := range vscodeExts {
		result := installExtension("vscode", ext)
		results = append(results, result)
	}

	cursorExts := cfg.GetStringSlice("editor.cursor.extensions")
	for _, ext := range cursorExts {
		result := installExtension("cursor", ext)
		results = append(results, result)
	}

	return results
}

func installEditor(editor string) Result {
	result := Result{
		Category: "install",
		Module:   "editor",
		Name:     editor,
	}

	// Check if already installed
	var checkCmd string
	switch editor {
	case "code", "vscode":
		checkCmd = "code"
	case "cursor":
		checkCmd = "cursor"
	case "zed":
		checkCmd = "zed"
	case "nvim", "neovim":
		checkCmd = "nvim"
	case "vim":
		checkCmd = "vim"
	default:
		result.Success = true
		result.Skipped = true
		result.Message = "manual install required"
		return result
	}

	if isToolInstalled(checkCmd) {
		result.Success = true
		result.Skipped = true
		result.Message = "already installed"
		return result
	}

	// Try to install via package manager
	pm := detectPackageManager()
	if pm == "" {
		result.Success = true
		result.Skipped = true
		result.Message = "manual install required"
		return result
	}

	// Map editor names to package names
	pkgName := editor
	switch editor {
	case "vscode", "code":
		if pm == "brew" {
			pkgName = "visual-studio-code"
		}
	case "neovim":
		pkgName = "neovim"
	}

	installResult := installTool(pm, pkgName)
	result.Success = installResult.Success
	result.Skipped = installResult.Skipped
	result.Message = installResult.Message
	result.Error = installResult.Error
	return result
}

func installExtension(editor, extension string) Result {
	result := Result{
		Category: "extension",
		Module:   "editor",
		Name:     extension,
	}

	var cmd *exec.Cmd
	switch editor {
	case "code", "vscode":
		cmd = exec.Command("code", "--install-extension", extension, "--force")
	case "cursor":
		cmd = exec.Command("cursor", "--install-extension", extension, "--force")
	default:
		result.Success = true
		result.Skipped = true
		result.Message = "extensions not supported for this editor"
		return result
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed
		if strings.Contains(string(output), "already installed") {
			result.Success = true
			result.Skipped = true
			result.Message = "already installed"
			return result
		}
		result.Error = fmt.Errorf("%v: %s", err, string(output))
		return result
	}

	result.Success = true
	result.Message = "installed"
	return result
}

// =============================================================================
// Terminal & Fonts
// =============================================================================

func applyTerminal(cfg *config.PactConfig) []Result {
	var results []Result

	font := cfg.GetString("terminal.font")
	if font != "" {
		result := installNerdFont(font)
		results = append(results, result)
	}

	return results
}

func installNerdFont(fontName string) Result {
	result := Result{
		Category: "font",
		Module:   "terminal",
		Name:     fontName,
	}

	// Normalize font name for nerd-fonts
	nerdFontName := strings.ReplaceAll(fontName, " ", "")
	nerdFontName = strings.ReplaceAll(nerdFontName, "Nerd Font", "")
	nerdFontName = strings.ReplaceAll(nerdFontName, "NerdFont", "")
	nerdFontName = strings.TrimSpace(nerdFontName)

	// Check if font is already installed
	if isFontInstalled(fontName) {
		result.Success = true
		result.Skipped = true
		result.Message = "already installed"
		return result
	}

	switch runtime.GOOS {
	case "darwin":
		// Use Homebrew cask
		pm := detectPackageManager()
		if pm == "brew" {
			// Try the font cask name
			caskName := "font-" + strings.ToLower(nerdFontName) + "-nerd-font"
			cmd := exec.Command("brew", "install", "--cask", caskName)
			output, err := cmd.CombinedOutput()
			if err != nil {
				// Try alternative naming
				caskName = "font-" + strings.ToLower(strings.ReplaceAll(nerdFontName, "Mono", "-mono")) + "-nerd-font"
				cmd = exec.Command("brew", "install", "--cask", caskName)
				output, err = cmd.CombinedOutput()
				if err != nil {
					result.Error = fmt.Errorf("failed to install font: %s", string(output))
					return result
				}
			}
			result.Success = true
			result.Message = "installed via Homebrew"
			return result
		}

	case "linux":
		// Download from nerd-fonts releases
		home, _ := os.UserHomeDir()
		fontDir := filepath.Join(home, ".local/share/fonts")
		os.MkdirAll(fontDir, 0755)

		downloadURL := fmt.Sprintf("https://github.com/ryanoasis/nerd-fonts/releases/latest/download/%s.zip", nerdFontName)
		tmpFile := filepath.Join(os.TempDir(), nerdFontName+".zip")

		if err := downloadFile(downloadURL, tmpFile); err != nil {
			result.Error = err
			return result
		}
		defer os.Remove(tmpFile)

		if err := extractZip(tmpFile, fontDir, ""); err != nil {
			result.Error = err
			return result
		}

		// Refresh font cache
		exec.Command("fc-cache", "-fv").Run()

		result.Success = true
		result.Message = "installed to ~/.local/share/fonts"
		return result

	case "windows":
		// Download and install to Windows fonts folder
		downloadURL := fmt.Sprintf("https://github.com/ryanoasis/nerd-fonts/releases/latest/download/%s.zip", nerdFontName)
		tmpFile := filepath.Join(os.TempDir(), nerdFontName+".zip")

		if err := downloadFile(downloadURL, tmpFile); err != nil {
			result.Error = err
			return result
		}
		defer os.Remove(tmpFile)

		home, _ := os.UserHomeDir()
		fontDir := filepath.Join(home, "AppData/Local/Microsoft/Windows/Fonts")
		os.MkdirAll(fontDir, 0755)

		if err := extractZip(tmpFile, fontDir, ""); err != nil {
			result.Error = err
			return result
		}

		result.Success = true
		result.Message = "installed to Windows Fonts"
		return result
	}

	result.Error = fmt.Errorf("font installation not supported on this OS")
	return result
}

func isFontInstalled(fontName string) bool {
	switch runtime.GOOS {
	case "darwin":
		// Check system font directories
		fontDirs := []string{
			"/Library/Fonts",
			"/System/Library/Fonts",
			filepath.Join(os.Getenv("HOME"), "Library/Fonts"),
		}
		for _, dir := range fontDirs {
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if strings.Contains(strings.ToLower(e.Name()), strings.ToLower(strings.ReplaceAll(fontName, " ", ""))) {
					return true
				}
			}
		}
	case "linux":
		cmd := exec.Command("fc-list", ":", "family")
		output, _ := cmd.Output()
		return strings.Contains(strings.ToLower(string(output)), strings.ToLower(fontName))
	case "windows":
		home, _ := os.UserHomeDir()
		fontDirs := []string{
			"C:\\Windows\\Fonts",
			filepath.Join(home, "AppData/Local/Microsoft/Windows/Fonts"),
		}
		for _, dir := range fontDirs {
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if strings.Contains(strings.ToLower(e.Name()), strings.ToLower(strings.ReplaceAll(fontName, " ", ""))) {
					return true
				}
			}
		}
	}
	return false
}

// =============================================================================
// Apps
// =============================================================================

func applyApps(cfg *config.PactConfig) []Result {
	var results []Result

	currentOS := runtime.GOOS
	appsKey := fmt.Sprintf("apps.%s", currentOS)

	// Get apps for current OS
	appsMap := cfg.GetMap(appsKey)
	if appsMap == nil {
		return results
	}

	// Check for install list
	if installList, ok := appsMap["install"].([]any); ok {
		for _, app := range installList {
			if appName, ok := app.(string); ok {
				result := installApp(appName)
				results = append(results, result)
			}
		}
	}

	// Check for shortcuts (just note them, don't install)
	if shortcuts, ok := appsMap["shortcuts"].(map[string]any); ok {
		for name := range shortcuts {
			results = append(results, Result{
				Category: "app",
				Module:   "apps",
				Name:     name,
				Success:  true,
				Skipped:  true,
				Message:  "shortcut configured",
			})
		}
	}

	return results
}

func installApp(appName string) Result {
	result := Result{
		Category: "app",
		Module:   "apps",
		Name:     appName,
	}

	pm := detectPackageManager()
	if pm == "" {
		result.Error = fmt.Errorf("no package manager available")
		return result
	}

	// Map common app names to package names
	pkgMap := map[string]map[string]string{
		"brave": {
			"brew":   "brave-browser",
			"winget": "Brave.Brave",
			"choco":  "brave",
		},
		"discord": {
			"brew":   "discord",
			"winget": "Discord.Discord",
			"choco":  "discord",
		},
		"spotify": {
			"brew":   "spotify",
			"winget": "Spotify.Spotify",
			"choco":  "spotify",
		},
		"steam": {
			"brew":   "steam",
			"winget": "Valve.Steam",
			"choco":  "steam",
		},
		"cursor": {
			"brew":   "cursor",
			"winget": "Cursor.Cursor",
		},
		"vscode": {
			"brew":   "visual-studio-code",
			"winget": "Microsoft.VisualStudioCode",
			"choco":  "vscode",
		},
		"slack": {
			"brew":   "slack",
			"winget": "SlackTechnologies.Slack",
			"choco":  "slack",
		},
		"notion": {
			"brew":   "notion",
			"winget": "Notion.Notion",
			"choco":  "notion",
		},
		"figma": {
			"brew":   "figma",
			"winget": "Figma.Figma",
			"choco":  "figma",
		},
		"docker": {
			"brew":   "docker",
			"winget": "Docker.DockerDesktop",
			"choco":  "docker-desktop",
		},
	}

	// Get the package name for this package manager
	pkgName := appName
	if pkgs, ok := pkgMap[strings.ToLower(appName)]; ok {
		if pkg, ok := pkgs[pm]; ok {
			pkgName = pkg
		}
	}

	// Check if already installed (simplified check)
	if isToolInstalled(strings.ToLower(appName)) {
		result.Success = true
		result.Skipped = true
		result.Message = "already installed"
		return result
	}

	var cmd *exec.Cmd
	switch pm {
	case "brew":
		cmd = exec.Command("brew", "install", "--cask", pkgName)
	case "winget":
		cmd = exec.Command("winget", "install", "--id", pkgName, "-e", "--silent", "--accept-package-agreements", "--accept-source-agreements")
	case "choco":
		cmd = exec.Command("choco", "install", pkgName, "-y")
	case "scoop":
		cmd = exec.Command("scoop", "install", pkgName)
	default:
		result.Error = fmt.Errorf("app installation not supported for %s", pm)
		return result
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = fmt.Errorf("%v: %s", err, string(output))
		return result
	}

	result.Success = true
	result.Message = "installed"
	return result
}

// =============================================================================
// LLM
// =============================================================================

func applyLLM(cfg *config.PactConfig) []Result {
	var results []Result

	// Install local runtime
	localRuntime := cfg.GetString("llm.local.runtime")
	if localRuntime != "" {
		pm := detectPackageManager()
		if pm != "" {
			result := installTool(pm, localRuntime)
			results = append(results, result)
		}

		// Pull local models
		models := cfg.GetStringSlice("llm.local.models")
		for _, model := range models {
			result := pullOllamaModel(localRuntime, model)
			results = append(results, result)
		}
	}

	return results
}

func pullOllamaModel(runtime, model string) Result {
	result := Result{
		Category: "configure",
		Module:   "llm",
		Name:     model,
	}

	if runtime != "ollama" {
		result.Skipped = true
		result.Success = true
		result.Message = "only ollama supported for model pulling"
		return result
	}

	// Check if ollama is installed
	if !isToolInstalled("ollama") {
		result.Error = fmt.Errorf("ollama not installed")
		return result
	}

	// Check if model already exists
	cmd := exec.Command("ollama", "list")
	output, _ := cmd.Output()
	if strings.Contains(string(output), model) {
		result.Success = true
		result.Skipped = true
		result.Message = "already pulled"
		return result
	}

	// Skip pulling for now - it takes too long for sync
	// User can run `ollama pull <model>` manually
	result.Success = true
	result.Skipped = true
	result.Message = fmt.Sprintf("run 'ollama pull %s' to download", model)
	return result
}

// =============================================================================
// Files
// =============================================================================

func applyFiles(cfg *config.PactConfig) []Result {
	var results []Result

	items, err := cfg.GetSyncItems()
	if err != nil {
		return results
	}

	for _, item := range items {
		result := syncFile(item)
		results = append(results, result)
	}

	return results
}

func applyModuleFiles(cfg *config.PactConfig, module string) []Result {
	var results []Result

	items, err := cfg.GetSyncItemsForModule(module)
	if err != nil {
		return results
	}

	for _, item := range items {
		result := syncFile(item)
		results = append(results, result)
	}

	return results
}

func syncFile(item config.SyncItem) Result {
	result := Result{
		Category: "file",
		Module:   item.Module,
		Name:     item.Name,
	}

	if _, err := os.Stat(item.Source); os.IsNotExist(err) {
		result.Error = fmt.Errorf("source not found: %s", item.Source)
		return result
	}

	strategy := item.Strategy
	if strategy == "" {
		strategy = "symlink"
	}

	targetDir := filepath.Dir(item.Target)
	os.MkdirAll(targetDir, 0755)

	os.RemoveAll(item.Target)

	switch strategy {
	case "symlink":
		if err := os.Symlink(item.Source, item.Target); err != nil {
			result.Error = err
			return result
		}
		result.Message = fmt.Sprintf("symlinked -> %s", item.Source)
	case "copy":
		cmd := exec.Command("cp", "-r", item.Source, item.Target)
		if err := cmd.Run(); err != nil {
			result.Error = err
			return result
		}
		result.Message = fmt.Sprintf("copied from %s", item.Source)
	default:
		result.Error = fmt.Errorf("unknown strategy: %s", strategy)
		return result
	}

	result.Success = true
	return result
}

// =============================================================================
// Helpers
// =============================================================================

func detectPackageManager() string {
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("brew"); err == nil {
			return "brew"
		}
	case "linux":
		if _, err := exec.LookPath("apt"); err == nil {
			return "apt"
		}
		if _, err := exec.LookPath("dnf"); err == nil {
			return "dnf"
		}
		if _, err := exec.LookPath("pacman"); err == nil {
			return "pacman"
		}
		if _, err := exec.LookPath("brew"); err == nil {
			return "brew"
		}
	case "windows":
		if _, err := exec.LookPath("winget"); err == nil {
			return "winget"
		}
		if _, err := exec.LookPath("scoop"); err == nil {
			return "scoop"
		}
		if _, err := exec.LookPath("choco"); err == nil {
			return "choco"
		}
	}
	return ""
}

func isToolInstalled(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

func installTool(pm, tool string) Result {
	result := Result{
		Category: "install",
		Module:   "cli",
		Name:     tool,
	}

	if isToolInstalled(tool) {
		result.Success = true
		result.Skipped = true
		result.Message = "already installed"
		return result
	}

	var cmd *exec.Cmd
	switch pm {
	case "brew":
		cmd = exec.Command("brew", "install", tool)
	case "apt":
		cmd = exec.Command("sudo", "apt", "install", "-y", tool)
	case "dnf":
		cmd = exec.Command("sudo", "dnf", "install", "-y", tool)
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", tool)
	case "winget":
		cmd = exec.Command("winget", "install", "--id", tool, "-e", "--silent")
	case "scoop":
		cmd = exec.Command("scoop", "install", tool)
	case "choco":
		cmd = exec.Command("choco", "install", tool, "-y")
	default:
		result.Error = fmt.Errorf("unsupported package manager: %s", pm)
		return result
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = fmt.Errorf("%v: %s", err, string(output))
		return result
	}

	result.Success = true
	result.Message = "installed"
	return result
}

func runGitConfig(key, value string) error {
	return exec.Command("git", "config", "--global", key, value).Run()
}

func downloadPromptTheme(promptTool, themeName, source string) Result {
	result := Result{
		Category: "configure",
		Module:   "shell",
		Name:     fmt.Sprintf("%s-theme", promptTool),
	}

	var themeDir string
	home, _ := os.UserHomeDir()

	switch promptTool {
	case "oh-my-posh":
		switch runtime.GOOS {
		case "darwin", "linux":
			themeDir = filepath.Join(home, ".config/oh-my-posh/themes")
		case "windows":
			themeDir = filepath.Join(home, "AppData/Local/Programs/oh-my-posh/themes")
		}
	case "starship":
		themeDir = filepath.Join(home, ".config")
	default:
		result.Skipped = true
		result.Message = "unknown prompt tool"
		return result
	}

	os.MkdirAll(themeDir, 0755)

	themePath := filepath.Join(themeDir, themeName+".omp.json")

	if _, err := os.Stat(themePath); err == nil {
		result.Success = true
		result.Skipped = true
		result.Message = "theme already exists"
		return result
	}

	cmd := exec.Command("curl", "-sSL", "-o", themePath, source)
	if output, err := cmd.CombinedOutput(); err != nil {
		result.Error = fmt.Errorf("failed to download theme: %v: %s", err, string(output))
		return result
	}

	result.Success = true
	result.Message = fmt.Sprintf("downloaded to %s", themePath)
	return result
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractTarGz(src, destDir, binaryName string) error {
	cmd := exec.Command("tar", "-xzf", src, "-C", destDir)
	return cmd.Run()
}

func extractZip(src, destDir, binaryName string) error {
	cmd := exec.Command("unzip", "-o", src, "-d", destDir)
	return cmd.Run()
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0755)
}
