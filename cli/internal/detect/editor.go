package detect

import (
	"os"
)

// Known editors in preference order
var knownEditors = []struct {
	name    string
	command string
}{
	{"zed", "zed"},
	{"cursor", "cursor"},
	{"vscode", "code"},
	{"neovim", "nvim"},
	{"vim", "vim"},
	{"nano", "nano"},
	{"emacs", "emacs"},
	{"sublime", "subl"},
	{"atom", "atom"},
}

// DetectEditor detects installed editors
func DetectEditor() EditorDetected {
	result := EditorDetected{
		Others: []string{},
	}

	// Check $EDITOR and $VISUAL first
	if editor := os.Getenv("EDITOR"); editor != "" {
		result.Default = normalizeEditorName(editor)
	} else if visual := os.Getenv("VISUAL"); visual != "" {
		result.Default = normalizeEditorName(visual)
	}

	// Find all installed editors
	var installed []string
	for _, e := range knownEditors {
		if isToolInstalled(e.command) {
			installed = append(installed, e.name)
		}
	}

	// If no default set, use first installed
	if result.Default == "" && len(installed) > 0 {
		result.Default = installed[0]
		installed = installed[1:]
	} else {
		// Remove default from others list
		var others []string
		for _, e := range installed {
			if e != result.Default {
				others = append(others, e)
			}
		}
		installed = others
	}

	result.Others = installed

	return result
}

// normalizeEditorName converts editor command to name
func normalizeEditorName(cmd string) string {
	switch cmd {
	case "code":
		return "vscode"
	case "nvim":
		return "neovim"
	case "subl":
		return "sublime"
	default:
		return cmd
	}
}
