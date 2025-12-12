# Pact Architecture Documentation

> Complete technical reference for Claude and AI consumption

---

## Overview

**Pact** is a portable development environment configuration tool that stores dev configs (shell, editor, AI prefs, themes) in a single GitHub repository.

**Components:**
1. **CLI (Go)** - Terminal interface for syncing, editing, and managing configurations
2. **Web App (SvelteKit)** - Browser-based editor and dashboard
3. **GitHub Repository (`my-pact`)** - The user's configuration storage

**Design Philosophy:**
- GitHub is the database (no separate backend)
- Edit anywhere (local editor or web UI)
- Cross-OS by default (Darwin, Windows, Linux)
- Secrets stay local (OS keychain, never in repo)
- Files not strings (configs as files, not inline JSON)
- **Local-first** - `.pact/` folder lives in your project directory (like `.git/`)

---

## Directory Structure

Pact works like `git` - it creates a `.pact/` folder in your project:

```
my-project/
├── .pact/                 # Cloned from your my-pact repo
│   ├── pact.json
│   ├── shell/
│   ├── editor/
│   └── ...
├── src/
└── package.json
```

**Directory Resolution:**
1. CLI searches for `.pact/` in current directory
2. Walks up the directory tree (like git)
3. Falls back to `~/.pact/` for backwards compatibility

**Token Storage:**
- GitHub token stored globally in OS keychain (not in `.pact/`)
- One authentication works across all projects

---

## CLI (Go)

### Entry Point

**File:** `/cli/main.go`
```go
package main

import "github.com/cloudboy-jh/pact/cmd"

func main() {
    cmd.Execute()
}
```

### Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | Command-line argument parsing |
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/bubbles` | TUI components |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/go-git/go-git/v5` | Git operations |
| `github.com/zalando/go-keyring` | OS keychain access |
| `github.com/pkg/browser` | Opens URLs in browser |
| `golang.org/x/term` | Terminal password input |

---

## CLI Commands

### 1. Root Command (`pact`)

**File:** `/cli/cmd/root.go`

Interactive TUI status display with quick actions.

**Behavior:**
- Checks if pact is initialized (via `config.Exists()`)
- If not initialized, exits with message to run `pact init`
- Runs interactive TUI showing status and quick actions

**Key Bindings:**
| Key | Action |
|-----|--------|
| `s` | Run `pact sync` |
| `e` | Run `pact edit` |
| `q` or `Ctrl+C` | Quit |

---

### 2. Init Command (`pact init`)

**File:** `/cli/cmd/init.go`

Authenticate with GitHub and clone/create the pact repo to `./.pact/` in current directory.

**Flags:**
| Flag | Type | Description |
|------|------|-------------|
| `--from` | string | Fork pact from another user (not yet implemented) |

**Flow:**
1. Check if `.pact/` exists in current directory tree → exit if yes
2. Check for existing token in keychain
3. If token exists and valid → use it
4. If no token → start GitHub Device Flow OAuth
5. Get user info from GitHub API
6. Check if `{username}/my-pact` repo exists
7. If not → create the repo
8. Clone repo to `./.pact/` in current directory
9. If no `pact.json` exists → create default config

**Default Config Template:**
```json
{
  "version": "1.0.0",
  "user": "{username}",
  "modules": {
    "shell": {},
    "editor": {},
    "git": {},
    "ai": {
      "providers": {},
      "prompts": {},
      "agents": {}
    },
    "tools": {
      "configs": {}
    }
  },
  "secrets": []
}
```

---

### 3. Sync Command (`pact sync [module]`)

**File:** `/cli/cmd/sync.go`

Pull latest changes from GitHub and apply configs.

**Arguments:**
| Arg | Description |
|-----|-------------|
| `[module]` | Optional specific module to sync (e.g., `shell`, `editor`) |

**Flow:**
1. Check if pact is initialized
2. Get token from keychain
3. Pull latest changes from GitHub (`git.Pull`)
4. Load `pact.json` configuration
5. If module specified → sync only that module
6. If no module → sync all modules
7. Convert results to UI format and render

---

### 4. Status Command (`pact status`)

**File:** `/cli/cmd/status.go`

