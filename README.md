<p align="center">
  <img src="Pact-readme-logo.png" alt="Pact Logo" width="400" />
</p>

# Pact

> Your portable dev identity. Shell, editor, AI prefs, themes — one kit, any machine.

Pact stores your entire development environment configuration in a single GitHub repo. Edit locally or in the browser, sync from terminal, apply anywhere.

## What is Pact?

### The Problem

Setting up a new machine sucks. Dotfiles repos are messy, unstructured, and hard to share. Your AI prompts, editor configs, and shell setup live in different places.

### The Solution

One `pact.json` manifest + organized files in a GitHub repo. CLI to edit and sync, web UI for visual editing, cross-OS support built in.

**Pact sync doesn't just symlink files — it installs tools, configures your shell, sets up git, downloads fonts, and more.**

### What's in a Pact

```
username/my-pact/
├── pact.json              # Your portable dev identity
├── shell/                 # .zshrc, .bashrc, profile.ps1
├── editor/                # nvim, vscode, cursor configs
├── terminal/              # Ghostty, Kitty, Alacritty
├── git/                   # .gitconfig, .gitignore_global
├── prompts/               # AI prompts (default.md, code-review.md)
├── skills/                # Custom AI skills
├── agents/                # CLAUDE.md, .cursorrules
├── tools/                 # lazygit, ripgrep, fzf configs
├── keybindings/           # Editor keybindings
├── snippets/              # Code snippets
└── theme/                 # Colors, wallpapers, icons
```

---

## User Guide

### Install

**macOS (Homebrew)**
```bash
brew install cloudboy-jh/tap/pact
```

**Linux (Homebrew)**
```bash
brew install cloudboy-jh/tap/pact
```

**Windows (Scoop)**
```powershell
scoop bucket add pact https://github.com/cloudboy-jh/pact-bucket
scoop install pact
```

**Linux/macOS (curl)**
```bash
curl -fsSL https://github.com/cloudboy-jh/pact/releases/latest/download/pact_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/').tar.gz | tar -xz -C /usr/local/bin pact
```

**From Source**
```bash
cd cli && go build -o pact . && sudo mv pact /usr/local/bin/
```

### Getting Started

```bash
cd my-project

# 1. Initialize (authenticates with GitHub, clones your pact repo to ./.pact/)
pact init

# 2. Sync everything - installs tools, configures shell, git, fonts, etc.
pact sync all

# Or sync specific modules
pact sync shell    # Install oh-my-posh, zoxide, configure prompt
pact sync cli      # Install bun, node, lazygit, etc.
pact sync git      # Configure git user, email, default branch

# 3. Check status
pact status
```

Pact works like `git` — it creates a `.pact/` folder in your project and walks up the directory tree to find it. Your GitHub token is stored globally in your OS keychain.

### Commands

| Command | Description |
|---------|-------------|
| `pact` | Interactive status with quick actions (s/e/q) |
| `pact init` | Authenticate with GitHub + setup your pact repo |
| `pact sync` | Interactive module picker - select which modules to apply |
| `pact sync all` | Apply everything |
| `pact sync <module>` | Apply specific module (shell, cli, git, editor, terminal, llm, apps) |
| `pact edit` | Edit pact.json in $EDITOR |
| `pact edit web` | Open web editor in browser |
| `pact push` | Commit and push local changes |
| `pact status` | Show status (non-interactive) |
| `pact secret set <name>` | Store a secret in OS keychain |
| `pact secret list` | List secrets and their status |
| `pact reset` | Remove all symlinks (keeps .pact/) |
| `pact nuke` | Full cleanup (symlinks + .pact/ + token) |

### What `pact sync` Does

| Module | What Gets Installed/Configured |
|--------|-------------------------------|
| `shell` | oh-my-posh/starship, downloads theme, zoxide/fzf, injects init into .zshrc |
| `cli` | Tools via brew/apt/winget (bun, node, lazygit, etc.) |
| `git` | Sets user.name, user.email, init.defaultBranch, enables LFS |
| `editor` | Installs editor, installs VSCode/Cursor extensions |
| `terminal` | Installs Nerd Fonts automatically |
| `llm` | Installs Ollama, shows commands to pull local models |
| `apps` | Installs apps via brew cask/winget (brave, discord, spotify, etc.) |

### Example Sync Output

```bash
$ pact sync all

Pulling latest changes...
✓ Pulled latest changes

Applying shell...
Applying cli...
Applying git...
Applying terminal...

Installations:
  ○ bun                  already installed
  ○ node                 already installed
  ✓ lazygit              installed
  ○ oh-my-posh           already installed
  ✓ zoxide               installed

Configuration:
  ✓ shell.oh-my-posh-theme downloaded
  ✓ shell.shell-config   added to .zshrc
  ✓ git.user.name        cloudboy-jh
  ✓ git.user.email       you@example.com
  ✓ git.init.defaultBranch main

Fonts:
  ○ JetBrainsMono Nerd Font already installed

Done: 6 applied, 8 skipped, 0 failed
```

