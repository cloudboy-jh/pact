package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/keyring"
)

var (
	// Colors
	emerald = lipgloss.Color("#34d399")
	amber   = lipgloss.Color("#fbbf24")
	red     = lipgloss.Color("#f87171")
	zinc400 = lipgloss.Color("#a1a1aa")
	zinc500 = lipgloss.Color("#71717a")
	zinc600 = lipgloss.Color("#52525b")
	zinc800 = lipgloss.Color("#27272a")
	zinc900 = lipgloss.Color("#18181b")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(zinc500)

	successStyle = lipgloss.NewStyle().
			Foreground(emerald)

	warningStyle = lipgloss.NewStyle().
			Foreground(amber)

	errorStyle = lipgloss.NewStyle().
			Foreground(red)

	dimStyle = lipgloss.NewStyle().
			Foreground(zinc600)

	moduleNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Width(14)

	statusTextStyle = lipgloss.NewStyle().
			Width(20)

	fileCountStyle = lipgloss.NewStyle().
			Foreground(zinc500)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(zinc800).
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(zinc500).
			Padding(0, 2)
)

// ModuleStatus represents the status of a module
type ModuleStatus struct {
	Name      string
	Status    string // "synced", "pending", "error", "not_configured"
	FileCount int
	Error     error
}

// GetModuleStatuses returns the status of all modules
func GetModuleStatuses(cfg *config.PactConfig) []ModuleStatus {
	modules := []string{"shell", "editor", "terminal", "git", "ai", "tools", "keybindings", "snippets", "fonts"}
	var statuses []ModuleStatus

	for _, module := range modules {
		status := ModuleStatus{
			Name:      module,
			FileCount: cfg.CountModuleFiles(module),
		}

		// Check if module is configured
		switch module {
		case "shell":
			if cfg.Modules.Shell == nil || len(cfg.Modules.Shell) == 0 {
				status.Status = "not_configured"
			} else if _, ok := cfg.Modules.Shell[config.GetCurrentOS()]; !ok {
				status.Status = "not_configured"
			} else {
				status.Status = "synced" // TODO: actually check sync status
			}
		case "editor":
			if cfg.Modules.Editor == nil || len(cfg.Modules.Editor) == 0 {
				status.Status = "not_configured"
			} else {
				status.Status = "synced"
			}
		case "terminal":
			if cfg.Modules.Terminal == nil {
				status.Status = "not_configured"
			} else {
				status.Status = "synced"
			}
		case "git":
			if cfg.Modules.Git == nil || len(cfg.Modules.Git) == 0 {
				status.Status = "not_configured"
			} else {
				status.Status = "synced"
			}
		case "ai":
			if cfg.Modules.AI == nil {
				status.Status = "not_configured"
			} else {
				status.Status = "synced"
			}
		case "tools":
			if cfg.Modules.Tools == nil {
				status.Status = "not_configured"
			} else {
				status.Status = "synced"
			}
		case "keybindings":
			if cfg.Modules.Keybindings == nil || len(cfg.Modules.Keybindings) == 0 {
				status.Status = "not_configured"
			} else {
				status.Status = "synced"
			}
		case "snippets":
			if cfg.Modules.Snippets == nil || len(cfg.Modules.Snippets) == 0 {
				status.Status = "not_configured"
			} else {
				status.Status = "synced"
			}
		case "fonts":
			if cfg.Modules.Fonts == nil || len(cfg.Modules.Fonts.Install) == 0 {
				status.Status = "not_configured"
			} else {
				status.Status = "synced"
			}
		}

		statuses = append(statuses, status)
	}

	return statuses
}

