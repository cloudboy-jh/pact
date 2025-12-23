# Changelog

All notable changes to this project will be documented in this file.

## [0.3.0] - 2025-12-22

### Added
- **`pact read` command** - Reverse sync to scan local environment and import to pact.json
  - Detects installed CLI tools, shell prompt, git config, editors, LLM providers, secrets
  - Discovers config files (.zshrc, .gitconfig, nvim/, vscode settings, etc.)
  - Shows drift between local machine and existing pact.json
  - Hierarchical TUI picker for selecting modules and individual items to import
  - Copies selected config files to `.pact/` directory
  - Auto-stores detected secrets in OS keychain
- **New flags for `pact read`**:
  - `--diff` - Only show differences, don't import
  - `--json` - Output detected config as JSON for scripting
  - `-y/--yes` - Import everything without prompts
  - `--dry-run` - Preview changes without modifying anything
- **Module filtering** - `pact read shell git` scans only specified modules
- **GitHub init prompt** - If pact not initialized, `pact read` offers to connect GitHub
- **New `detect` package** with cross-platform detection logic:
  - `detect/tools.go` - CLI tool detection
  - `detect/shell.go` - Shell prompt and tools detection
  - `detect/git.go` - Git config detection
  - `detect/editor.go` - Editor detection
  - `detect/llm.go` - Ollama and coding agents detection
  - `detect/secrets.go` - Environment variable scanning
  - `detect/configs.go` - Config file discovery and copying
  - `detect/diff.go` - Compare detected vs pact.json
  - `detect/merge.go` - Merge selections into pact.json
  - Platform-specific stubs for future expansion

### Changed
- Shell tools (zoxide, fzf) now detected separately from CLI tools for cleaner categorization

## [0.2.1] - 2024-12-17

### Added
- **Scrollable module list** - Status view now paginates when content exceeds terminal height
  - Arrow keys or `j`/`k` (vim-style) to scroll
  - Shows "... N more above/below" indicators when scrolled
- **Editor choice prompt** - Pressing `e` now prompts for web or local editor if not configured
  - Option to open web editor at pact-dev.com
  - Option to use local editor (respects `EDITOR`/`VISUAL` env vars, falls back to OS defaults)
  - Supports custom editor via `editor.default` in pact.json (e.g., `"web"`, `"local"`, `"code"`, `"vim"`)

### Changed
- Help bar now shows `[r] refresh` option
- Terminal dimensions are now detected for proper pagination

### Fixed
- Module list no longer scrolls infinitely without bounds
- Editor action now properly falls back based on OS (notepad on Windows, open -t on macOS, nano/vim on Linux)

## [0.2.0] - 2024-12-14

### Added
- **Full apply system** - `pact sync` now installs and configures your entire dev environment
- **Tool installation** via brew/apt/winget for CLI tools (bun, node, lazygit, etc.)
- **Custom tool installation** from GitHub releases (pact, churn, annotr)
- **Shell configuration**:
  - Installs oh-my-posh/starship
  - Downloads prompt themes automatically
  - Injects init commands into .zshrc/.bashrc/PowerShell profile
  - Installs shell tools (zoxide, fzf, direnv)
- **Git configuration** - Sets user.name, user.email, init.defaultBranch, enables LFS
- **Nerd Font installation** - Auto-downloads and installs fonts via brew cask or direct download
- **Editor support** - Installs editors and VSCode/Cursor extensions
- **App installation** via brew cask/winget (brave, discord, spotify, etc.)
- **Ollama integration** - Installs ollama and manages local models
- **Interactive status** (`pact status`) with key bindings:
  - `s` - Run sync
  - `e` - Edit config
  - `r` - Refresh
  - `q` - Quit
- **Categorized sync output** - Shows installs, configuration, fonts, extensions, apps separately
- **Flexible pact.json parser** - Works with any config structure (not rigid schema)
- Fuma docs documentation

### Changed
- Config parser now uses `map[string]any` for flexibility instead of rigid Go structs
- `pact sync` shows helpful message when no files configured
- `pact sync all` applies all modules at once
- Module picker shows details about each module (tools, preferences)

### Fixed
- `pact sync` properly pulls and reads pact.json
- `pact status` now interactive (was exiting immediately)
- Config loading works with user's actual pact.json format

## [0.1.0] - 2024-12-11

### Added
- ASCII logo branding for CLI with colored output using lipgloss
- `--version` / `-v` flag to display logo and version
- Logo displayed on `pact init` welcome message
- Logo displayed on `pact --help` output
- Logo displayed when running `pact` without initialization
- Scoop package manager support for Windows users
- Automated Scoop bucket updates via GoReleaser on release
- Interactive module picker for `pact sync`
- Theme module support
- Local `.pact/` directory (like git)

### Changed
- Version variable moved to `internal/ui` package for centralized branding
