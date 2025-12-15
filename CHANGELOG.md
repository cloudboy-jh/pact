# Changelog

All notable changes to this project will be documented in this file.

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