Display current status of all modules and secrets (non-interactive).

---

### 5. Push Command (`pact push`)

**File:** `/cli/cmd/push.go`

Commit and push local changes to GitHub.

**Flags:**
| Flag | Type | Description |
|------|------|-------------|
| `-m, --message` | string | Commit message |
| `--force` | bool | Force push (not implemented) |

**Flow:**
1. Check if pact is initialized
2. Get token from keychain
3. Check for uncommitted changes (`git.HasChanges`)
4. If no changes → exit
5. Get commit message (from flag or prompt)
6. Default message: "Update pact configuration"
7. Push changes to GitHub (`git.Push`)

---

### 6. Edit Command (`pact edit [path]`)

**File:** `/cli/cmd/edit.go`

Edit pact files locally or open web editor.

**Arguments:**
| Arg | Description |
|-----|-------------|
| (none) | Opens `~/.pact/pact.json` in `$EDITOR` |
| `web` | Opens web editor in browser |
| `<path>` | Opens specified path relative to `~/.pact/` |

**Environment Variables:**
| Variable | Purpose |
|----------|---------|
| `PACT_WEB_URL` | Override web editor URL (default: `https://pact-ckn.pages.dev`) |
| `EDITOR` | Preferred editor |
| `VISUAL` | Fallback editor |

**Editor Selection Order:**
1. `$EDITOR`
2. `$VISUAL`
3. `vim` (if available)
4. `nano` (if available)
5. `vi` (fallback)

---

### 7. Secret Commands (`pact secret`)

**File:** `/cli/cmd/secret.go`

#### `pact secret set <name>`
- Prompts for secret value (hidden input using `term.ReadPassword`)
- Stores in OS keychain via `keyring.SetSecret`

#### `pact secret list`
- Lists secrets configured in `pact.json`
- Shows status: `● set` or `○ not set`

#### `pact secret remove <name>`
- Removes secret from OS keychain

---

### 8. Reset Command (`pact reset`)

**File:** `/cli/cmd/reset.go`

Remove all symlinks created by pact (keeps `~/.pact/` intact).

---

### 9. Nuke Command (`pact nuke`)

**File:** `/cli/cmd/nuke.go`

Complete removal of pact (symlinks + `~/.pact/` + token).

**Flags:**
| Flag | Type | Description |
|------|------|-------------|
| `-f, --force` | bool | Skip confirmation prompt |

**Actions:**
1. Remove all symlinks (via `sync.RemoveAllSymlinks`)
2. Delete `~/.pact/` directory
3. Remove token from keychain

---

## Internal Packages

### 1. Config Package (`internal/config/pact.go`)

Parse and manage `pact.json` configuration.

#### Data Structures

