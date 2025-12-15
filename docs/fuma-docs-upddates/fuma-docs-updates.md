# Pact Documentation

> Your portable dev identity. Shell, editor, AI prefs, themes — one kit, any machine.

## Overview

Pact stores your entire development environment configuration in a single GitHub repo. Edit locally or in the browser, sync from terminal, apply anywhere.

## Installation

### macOS/Linux (Homebrew)

```bash
brew install cloudboy-jh/tap/pact
```

### From GitHub Releases

Download the latest release from [GitHub Releases](https://github.com/cloudboy-jh/pact/releases).

### From Source

```bash
cd cli
go build -o pact .
sudo mv pact /usr/local/bin/
```

## Quick Start

```bash
cd my-project

# Initialize (authenticates with GitHub, clones your pact repo)
pact init

# Sync all modules - installs tools, configures shell, git, etc.
pact sync all

# Or sync specific modules
pact sync shell    # Install shell tools, configure prompt
pact sync cli      # Install CLI tools
pact sync git      # Configure git settings

# Check status
pact status
```

## Commands

### `pact`

Interactive status display with quick actions.

**Key Bindings:**
- `s` - Run sync
- `e` - Edit pact.json
- `r` - Refresh
- `q` - Quit

### `pact init`

Authenticate with GitHub and clone/create your pact repo.

```bash
pact init
```

Creates a `.pact/` folder in your current directory (like `.git/`).

### `pact sync [module]`

Pull latest changes and apply configurations.

```bash
# Interactive mode - pick modules to sync
pact sync

# Sync everything
pact sync all

# Sync specific modules
pact sync shell     # Shell tools + prompt
pact sync cli       # CLI tools
pact sync git       # Git configuration
pact sync editor    # Editor setup
pact sync terminal  # Fonts
pact sync llm       # Local LLM runtime
pact sync apps      # Applications
```

#### What Gets Installed/Configured

| Module | Actions |
|--------|---------|
| `shell` | Install oh-my-posh/starship, download theme, install zoxide/fzf, inject init into .zshrc |
| `cli` | Install tools via brew/apt/winget (bun, node, lazygit, etc.) |
| `git` | Set user.name, user.email, init.defaultBranch, enable LFS |
| `editor` | Install editor, install extensions (VSCode/Cursor) |
| `terminal` | Install Nerd Fonts |
| `llm` | Install Ollama, show commands to pull models |
| `apps` | Install apps via brew cask/winget |

### `pact push`

Commit and push local changes to GitHub.

```bash
pact push -m "Update shell config"
```

### `pact edit [path]`

Edit pact files.

```bash
pact edit           # Edit pact.json in $EDITOR
pact edit web       # Open web editor in browser
pact edit shell     # Edit shell configs
```

### `pact status`

Show current status (non-interactive).

### `pact secret`

Manage secrets stored in OS keychain.

```bash
pact secret set ANTHROPIC_API_KEY    # Store a secret
pact secret list                      # List secrets status
pact secret remove ANTHROPIC_API_KEY  # Remove a secret
```

### `pact reset`

Remove all symlinks created by pact (keeps `.pact/` intact).

### `pact nuke`

Complete removal (symlinks + `.pact/` + token).

```bash
pact nuke           # With confirmation
pact nuke --force   # Skip confirmation
```

## Configuration

### pact.json Schema

Your `pact.json` is flexible - add whatever you want. Pact reads it loosely and extracts what it can.

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
    "theme": "Catppuccin",
    "extensions": ["esbenp.prettier-vscode", "dbaeumer.vscode-eslint"]
  },

  "llm": {
    "providers": ["claude", "openai"],
    "local": {
      "runtime": "ollama",
      "models": ["qwen-coder", "mistral"]
    }
  },

  "cli": {
    "tools": ["bun", "node", "lazygit", "ripgrep", "fzf"],
    "custom": ["pact", "churn"]
  },

  "apps": {
    "darwin": {
      "install": ["brave", "discord", "spotify"]
    },
    "windows": {
      "install": ["brave", "discord", "spotify"]
    }
  },

  "secrets": [
    "ANTHROPIC_API_KEY",
    "OPENAI_API_KEY"
  ]
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
        "target": {
          "darwin": "~/.zshrc",
          "linux": "~/.zshrc"
        }
      }
    }
  },

  "editor": {
    "default": "vscode",
    "files": {
      "settings": {
        "source": "editor/vscode/settings.json",
        "target": {
          "darwin": "~/Library/Application Support/Code/User/settings.json",
          "linux": "~/.config/Code/User/settings.json",
          "windows": "~/AppData/Roaming/Code/User/settings.json"
        }
      }
    }
  }
}
```

### Sync Strategies

| Strategy | Behavior |
|----------|----------|
| `symlink` (default) | Creates symlink from target to source |
| `copy` | Copies file to target location |

```json
{
  "files": {
    "config": {
      "source": "some/config.json",
      "target": "~/.config/app/config.json",
      "strategy": "copy"
    }
  }
}
```

## Cross-Platform Support

Pact works on macOS, Linux, and Windows.

### Package Managers

| OS | Package Managers |
|----|------------------|
| macOS | Homebrew |
| Linux | apt, dnf, pacman, Homebrew |
| Windows | winget, scoop, chocolatey |

### OS-Specific Targets

```json
{
  "target": {
    "darwin": "~/.config/app",
    "linux": "~/.config/app",
    "windows": "~/AppData/Local/app"
  }
}
```

## Secrets

Secrets are stored in your OS keychain, never in the repo.

| OS | Backend |
|----|---------|
| macOS | Keychain |
| Linux | libsecret / gnome-keyring |
| Windows | Windows Credential Manager |

## Repository Structure

```
username/my-pact/
├── pact.json              # Main config
├── shell/                 # Shell configs
│   ├── .zshrc
│   └── profile.ps1
├── editor/                # Editor configs
│   ├── vscode/
│   └── nvim/
├── git/                   # Git configs
│   ├── .gitconfig
│   └── .gitignore_global
├── terminal/              # Terminal configs
└── prompts/               # AI prompts
```

## How It Works

1. **GitHub is the database** - Your `my-pact` repo stores everything
2. **Local `.pact/` folder** - Cloned repo lives in your project (like `.git/`)
3. **Token in keychain** - One auth works across all projects
4. **Sync applies configs** - Installs tools, creates symlinks, configures apps

### Sync Flow

```
pact sync all
    │
    ├── git pull (get latest from GitHub)
    │
    ├── cli module
    │   └── brew install bun node lazygit...
    │
    ├── shell module
    │   ├── brew install oh-my-posh zoxide
    │   ├── download theme
    │   └── inject init into .zshrc
    │
    ├── git module
    │   ├── git config --global user.name
    │   ├── git config --global user.email
    │   └── git lfs install
    │
    ├── terminal module
    │   └── install Nerd Font
    │
    └── files
        └── symlink dotfiles
```

## Web Editor

Edit your pact config in the browser:

```bash
pact edit web
```

Or visit [pact-ckn.pages.dev](https://pact-ckn.pages.dev).

## Development

### Building the CLI

```bash
cd cli
go mod tidy
go build -o pact .
./pact --help
```

### Running Tests

```bash
cd cli
go test ./...
```

## License

MIT
