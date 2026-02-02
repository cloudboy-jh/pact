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
	zinc500 = lipgloss.Color("#71717a")
	zinc600 = lipgloss.Color("#52525b")
	zinc800 = lipgloss.Color("#27272a")

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
	Status    string // "configured", "has_files", "not_configured"
	FileCount int
	Details   string
}

// GetModuleStatuses returns the status of all modules found in config
func GetModuleStatuses(cfg *config.PactConfig) []ModuleStatus {
	var statuses []ModuleStatus

	// Get all modules from config (top-level objects)
	modules := cfg.GetModules()

	for _, module := range modules {
		status := ModuleStatus{
			Name:      module,
			FileCount: cfg.CountModuleFiles(module),
		}

		// Check if module has any files configured
		if status.FileCount > 0 {
			status.Status = "has_files"
		} else if cfg.HasKey(module) {
			status.Status = "configured"
		} else {
			status.Status = "not_configured"
		}

		// Get some details about the module
		status.Details = getModuleDetails(cfg, module)

		statuses = append(statuses, status)
	}

	return statuses
}

// getModuleDetails extracts useful info about a module
func getModuleDetails(cfg *config.PactConfig, module string) string {
	var details []string

	switch module {
	case "shell":
		if tool := cfg.GetString("shell.prompt.tool"); tool != "" {
			details = append(details, tool)
		}
		if tools := cfg.GetStringSlice("shell.tools"); len(tools) > 0 {
			details = append(details, tools...)
		}
	case "editor":
		if def := cfg.GetString("editor.default"); def != "" {
			details = append(details, def)
		}
	case "terminal":
		if font := cfg.GetString("terminal.font"); font != "" {
			details = append(details, font)
		}
	case "git":
		if user := cfg.GetString("git.user"); user != "" {
			details = append(details, user)
		}
	case "llm":
		if providers := cfg.GetStringSlice("llm.providers"); len(providers) > 0 {
			details = append(details, providers...)
		}
	case "cli":
		if tools := cfg.GetStringSlice("cli.tools"); len(tools) > 0 {
			if len(tools) > 3 {
				details = append(details, tools[:3]...)
				details = append(details, "...")
			} else {
				details = append(details, tools...)
			}
		}
	}

	if len(details) > 0 {
		return strings.Join(details, ", ")
	}
	return ""
}

func getReservedLines(hasSecrets bool) int {
	// Reserve lines for: header(2) + box borders(2) + help(1) + secrets(2 if present)
	reserved := 2 + 2 + 1
	if hasSecrets {
		reserved += 2
	}
	return reserved
}

func getAvailableHeight(termHeight int, hasSecrets bool) int {
	return termHeight - getReservedLines(hasSecrets)
}