```go
// Root configuration
type PactConfig struct {
    Version string        `json:"version"`
    User    string        `json:"user"`
    Modules ModulesConfig `json:"modules"`
    Secrets []string      `json:"secrets"`
}

// All module configurations
type ModulesConfig struct {
    Shell       map[string]ModuleEntry `json:"shell,omitempty"`
    Editor      map[string]EditorEntry `json:"editor,omitempty"`
    Terminal    *TerminalEntry         `json:"terminal,omitempty"`
    Git         map[string]ModuleEntry `json:"git,omitempty"`
    AI          *AIConfig              `json:"ai,omitempty"`
    Tools       *ToolsConfig           `json:"tools,omitempty"`
    Keybindings map[string]ModuleEntry `json:"keybindings,omitempty"`
    Snippets    map[string]ModuleEntry `json:"snippets,omitempty"`
    Fonts       *FontsConfig           `json:"fonts,omitempty"`
    Runtimes    *RuntimesConfig        `json:"runtimes,omitempty"`
}

// Basic module entry (source → target mapping)
type ModuleEntry struct {
    Source   string      `json:"source"`
    Target   interface{} `json:"target"` // string or map[string]string for OS-specific
    Strategy string      `json:"strategy,omitempty"` // "symlink" or "copy"
}

// Editor entry
type EditorEntry struct {
    Source   string      `json:"source"`
    Target   interface{} `json:"target"`
    Strategy string      `json:"strategy,omitempty"`
}

// Terminal emulator configuration
type TerminalEntry struct {
    Emulator string      `json:"emulator"`
    Source   string      `json:"source"`
    Target   interface{} `json:"target"`
    Strategy string      `json:"strategy,omitempty"`
}

// AI module configuration
type AIConfig struct {
    Providers map[string]ProviderConfig `json:"providers,omitempty"`
    Prompts   map[string]string         `json:"prompts,omitempty"`
    Skills    string                    `json:"skills,omitempty"`
    Agents    map[string]AgentEntry     `json:"agents,omitempty"`
}

type ProviderConfig struct {
    DefaultModel string   `json:"defaultModel,omitempty"`
    Models       []string `json:"models,omitempty"`
}

type AgentEntry struct {
    Source   string `json:"source"`
    Target   string `json:"target,omitempty"`
    Strategy string `json:"strategy,omitempty"`
}

// Tools configuration
type ToolsConfig struct {
    Configs  map[string]ModuleEntry `json:"configs,omitempty"`
    Packages *PackagesConfig        `json:"packages,omitempty"`
}

type PackagesConfig struct {
    Brew  []string `json:"brew,omitempty"`
    NPM   []string `json:"npm,omitempty"`
    Cargo []string `json:"cargo,omitempty"`
    Go    []string `json:"go,omitempty"`
}

// Fonts configuration
type FontsConfig struct {
    Install []string `json:"install,omitempty"`
}

// Runtime versions
type RuntimesConfig struct {
    Node    string `json:"node,omitempty"`
    Python  string `json:"python,omitempty"`
    Go      string `json:"go,omitempty"`
    Manager string `json:"manager,omitempty"`
}

// Sync item (processed for syncing)
type SyncItem struct {
    Module   string
    Name     string
    Source   string
    Target   string
    Strategy string
    IsDir    bool
}
```

#### Key Functions

| Function | Purpose |
|----------|---------|
| `GetPactDir()` | Finds `.pact/` by walking up directory tree (falls back to `~/.pact`) |
| `GetLocalPactDir()` | Returns `.pact/` in current working directory |
| `FindPactDir()` | Searches for `.pact/` without fallback (returns empty if not found) |
| `GetConfigPath()` | Returns path to `pact.json` |
| `Load()` | Reads and parses `pact.json` |
| `Exists()` | Checks if `pact.json` exists |
| `GetCurrentOS()` | Returns OS name (`darwin`, `linux`, `windows`) |
| `ExpandPath(path)` | Expands `~` to home directory |
| `ResolveTarget(target)` | Resolves OS-specific target paths |
| `GetSyncItems()` | Returns all items to sync for current OS |
| `CountModuleFiles(module)` | Counts files in a module |

#### Target Resolution

Targets can be:
1. **String:** Direct path (e.g., `"~/.zshrc"`)
2. **Map:** OS-specific paths:
```json
{
  "target": {
    "darwin": "~/.config/nvim",
    "linux": "~/.config/nvim",
    "windows": "~/AppData/Local/nvim"
  }
}
```

---

### 2. Auth Package (`internal/auth/oauth.go`)

GitHub OAuth device flow authentication.

#### Constants
```go
const (
    defaultClientID = "Ov23liB8Z30c0BkX2nXF"
    deviceCodeURL   = "https://github.com/login/device/code"
    tokenURL        = "https://github.com/login/oauth/access_token"
    scopes          = "repo"
)
```

#### Environment Variables
| Variable | Purpose |
|----------|---------|
| `GITHUB_CLIENT_ID` | Override default OAuth client ID |

#### Data Structures
```go
type DeviceCodeResponse struct {
    DeviceCode      string `json:"device_code"`
    UserCode        string `json:"user_code"`
    VerificationURI string `json:"verification_uri"`
    ExpiresIn       int    `json:"expires_in"`
    Interval        int    `json:"interval"`
}

type TokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    Scope       string `json:"scope"`
    Error       string `json:"error,omitempty"`
}

type GitHubUser struct {
    Login     string `json:"login"`
    ID        int64  `json:"id"`
    AvatarURL string `json:"avatar_url"`
    Name      string `json:"name"`
}
```

#### Key Functions

