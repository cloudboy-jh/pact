<p align="center">
  <img src="cropped-pact-logo-black.png" alt="Pact Logo" width="200" />
</p>

# Pact

> Your portable dev identity. Shell, editor, AI prefs, themes — one kit, any machine.

Pact stores your entire development environment configuration in a single GitHub repo. Edit in the browser, sync from terminal, apply anywhere.

## The Problem

Setting up a new machine sucks. Dotfiles repos are messy, unstructured, and hard to share. Your AI prompts, editor configs, and shell setup live in different places.

## The Solution

One `pact.json` manifest + organized files in a GitHub repo. Web UI to edit, CLI to sync, cross-OS support built in.

## Quick Start

### CLI

```bash
# Install (from source)
cd cli
go build -o pact .

# Initialize (authenticates with GitHub, creates/clones your pact repo)
pact init

# Sync all configs
pact sync

# Interactive status
pact
```

### Web UI

```bash
cd web
npm install
npm run dev
# Open http://localhost:5173
```

## Commands

| Command | Description |
|---------|-------------|
| `pact` | Interactive status box with quick actions |
| `pact init` | Authenticate with GitHub + clone your pact repo |
| `pact sync` | Pull latest + apply all module configs |
| `pact sync <module>` | Sync a specific module only |
| `pact push` | Commit and push local changes |
| `pact edit` | Open web editor in browser |
| `pact status` | Show status (non-interactive) |
| `pact secret set <name>` | Store a secret in OS keychain |
| `pact secret list` | List secrets and their status |
| `pact secret remove <name>` | Remove a secret from keychain |
| `pact reset` | Remove all symlinks (keeps ~/.pact/) |
| `pact nuke` | Full cleanup (symlinks + ~/.pact/ + token) |

## What's in a Pact

```
username/pact/
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

## pact.json Example

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

## Sync Strategies

| Strategy | Behavior | Use when |
|----------|----------|----------|
| `symlink` | Creates symlink from target to source | Edits in ~/.pact/ reflect immediately |
| `copy` | Copies file to target location | App doesn't follow symlinks |

## Cross-OS Support

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

## Secrets

Secrets are stored in your OS keychain, never in the repo:

| OS | Backend |
|----|---------|
| macOS | Keychain |
| Linux | libsecret / gnome-keyring |
| Windows | Windows Credential Manager |

```bash
pact secret set ANTHROPIC_API_KEY
# Enter value: ****
# Stored in keychain

pact secret list
#   ANTHROPIC_API_KEY    ● set
#   OPENAI_API_KEY       ○ not set
```

## Configuration

### GitHub OAuth App

1. Go to **GitHub Settings → Developer settings → OAuth Apps → New OAuth App**
2. Set:
   - **Application name**: `Pact`
   - **Homepage URL**: `http://localhost:5173`
   - **Authorization callback URL**: `http://localhost:5173/auth/callback`
3. Note your **Client ID** and generate a **Client Secret**

### CLI Configuration

Edit `cli/internal/auth/oauth.go`:

```go
const ClientID = "your-client-id"
```

### Web Configuration

1. Edit `web/src/lib/github.ts`:
   ```ts
   export const GITHUB_CLIENT_ID = 'your-client-id';
   ```

2. Set environment variable:
   ```bash
   export GITHUB_CLIENT_SECRET=your-client-secret
   ```

## Project Structure

```
pactv1/
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

## Tech Stack

| Layer | Tech |
|-------|------|
| CLI | Go, Cobra, Bubbletea, Lip Gloss |
| CLI Git | go-git |
| CLI Secrets | go-keyring |
| Web | SvelteKit, Tailwind CSS |
| Auth | GitHub OAuth |
| Storage | GitHub repo (user/pact) |

## Design Principles

1. **GitHub is the database** — No separate backend, your repo is the source of truth
2. **Web for editing, CLI for syncing** — Each tool does one thing well
3. **Cross-OS by default** — Darwin, Windows, Linux configs coexist
4. **Secrets stay local** — API keys in OS keychain, never in repo
5. **Files not strings** — Configs as files, not inline JSON

## Development

### CLI

```bash
cd cli
go mod tidy
go build -o pact .
./pact --help
```

### Web

```bash
cd web
npm install
npm run dev
```

## License

MIT