func getMaxScrollForAvailable(totalLines int, available int) int {
	if totalLines <= 0 || available <= 0 {
		return 0
	}
	if totalLines <= available {
		return 0
	}
	// When scrolling, one line is always used by a scroll indicator at top or bottom
	maxVisibleAtEdge := available - 1
	if maxVisibleAtEdge < 0 {
		maxVisibleAtEdge = 0
	}
	maxScroll := totalLines - maxVisibleAtEdge
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func getMaxVisible(totalLines int, scrollOffset int, available int) int {
	if available <= 0 {
		return 0
	}
	maxVisible := available
	// Reserve line for "above" indicator if we're not at the top
	if scrollOffset > 0 {
		maxVisible--
	}
	// Reserve line for "below" indicator if content remains
	if scrollOffset+maxVisible < totalLines {
		maxVisible--
	}
	if maxVisible < 0 {
		return 0
	}
	return maxVisible
}

// GetMaxScroll calculates the maximum scroll offset based on content and terminal height
func GetMaxScroll(cfg *config.PactConfig, termHeight int) int {
	statuses := GetModuleStatuses(cfg)
	secrets := cfg.GetSecrets()

	// Handle edge cases
	if termHeight <= 0 {
		return 0
	}

	availableHeight := getAvailableHeight(termHeight, len(secrets) > 0)
	return getMaxScrollForAvailable(len(statuses), availableHeight)
}

// RenderStatus renders the status box with optional scrolling
// scrollOffset: how many lines to skip from the top of the module list
// termHeight: terminal height for pagination (0 = no pagination)
func RenderStatus(cfg *config.PactConfig, scrollOffset int, termHeight int) string {
	var sb strings.Builder
	secrets := cfg.GetSecrets()
	hasSecrets := len(secrets) > 0

	// Header
	name := cfg.GetString("name")
	if name == "" {
		name = "pact"
	}
	hostname, _ := os.Hostname()
	header := fmt.Sprintf("%s%s%s",
		titleStyle.Render(name),
		strings.Repeat(" ", 30-len(name)),
		subtitleStyle.Render(hostname),
	)
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Modules
	statuses := GetModuleStatuses(cfg)
	if len(statuses) == 0 {
		sb.WriteString(dimStyle.Render("No modules configured"))
		sb.WriteString("\n")
	} else {
		availableHeight := getAvailableHeight(termHeight, hasSecrets)
		if termHeight == 0 || availableHeight <= 0 || availableHeight >= len(statuses) {
			// No pagination needed - show all
			for _, status := range statuses {
				line := renderModuleLine(status)
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		} else {
			// Pagination active
			// Validate scrollOffset bounds
			if scrollOffset < 0 {
				scrollOffset = 0
			}
			maxScroll := getMaxScrollForAvailable(len(statuses), availableHeight)
			if scrollOffset > maxScroll {
				scrollOffset = maxScroll
			}

			maxVisible := getMaxVisible(len(statuses), scrollOffset, availableHeight)

			endIndex := scrollOffset + maxVisible
			if endIndex > len(statuses) {
				endIndex = len(statuses)
			}

			// Show scroll up indicator if not at top
			if scrollOffset > 0 {
				sb.WriteString(dimStyle.Render(fmt.Sprintf("  ... %d more above (k to scroll)", scrollOffset)))
				sb.WriteString("\n")
			}

			// Render visible modules
			for i := scrollOffset; i < endIndex; i++ {
				line := renderModuleLine(statuses[i])
				sb.WriteString(line)
				sb.WriteString("\n")
			}

			// Show scroll down indicator if not at bottom
			remaining := len(statuses) - endIndex
			if remaining > 0 {
				sb.WriteString(dimStyle.Render(fmt.Sprintf("  ... %d more below (j to scroll)", remaining)))
				sb.WriteString("\n")
			}
		}
	}

	// Secrets
	if hasSecrets {
		sb.WriteString("\n")
		secretsLine := renderSecretsLine(secrets)
		sb.WriteString(secretsLine)
	}

	content := sb.String()
	box := boxStyle.Render(content)

	// Help line (updated with scroll hint)
	help := helpStyle.Render("[s] sync  [e] edit  [r] refresh  [j/k] scroll  [q] quit")

	return box + "\n" + help
}

func renderModuleLine(status ModuleStatus) string {
	name := moduleNameStyle.Render(status.Name)
	dashes := dimStyle.Render(strings.Repeat("─", 2))

	var statusIcon, statusText string
	switch status.Status {
	case "has_files":
		statusIcon = successStyle.Render("✓")
		statusText = successStyle.Render("ready")
	case "configured":
		statusIcon = successStyle.Render("●")
		statusText = dimStyle.Render("config only")
	case "not_configured":
		statusIcon = dimStyle.Render(" ")
		statusText = dimStyle.Render("not configured")
	}

	statusPart := statusTextStyle.Render(fmt.Sprintf("%s %s", statusIcon, statusText))

	var extra string
	if status.Details != "" {
		extra = fileCountStyle.Render(status.Details)
	} else if status.FileCount > 0 {
		unit := "files"
		if status.FileCount == 1 {
			unit = "file"
		}
		extra = fileCountStyle.Render(fmt.Sprintf("%d %s", status.FileCount, unit))
	}

	return fmt.Sprintf("%s %s %s  %s", name, dashes, statusPart, extra)
}

func renderSecretsLine(secrets []string) string {
	if len(secrets) == 0 {
		return dimStyle.Render("secrets ──────── none configured")
	}

	setCount := 0
	for _, secret := range secrets {
		if keyring.HasSecret(secret) {
			setCount++
		}
	}

	name := moduleNameStyle.Render("secrets")
	dashes := dimStyle.Render(strings.Repeat("─", 2))

	var statusPart string
	if setCount == len(secrets) {
		statusPart = successStyle.Render(fmt.Sprintf("%d/%d set", setCount, len(secrets)))
	} else {
		statusPart = warningStyle.Render(fmt.Sprintf("%d/%d set", setCount, len(secrets)))
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