| Function | Purpose |
|----------|---------|
| `GetClientID()` | Returns client ID (env or default) |
| `RequestDeviceCode()` | Initiates device flow, returns user code |
| `PollForToken(deviceCode, interval)` | Polls for access token |
| `GetUser(token)` | Fetches authenticated user info |
| `RepoExists(token, username)` | Checks if `{username}/my-pact` exists |
| `CreateRepo(token)` | Creates `my-pact` repository |

#### Device Flow Process
1. Request device code from GitHub
2. User visits `https://github.com/login/device`
3. User enters the user code
4. CLI polls for token (respects `interval`)
5. Handle responses: `authorization_pending`, `slow_down`, `expired_token`, `access_denied`

---

### 3. Git Package (`internal/git/git.go`)

Git operations using go-git library.

#### Key Functions

| Function | Purpose |
|----------|---------|
| `Clone(token, username, targetDir)` | Clones `{username}/my-pact` to specified directory |
| `Pull(token, pactDir)` | Pulls latest changes |
| `Push(token, pactDir, message)` | Stages, commits, and pushes changes |
| `HasChanges(pactDir)` | Checks if working directory has changes |
| `GetStatus(pactDir)` | Returns git status string |

#### Authentication
Uses HTTP Basic Auth with token:
```go
Auth: &http.BasicAuth{
    Username: "x-access-token",
    Password: token,
}
```

#### Commit Author
- Uses git config if available
- Fallback: `pact <pact@users.noreply.github.com>`

---

### 4. Keyring Package (`internal/keyring/keyring.go`)

Secure storage in OS keychain.

**Service Name:** `pact`

#### Functions

| Function | Purpose |
|----------|---------|
| `SetToken(token)` | Stores GitHub token (key: `github_token`) |
| `GetToken()` | Retrieves GitHub token |
| `DeleteToken()` | Removes GitHub token |
| `HasToken()` | Checks if token exists |
| `SetSecret(name, value)` | Stores a named secret |
| `GetSecret(name)` | Retrieves a named secret |
| `DeleteSecret(name)` | Removes a named secret |
| `HasSecret(name)` | Checks if secret exists |

#### OS Backends
| OS | Backend |
|----|---------|
| macOS | Keychain |
| Linux | libsecret / gnome-keyring |
| Windows | Windows Credential Manager |

---

### 5. Sync Package (`internal/sync/sync.go`)

Apply configurations via symlinks or copies.

#### Data Structures
```go
type Result struct {
    Module  string
    Name    string
    Success bool
    Error   error
    Skipped bool
    Message string
}
```

#### Key Functions

| Function | Purpose |
|----------|---------|
| `SyncAll(cfg)` | Syncs all items from config |
| `SyncModule(cfg, module)` | Syncs only a specific module |
| `RemoveAllSymlinks(cfg)` | Removes all symlinks (for reset/nuke) |

#### Sync Strategies

| Strategy | Behavior |
|----------|----------|
| `symlink` (default) | Creates symbolic link from target to source |
| `copy` | Copies file/directory to target location |

#### Sync Process
1. Check if source exists
2. Determine strategy (default: `symlink`)
3. Create target parent directory if needed
4. Remove existing target (file or directory)
5. Apply strategy:
   - **symlink:** `os.Symlink(absSource, target)`
   - **copy:** Recursively copy files, preserving permissions

---

### 6. UI Package (`internal/ui/status.go`)

TUI rendering using Lip Gloss.

#### Colors
```go
emerald = lipgloss.Color("#34d399")  // Success
amber   = lipgloss.Color("#fbbf24")  // Warning
red     = lipgloss.Color("#f87171")  // Error
zinc400 = lipgloss.Color("#a1a1aa")
zinc500 = lipgloss.Color("#71717a")
zinc600 = lipgloss.Color("#52525b")
zinc800 = lipgloss.Color("#27272a")
zinc900 = lipgloss.Color("#18181b")
```

#### Key Functions

| Function | Purpose |
|----------|---------|
| `GetModuleStatuses(cfg)` | Gets status of all modules |
| `RenderStatus(cfg)` | Renders the status box |
| `RenderSyncResults(results)` | Renders sync operation results |

---