// RenderStatus renders the status box
func RenderStatus(cfg *config.PactConfig) string {
	var sb strings.Builder

	// Header
	hostname, _ := os.Hostname()
	header := fmt.Sprintf("%s%s%s",
		titleStyle.Render("pact"),
		strings.Repeat(" ", 30),
		subtitleStyle.Render(hostname),
	)
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Modules
	statuses := GetModuleStatuses(cfg)
	for _, status := range statuses {
		line := renderModuleLine(status)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Secrets
	sb.WriteString("\n")
	secretsLine := renderSecretsLine(cfg)
	sb.WriteString(secretsLine)

	content := sb.String()
	box := boxStyle.Render(content)

	// Help line
	help := helpStyle.Render("[s] sync  [e] edit (web)  [q] quit")

	return box + "\n" + help
}

func renderModuleLine(status ModuleStatus) string {
	name := moduleNameStyle.Render(status.Name)
	dashes := dimStyle.Render(strings.Repeat("─", 2))

	var statusIcon, statusText string
	switch status.Status {
	case "synced":
		statusIcon = successStyle.Render("✓")
		statusText = successStyle.Render("synced")
	case "pending":
		statusIcon = warningStyle.Render("⚠")
		statusText = warningStyle.Render("pending")
	case "error":
		statusIcon = errorStyle.Render("✗")
		statusText = errorStyle.Render("error")
	case "not_configured":
		statusIcon = dimStyle.Render(" ")
		statusText = dimStyle.Render("not configured")
	}

	statusPart := statusTextStyle.Render(fmt.Sprintf("%s %s", statusIcon, statusText))

	var filesPart string
	if status.FileCount > 0 {
		unit := "files"
		if status.FileCount == 1 {
			unit = "file"
		}
		filesPart = fileCountStyle.Render(fmt.Sprintf("%d %s", status.FileCount, unit))
	}

	return fmt.Sprintf("%s %s %s  %s", name, dashes, statusPart, filesPart)
}

func renderSecretsLine(cfg *config.PactConfig) string {
	if len(cfg.Secrets) == 0 {
		return dimStyle.Render("secrets ──────── none configured")
	}

	setCount := 0
	for _, secret := range cfg.Secrets {
		if keyring.HasSecret(secret) {
			setCount++
		}
	}

	name := moduleNameStyle.Render("secrets")
	dashes := dimStyle.Render(strings.Repeat("─", 2))

	var statusPart string
	if setCount == len(cfg.Secrets) {
		statusPart = successStyle.Render(fmt.Sprintf("%d/%d set", setCount, len(cfg.Secrets)))
	} else {
		statusPart = warningStyle.Render(fmt.Sprintf("%d/%d set", setCount, len(cfg.Secrets)))
	}

	return fmt.Sprintf("%s %s %s", name, dashes, statusPart)
}

// RenderSyncResults renders the results of a sync operation
func RenderSyncResults(results []SyncResult) string {
	var sb strings.Builder

	successCount := 0
	failCount := 0
	skipCount := 0

	for _, result := range results {
		var icon, status string
		if result.Success {
			icon = successStyle.Render("✓")
			status = successStyle.Render("synced")
			successCount++
		} else if result.Skipped {
			icon = dimStyle.Render("○")
			status = dimStyle.Render("skipped")
			skipCount++
		} else {
			icon = errorStyle.Render("✗")
			status = errorStyle.Render("error")
			failCount++
		}

		name := fmt.Sprintf("%s/%s", result.Module, result.Name)
		sb.WriteString(fmt.Sprintf("  %s %s %s\n", icon, moduleNameStyle.Render(name), status))

		if result.Error != nil {
			sb.WriteString(fmt.Sprintf("    %s\n", dimStyle.Render(result.Error.Error())))
		}
	}

	sb.WriteString("\n")
	summary := fmt.Sprintf("  %s synced, %s failed, %s skipped",
		successStyle.Render(fmt.Sprintf("%d", successCount)),
		errorStyle.Render(fmt.Sprintf("%d", failCount)),
		dimStyle.Render(fmt.Sprintf("%d", skipCount)),
	)
	sb.WriteString(summary)

	return sb.String()
}

// SyncResult is a simplified result for UI rendering
type SyncResult struct {
	Module  string
	Name    string
	Success bool
	Skipped bool
	Error   error
}
