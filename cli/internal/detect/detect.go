package detect

import (
	"runtime"
)

// DetectedConfig holds everything found on the machine
type DetectedConfig struct {
	CLI         CLIDetected      `json:"cli,omitempty"`
	Shell       ShellDetected    `json:"shell,omitempty"`
	Git         GitDetected      `json:"git,omitempty"`
	Editor      EditorDetected   `json:"editor,omitempty"`
	Terminal    TerminalDetected `json:"terminal,omitempty"`
	LLM         LLMDetected      `json:"llm,omitempty"`
	Secrets     []SecretDetected `json:"secrets,omitempty"`
	ConfigFiles []ConfigFile     `json:"configFiles,omitempty"`
}

// CLIDetected holds detected CLI tools
type CLIDetected struct {
	Tools  []string `json:"tools,omitempty"`
	Custom []string `json:"custom,omitempty"`
}

// ShellDetected holds shell configuration info
type ShellDetected struct {
	Type   string      `json:"type,omitempty"`
	Prompt *PromptInfo `json:"prompt,omitempty"`
	Tools  []string    `json:"tools,omitempty"`
}

// PromptInfo holds prompt tool configuration
type PromptInfo struct {
	Tool   string `json:"tool"`
	Theme  string `json:"theme,omitempty"`
	Source string `json:"source,omitempty"`
}

// GitDetected holds git configuration
type GitDetected struct {
	User          string `json:"user,omitempty"`
	Email         string `json:"email,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
	LFS           bool   `json:"lfs,omitempty"`
}

// EditorDetected holds editor information
type EditorDetected struct {
	Default string   `json:"default,omitempty"`
	Others  []string `json:"others,omitempty"`
	Theme   string   `json:"theme,omitempty"`
	Keymap  string   `json:"keymap,omitempty"`
}

// TerminalDetected holds terminal configuration
type TerminalDetected struct {
	Font     string `json:"font,omitempty"`
	FontSize int    `json:"fontSize,omitempty"`
}

// LLMDetected holds LLM-related configuration
type LLMDetected struct {
	Providers []string  `json:"providers,omitempty"`
	Local     *LocalLLM `json:"local,omitempty"`
	Coding    *Coding   `json:"coding,omitempty"`
}

// LocalLLM holds local LLM runtime info
type LocalLLM struct {
	Runtime string   `json:"runtime,omitempty"`
	Models  []string `json:"models,omitempty"`
}

// Coding holds coding agent info
type Coding struct {
	Agents []string `json:"agents,omitempty"`
}

// SecretDetected holds info about a detected secret
type SecretDetected struct {
	Name       string `json:"name"`
	InEnv      bool   `json:"inEnv"`
	InKeychain bool   `json:"inKeychain"`
	InPactJSON bool   `json:"inPactJson"`
}

// ConfigFile represents a discovered config file
type ConfigFile struct {
	Name       string `json:"name"`
	SourcePath string `json:"sourcePath"`
	DestPath   string `json:"destPath"`
	Module     string `json:"module"`
	Exists     bool   `json:"exists"`
	IsDir      bool   `json:"isDir"`
}

// ScanOptions configures what to scan
type ScanOptions struct {
	Modules      []string // Specific modules to scan (empty = all)
	IncludeFiles bool     // Whether to scan for config files
}

// Scan performs a full environment scan
func Scan(opts ScanOptions) *DetectedConfig {
	detected := &DetectedConfig{}

	modules := opts.Modules
	if len(modules) == 0 {
		modules = []string{"cli", "shell", "git", "editor", "llm", "secrets"}
	}

	moduleSet := make(map[string]bool)
	for _, m := range modules {
		moduleSet[m] = true
	}

	// Always scan config files if no specific modules requested
	if len(opts.Modules) == 0 {
		opts.IncludeFiles = true
	}

	if moduleSet["cli"] {
		detected.CLI = DetectCLITools()
	}

	if moduleSet["shell"] {
		detected.Shell = DetectShell()
	}

	if moduleSet["git"] {
		detected.Git = DetectGit()
	}

	if moduleSet["editor"] {
		detected.Editor = DetectEditor()
	}

	if moduleSet["llm"] {
		detected.LLM = DetectLLM()
	}

	if moduleSet["secrets"] {
		detected.Secrets = DetectSecrets(nil)
	}

	if opts.IncludeFiles {
		allConfigs := DiscoverConfigFiles()
		// Filter config files by requested modules
		if len(opts.Modules) > 0 {
			var filtered []ConfigFile
			for _, cf := range allConfigs {
				if moduleSet[cf.Module] {
					filtered = append(filtered, cf)
				}
			}
			detected.ConfigFiles = filtered
		} else {
			detected.ConfigFiles = allConfigs
		}
	}

	return detected
}

// GetCurrentOS returns the current operating system
func GetCurrentOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return runtime.GOOS
	}
}