## pact.json Schema

### Full Schema Example

```json
{
  "version": "1.0.0",
  "user": "username",
  "modules": {
    "shell": {
      "darwin": {
        "source": "./shell/darwin.zshrc",
        "target": "~/.zshrc",
        "strategy": "symlink"
      },
      "linux": {
        "source": "./shell/linux.zshrc",
        "target": "~/.zshrc"
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
        }
      },
      "vscode": {
        "source": "./editor/vscode/settings.json",
        "target": "~/.config/Code/User/settings.json"
      }
    },
    "terminal": {
      "emulator": "ghostty",
      "source": "./terminal/ghostty.conf",
      "target": "~/.config/ghostty/config"
    },
    "git": {
      "config": {
        "source": "./git/.gitconfig",
        "target": "~/.gitconfig"
      },
      "ignore": {
        "source": "./git/.gitignore_global",
        "target": "~/.gitignore_global"
      }
    },
    "ai": {
      "providers": {
        "anthropic": {
          "defaultModel": "claude-3-5-sonnet-20241022",
          "models": ["claude-3-5-sonnet-20241022", "claude-3-haiku-20240307"]
        }
      },
      "prompts": {
        "default": "prompts/default.md",
        "code-review": "prompts/code-review.md"
      },
      "skills": "skills/",
      "agents": {
        "claude": {
          "source": "./agents/CLAUDE.md",
          "target": "~/.config/claude/CLAUDE.md"
        },
        "cursor": {
          "source": "./agents/.cursorrules"
        }
      }
    },
    "tools": {
      "configs": {
        "lazygit": {
          "source": "./tools/lazygit.yml",
          "target": "~/.config/lazygit/config.yml"
        }
      },
      "packages": {
        "brew": ["ripgrep", "fzf", "bat"],
        "npm": ["typescript", "eslint"],
        "cargo": ["exa"],
        "go": ["github.com/jesseduffield/lazygit"]
      }
    },
    "keybindings": {
      "vscode": {
        "source": "./keybindings/vscode.json",
        "target": "~/.config/Code/User/keybindings.json"
      }
    },
    "snippets": {
      "vscode": {
        "source": "./snippets/vscode/",
        "target": "~/.config/Code/User/snippets/"
      }
    },
    "fonts": {
      "install": ["JetBrains Mono", "Fira Code"]
    },
    "runtimes": {
      "node": "20.10.0",
      "python": "3.12.0",
      "go": "1.22.0",
      "manager": "asdf"
    }
  },
  "secrets": [
    "ANTHROPIC_API_KEY",
    "OPENAI_API_KEY",
    "GITHUB_TOKEN"
  ]
}
```

### Module Types

| Module | Key Type | Description |
|--------|----------|-------------|
| `shell` | OS name (`darwin`, `linux`, `windows`) | Shell configs per OS |
| `editor` | Editor name (freeform) | Editor configs (nvim, vscode, etc.) |
| `terminal` | Single object | Terminal emulator config |
| `git` | Config name (freeform) | Git-related configs |
| `ai` | Nested structure | AI providers, prompts, agents |
| `tools` | `configs` + `packages` | Tool configs and package lists |
| `keybindings` | App name (freeform) | Keybinding files |
| `snippets` | App name (freeform) | Snippet directories |
| `fonts` | `install` array | Fonts to install |
| `runtimes` | Runtime name | Version manager configs |

---

## Web Application (SvelteKit)

### Tech Stack
- **Framework:** SvelteKit
- **Styling:** Tailwind CSS
- **Code Editor:** CodeMirror 6
- **Icons:** Lucide Svelte
- **Deployment:** Cloudflare Pages

### Routes

#### 1. Landing Page (`/`)

**File:** `/web/src/routes/+page.svelte`

Marketing page with GitHub sign-in. Hero section, features grid, CLI preview.

---

#### 2. Auth Server Route (`/auth`)

**File:** `/web/src/routes/auth/+server.ts`

Redirects to GitHub OAuth authorization with `client_id` and `scope: "repo"`.

---

#### 3. Auth Callback Page (`/auth/callback`)

**File:** `/web/src/routes/auth/callback/+page.svelte`

