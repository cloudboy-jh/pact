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

### What's in a Pact

```
username/my-pact/
├── pact.json              # Manifest
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
└── fonts/                 # Font preferences
```

---

## User Guide

### Install

**macOS (Homebrew)**
```bash
# Coming soon
brew install pact
```

**From GitHub Releases**
```bash
# Download the latest release for your platform from:
# https://github.com/cloudboy-jh/pact/releases

# macOS/Linux
curl -sSL https://github.com/cloudboy-jh/pact/releases/latest/download/pact-$(uname -s)-$(uname -m) -o pact
chmod +x pact
sudo mv pact /usr/local/bin/
```

**From Source**
```bash
cd cli
go build -o pact .
sudo mv pact /usr/local/bin/
```

### Getting Started

```bash
# 1. Initialize (authenticates with GitHub, creates/clones your pact repo)
pact init

# 2. Edit your configuration
pact edit                 # Opens pact.json in your $EDITOR
pact edit shell           # Edit shell configs
pact edit web             # Open web editor in browser

# 3. Sync configs to your system
pact sync

# 4. Check status
pact
```

### Commands

| Command | Description |
|---------|-------------|
| `pact` | Interactive status with quick actions |
| `pact init` | Authenticate with GitHub + setup your pact repo |
| `pact edit` | Edit pact.json in $EDITOR |
| `pact edit <path>` | Edit specific file/module (e.g., `pact edit shell`) |
| `pact edit web` | Open web editor in browser |
| `pact sync` | Pull latest + apply all module configs |
| `pact sync <module>` | Sync a specific module only |
| `pact push` | Commit and push local changes |
| `pact status` | Show status (non-interactive) |
| `pact secret set <name>` | Store a secret in OS keychain |
| `pact secret list` | List secrets and their status |
| `pact secret remove <name>` | Remove a secret from keychain |
| `pact reset` | Remove all symlinks (keeps ~/.pact/) |
| `pact nuke` | Full cleanup (symlinks + ~/.pact/ + token) |

### Editing Your Pact

**Option 1: Local Editor (CLI)**
```bash
pact edit                 # Edit pact.json
pact edit shell/zshrc     # Edit specific file
pact push                 # Commit and push changes
```

**Option 2: Web Editor**
```bash
pact edit web             # Opens browser
# Make changes in the web UI
pact sync                 # Pull changes to local
```

### Syncing Across Machines

On a new machine:
```bash
pact init                 # Clones your existing pact repo
pact sync                 # Applies all configs
```

Changes sync through GitHub — edit on one machine, pull on another.

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

### pact.json Example

```json
{
  "version": "1.0.0",
  "user": "your-username",
  "modules": {
    "shell": {
      "darwin": {
        "source": "./shell/darwin.zshrc",
        "target": "~/.zshrc",
        "strategy": "symlink"
      },
      "windows": {
        "source": "./shell/windows.ps1",
        "target": "~/Documents/PowerShell/profile.ps1",
        "strategy": "copy"
      }
    },
    "editor": {
      "neovim": {
        "source": "./editor/nvim/",
        "target": {
          "darwin": "~/.config/nvim",
          "linux": "~/.config/nvim",
          "windows": "~/AppData/Local/nvim"
        },
        "strategy": "symlink"
      }
    },
    "git": {
      "config": {
        "source": "./git/.gitconfig",
        "target": "~/.gitconfig",
        "strategy": "symlink"
      }
    }
  },
  "secrets": [
    "ANTHROPIC_API_KEY",
    "OPENAI_API_KEY",
    "GITHUB_TOKEN"
  ]
}
```

### Sync Strategies

| Strategy | Behavior | Use when |
|----------|----------|----------|
| `symlink` | Creates symlink from target to source | Edits in ~/.pact/ reflect immediately |
| `copy` | Copies file to target location | App doesn't follow symlinks |

### Cross-OS Support

Each module can have OS-specific configs:

```json
{
  "shell": {
    "darwin": { "source": "./shell/darwin.zshrc", "target": "~/.zshrc" },
    "linux": { "source": "./shell/linux.zshrc", "target": "~/.zshrc" },
    "windows": { "source": "./shell/windows.ps1", "target": "~/Documents/PowerShell/profile.ps1" }
  }
}
```

The CLI detects your current OS and applies the right config.

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

For local development, set the CLI to use the local web server:
```bash
export PACT_WEB_URL=http://localhost:5173
```

### GitHub OAuth Setup

1. Go to **GitHub Settings → Developer settings → OAuth Apps → New OAuth App**
2. Set:
   - **Application name**: `Pact`
   - **Homepage URL**: `http://localhost:5173`
   - **Authorization callback URL**: `http://localhost:5173/auth/callback`
3. Note your **Client ID** and generate a **Client Secret**

**CLI Configuration** — Edit `cli/internal/auth/oauth.go`:
```go
const ClientID = "your-client-id"
```

**Web Configuration** — Edit `web/src/lib/github.ts`:
```ts
export const GITHUB_CLIENT_ID = 'your-client-id';
```

Set environment variable:
```bash
export GITHUB_CLIENT_SECRET=your-client-secret
```

### Project Structure

```
pact/
├── cli/                    # Go CLI
│   ├── cmd/                # Cobra commands
│   ├── internal/
│   │   ├── auth/           # GitHub OAuth device flow
│   │   ├── config/         # pact.json parsing
│   │   ├── git/            # Git operations
│   │   ├── keyring/        # OS keychain
│   │   ├── sync/           # Symlink/copy logic
│   │   └── ui/             # TUI (Lip Gloss + Bubbletea)
│   ├── go.mod
│   └── main.go
│
└── web/                    # SvelteKit web app
    ├── src/
    │   ├── lib/
    │   │   ├── github.ts   # GitHub API client
    │   │   └── stores/     # Auth state
    │   └── routes/
    │       ├── +page.svelte           # Landing
    │       ├── dashboard/             # Dashboard
    │       ├── editor/[module]/       # File editor
    │       └── auth/callback/         # OAuth callback
    ├── package.json
    └── svelte.config.js
```

### Tech Stack

| Layer | Tech |
|-------|------|
| CLI | Go, Cobra, Bubbletea, Lip Gloss |
| CLI Git | go-git |
| CLI Secrets | go-keyring |
| Web | SvelteKit, Tailwind CSS |
| Auth | GitHub OAuth |
| Storage | GitHub repo (user/my-pact) |

---

## Design Principles

1. **GitHub is the database** — No separate backend, your repo is the source of truth
2. **Edit anywhere** — Local editor or web UI, your choice
3. **Cross-OS by default** — Darwin, Windows, Linux configs coexist
4. **Secrets stay local** — API keys in OS keychain, never in repo
5. **Files not strings** — Configs as files, not inline JSON

## License

MIT