### pact.json Example

Your config is flexible — add whatever you want:

```json
{
  "name": "your-username",
  "version": "1.0.0",

  "shell": {
    "prompt": {
      "tool": "oh-my-posh",
      "theme": "capr4n",
      "source": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/capr4n.omp.json"
    },
    "tools": ["zoxide", "fzf"]
  },

  "git": {
    "user": "your-username",
    "email": "you@example.com",
    "defaultBranch": "main",
    "lfs": true
  },

  "terminal": {
    "font": "JetBrainsMono Nerd Font",
    "fontSize": 14
  },

  "editor": {
    "default": "zed",
    "extensions": ["esbenp.prettier-vscode"]
  },

  "llm": {
    "providers": ["claude", "openai"],
    "local": {
      "runtime": "ollama",
      "models": ["qwen-coder", "mistral"]
    }
  },

  "cli": {
    "tools": ["bun", "node", "lazygit", "ripgrep"],
    "custom": ["pact", "churn"]
  },

  "apps": {
    "darwin": {
      "install": ["brave", "discord", "spotify"]
    }
  },

  "secrets": ["ANTHROPIC_API_KEY", "OPENAI_API_KEY"]
}
```

### File Syncing

Add `files` entries to any module to sync dotfiles:

```json
{
  "shell": {
    "tools": ["zoxide"],
    "files": {
      "zshrc": {
        "source": "shell/.zshrc",
        "target": "~/.zshrc"
      }
    }
  }
}
```

OS-specific targets:

```json
{
  "target": {
    "darwin": "~/.config/app",
    "linux": "~/.config/app",
    "windows": "~/AppData/Local/app"
  }
}
```

### Secrets

Secrets are stored in your OS keychain, never in the repo:

```bash
pact secret set ANTHROPIC_API_KEY
# Enter value: ****
# Stored in keychain

pact secret list
#   ANTHROPIC_API_KEY    ● set
#   OPENAI_API_KEY       ○ not set
```

| OS | Backend |
|----|---------|
| macOS | Keychain |
| Linux | libsecret / gnome-keyring |
| Windows | Windows Credential Manager |

### Cross-Platform Support

Pact works on macOS, Linux, and Windows with automatic package manager detection:

| OS | Package Managers |
|----|------------------|
| macOS | Homebrew |
| Linux | apt, dnf, pacman, Homebrew |
| Windows | winget, scoop, chocolatey |

---

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- A GitHub OAuth App (for authentication)

### Building the CLI

```bash
cd cli
go mod tidy
go build -o pact .
./pact --help
```

### Running the Web App

```bash
cd web
npm install
npm run dev
# Open http://localhost:5173
```

### Project Structure

```
pact/
├── cli/                    # Go CLI
│   ├── cmd/                # Cobra commands
│   ├── internal/
│   │   ├── apply/          # Tool installation & config logic
│   │   ├── auth/           # GitHub OAuth device flow
│   │   ├── config/         # pact.json parsing
│   │   ├── git/            # Git operations
│   │   ├── keyring/        # OS keychain
│   │   ├── sync/           # Symlink/copy logic
│   │   └── ui/             # TUI (Lip Gloss)
│   ├── go.mod
│   └── main.go
│
└── web/                    # SvelteKit web app
    ├── src/
    │   ├── lib/
    │   │   ├── github.ts   # GitHub API client
    │   │   └── stores/     # Auth state
    │   └── routes/
    ├── package.json
    └── svelte.config.js
```

### Tech Stack

| Layer | Tech |
|-------|------|
| CLI | Go, Cobra, Lip Gloss |
| CLI Git | go-git |
| CLI Secrets | go-keyring |
| Web | SvelteKit, Tailwind CSS |
| Auth | GitHub OAuth |
| Storage | GitHub repo (user/my-pact) |

---

## Design Principles

1. **GitHub is the database** — No separate backend, your repo is the source of truth
2. **Actually apply configs** — Not just symlinks, but installs tools and configures apps
3. **Cross-OS by default** — Darwin, Windows, Linux support built in
4. **Secrets stay local** — API keys in OS keychain, never in repo
5. **Flexible config** — Your pact.json, your structure

---

## Releases

Releases are automatic on every push to `master` that changes CLI code.

| Version | Date | Notes |
|---------|------|-------|
| v0.2.x | Dec 2025 | Full apply system - installs tools, fonts, apps, configures shell/git |
| v0.1.x | Dec 2025 | Interactive sync, theme module, local .pact/ |
| v0.1.0 | Dec 2025 | Initial release with Homebrew support |

---

## License

MIT