Handles OAuth callback:
1. Get `code` from URL params
2. Exchange code for token via `/api/auth/callback`
3. Get user info from GitHub API
4. Store token and user in auth store
5. Redirect to `/dashboard` or `/setup`

---

#### 4. Auth Callback API (`/api/auth/callback`)

**File:** `/web/src/routes/api/auth/callback/+server.ts`

Server-side token exchange (keeps client secret secure).

---

#### 5. Setup Page (`/setup`)

**File:** `/web/src/routes/setup/+page.svelte`

First-time setup for new users. Creates `my-pact` repo with default `pact.json`.

---

#### 6. Dashboard (`/dashboard`)

**File:** `/web/src/routes/dashboard/+page.svelte`

Main dashboard showing all modules and their status. Module list with status indicators, secrets sidebar, quick commands reference.

---

#### 7. Editor Page (`/editor`)

**File:** `/web/src/routes/editor/+page.svelte`

Full-featured file editor for `pact.json` and other files.

**Features:**
- Left sidebar with file tree
- CodeMirror editor with JSON syntax highlighting
- Section highlighting (from URL param `?section=...`)
- Auto-save with 1.5s debounce
- "Push to GitHub" button
- "Open in AI" dropdown (Claude, ChatGPT, Gemini, Grok)

---

### Components

#### CodeEditor (`/web/src/lib/components/CodeEditor.svelte`)

CodeMirror 6-based code editor.

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `content` | string | `''` | Editor content |
| `language` | `'json' \| 'text'` | `'json'` | Syntax highlighting mode |
| `readonly` | boolean | `false` | Read-only mode |
| `highlightLines` | `{from, to} \| null` | `null` | Lines to highlight |

**Events:**
| Event | Payload | Description |
|-------|---------|-------------|
| `change` | string | Content changed |
| `click` | void | Editor clicked |

---

### Stores

#### Auth Store (`/web/src/lib/stores/auth.ts`)

Manages authentication state.

```typescript
interface AuthState {
    token: string | null;
    user: User | null;
    loading: boolean;
}
```

**Storage:** `localStorage` (`github_token`, `github_user`)

**Methods:**
| Method | Purpose |
|--------|---------|
| `setToken(token)` | Stores token |
| `setUser(user)` | Stores user info |
| `logout()` | Clears all auth data |
| `initialize()` | Loads from storage, validates token |

---

### GitHub Client (`/web/src/lib/github.ts`)

GitHub API wrapper.

**Methods:**
| Method | Purpose |
|--------|---------|
| `repoExists(username)` | Checks if `my-pact` repo exists |
| `createRepo()` | Creates `my-pact` repo |
| `getContents(username, path)` | Gets directory contents |
| `getFileContent(username, path)` | Gets file content (decoded) |
| `updateFile(username, path, content, message, sha?)` | Creates/updates file |
| `deleteFile(username, path, sha, message)` | Deletes file |
| `getPactConfig(username)` | Gets parsed `pact.json` |
| `savePactConfig(username, config, sha?)` | Saves `pact.json` |

---

## Environment Variables

### CLI

| Variable | Purpose | Default |
|----------|---------|---------|
| `GITHUB_CLIENT_ID` | OAuth app client ID | `Ov23liB8Z30c0BkX2nXF` |
| `PACT_WEB_URL` | Web editor URL | `https://pact-ckn.pages.dev` |
| `EDITOR` | Preferred text editor | (varies) |
| `VISUAL` | Fallback text editor | (varies) |

### Web

| Variable | Purpose | Where Set |
|----------|---------|-----------|
| `GITHUB_CLIENT_ID` | OAuth app client ID | Code or env |
| `GITHUB_CLIENT_SECRET` | OAuth app secret | Cloudflare dashboard |

---

## Deployment

### CLI (GoReleaser)

**File:** `/cli/.goreleaser.yaml`

**Platforms:** Linux, macOS, Windows (amd64, arm64)

**Distribution:**
- GitHub Releases (tar.gz, zip for Windows)
- Homebrew via `cloudboy-jh/homebrew-tap`

### Web (Cloudflare Pages)

**File:** `/web/wrangler.toml`

**Deployment URL:** `https://pact-ckn.pages.dev`

