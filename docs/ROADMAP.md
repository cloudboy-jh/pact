# Pact Roadmap

## Completed

### Phase 1 - Core CLI
- [x] Set up Go CLI project structure and dependencies
- [x] Implement pact.json config parsing
- [x] Implement GitHub Device Flow OAuth
- [x] Implement keyring token storage
- [x] Implement `pact init` command
- [x] Implement `pact sync` command
- [x] Implement `pact push` command
- [x] Implement `pact edit` command
- [x] Implement interactive status TUI (root command)

### Phase 2 - Extended CLI
- [x] Implement `pact secret set/list/remove`
- [x] Implement copy strategy (in addition to symlink)
- [x] Implement partial sync (`pact sync <module>`)
- [x] Implement `pact reset` command
- [x] Implement `pact nuke` command

### Phase 3 - Web UI
- [x] Set up SvelteKit project with Tailwind
- [x] Implement landing page with GitHub OAuth
- [x] Implement OAuth callback and session management
- [x] Implement dashboard page
- [x] Implement module editor
- [x] Implement auto-create repo on first login

---

## Phase 4 - Polish & Testing

### Code Quality
- [ ] Add the `/editor/[module]` route if missing
- [ ] End-to-end testing - Test full flow: init → sync → edit in web → push
- [ ] Improve error handling with better messages in CLI/Web
- [ ] Add unit tests for core CLI functions
- [ ] Add integration tests for GitHub API interactions

### UX Improvements
- [ ] Add loading spinners/states in web UI
- [ ] Add success/error toast notifications
- [ ] Improve TUI with progress indicators during sync/push

---

## Phase 5 - New Features

### Conflict & Diff Management
- [ ] Conflict detection - Warn when local files differ from repo
- [ ] Diff viewer - Show what will change before sync
- [ ] Interactive conflict resolution in TUI

### Templates & Presets
- [ ] Module templates for common tools:
  - [ ] Git (.gitconfig, .gitignore_global)
  - [ ] SSH (~/.ssh/config)
  - [ ] VS Code (settings.json, keybindings.json)
  - [ ] Zsh/Bash (.zshrc, .bashrc, aliases)
  - [ ] Vim/Neovim (.vimrc, init.lua)
  - [ ] Windows Terminal (settings.json)
  - [ ] PowerShell (profile.ps1)

### Advanced Features
- [ ] Multi-repo support - Manage multiple pact repos
- [ ] Import/export configurations
- [ ] Backup before sync (local snapshots)
- [ ] Encrypted secrets in repo (age/sops)

---

## Phase 6 - Distribution

### CLI Distribution
- [ ] GitHub Actions workflow to build binaries for:
  - [ ] Windows (amd64, arm64)
  - [ ] macOS (amd64, arm64)
  - [ ] Linux (amd64, arm64)
- [ ] Homebrew formula (macOS/Linux)
- [ ] Scoop manifest (Windows)
- [ ] AUR package (Arch Linux)
- [ ] APT/RPM packages

### Web Deployment
- [ ] Deploy web UI to Vercel
- [ ] Custom domain setup
- [ ] Environment variable configuration
- [ ] Analytics integration

### Documentation
- [ ] User guide with examples
- [ ] API documentation
- [ ] Contributing guide
- [ ] Video walkthrough

---

## Future Ideas (Backlog)

- [ ] Team/org shared configurations
- [ ] Machine-specific overrides
- [ ] Plugin system for custom sync strategies
- [ ] Desktop app (Tauri/Electron)
- [ ] Mobile companion app for viewing configs
- [ ] Integration with dotfile managers (chezmoi, yadm)
- [ ] Cloud sync alternatives (beyond GitHub)