### GitHub Actions

**File:** `/.github/workflows/release.yml`

**Trigger:** Push tag matching `v*`

**Secrets Required:**
- `GITHUB_TOKEN` (automatic)
- `HOMEBREW_TAP_GITHUB_TOKEN` (for Homebrew formula)

---

## Data Flow

### Authentication Flow (CLI)

```
1. User runs `pact init` in their project directory
2. CLI checks OS keychain for existing token
3. If no token: start GitHub Device Flow OAuth
   - CLI requests device code from GitHub
   - User visits github.com/login/device, enters code
   - CLI polls for token
4. Token stored globally in OS keychain (works across all projects)
5. Clone my-pact repo to ./.pact/ in current directory
6. Token used for all GitHub operations
```

### Authentication Flow (Web)

```
1. User clicks "Sign in with GitHub"
2. Redirect to GitHub OAuth authorize
3. GitHub redirects to /auth/callback with code
4. Server exchanges code for token
5. Token stored in localStorage
6. Token used for all GitHub API calls
```

### Sync Flow

```
1. User runs `pact sync` (from anywhere in project tree)
2. CLI finds .pact/ by walking up directory tree
3. CLI pulls latest from GitHub (git pull)
4. CLI loads pact.json
5. CLI gets sync items for current OS
6. For each item:
   a. Check if source exists
   b. Create target directory if needed
   c. Remove existing target
   d. Apply strategy (symlink or copy)
7. Report results
```

### Edit Flow (Web)

```
1. User opens web editor
2. Web fetches pact.json and repo contents from GitHub
3. User edits in CodeMirror
4. Changes auto-save (debounced 1.5s)
5. Web commits directly to GitHub
6. User runs `pact sync` on machine to pull changes
```

---

## Repository Structure (User's my-pact repo)

```
username/my-pact/
├── pact.json              # Manifest file
├── shell/                 # Shell configurations
│   ├── darwin.zshrc
│   ├── linux.zshrc
│   └── windows.ps1
├── editor/                # Editor configurations
│   ├── nvim/
│   ├── vscode/
│   └── cursor/
├── terminal/              # Terminal emulator configs
├── git/                   # Git configurations
│   ├── .gitconfig
│   └── .gitignore_global
├── prompts/               # AI prompts
├── skills/                # Custom AI skills
├── agents/                # Agent configs (CLAUDE.md, .cursorrules)
├── tools/                 # Tool configurations
├── keybindings/           # Editor keybindings
├── snippets/              # Code snippets
└── fonts/                 # Font preferences
```

---

## Project Structure (This repo)

```
pact/
├── cli/                    # Go CLI
│   ├── cmd/                # Cobra commands
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── sync.go
│   │   ├── push.go
│   │   ├── edit.go
│   │   ├── status.go
│   │   ├── secret.go
│   │   ├── reset.go
│   │   └── nuke.go
│   ├── internal/
│   │   ├── auth/           # GitHub OAuth device flow
│   │   ├── config/         # pact.json parsing
│   │   ├── git/            # Git operations
│   │   ├── keyring/        # OS keychain
│   │   ├── sync/           # Symlink/copy logic
│   │   └── ui/             # TUI (Lip Gloss + Bubbletea)
│   ├── .goreleaser.yaml
│   ├── go.mod
│   └── main.go
│
├── web/                    # SvelteKit web app
│   ├── src/
│   │   ├── lib/
│   │   │   ├── github.ts   # GitHub API client
│   │   │   ├── stores/     # Auth state
│   │   │   └── components/ # CodeEditor
│   │   └── routes/
│   │       ├── +page.svelte           # Landing
│   │       ├── dashboard/             # Dashboard
│   │       ├── editor/                # File editor
│   │       ├── setup/                 # First-time setup
│   │       ├── auth/                  # OAuth redirect
│   │       └── api/auth/callback/     # Token exchange
│   ├── wrangler.toml
│   └── package.json
│
├── docs/
│   ├── ARCHITECTURE.md     # This file
│   └── ROADMAP.md
│
├── .github/
│   └── workflows/
│       └── release.yml     # GoReleaser on tag
│
└── README.md
```
