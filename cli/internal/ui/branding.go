package ui

import "github.com/charmbracelet/lipgloss"

// ASCII logo for Pact CLI
const Logo = `██████   ████   ████  ████████
██   ██ ██  ██ ██        ██
██████  ██████ ██        ██
██      ██  ██ ██        ██
██      ██  ██  ████     ██`

// Tagline displayed under the logo
const Tagline = "Cross-platform environment manager"

// Version can be set at build time via ldflags
var Version = "dev"

// Brand color style for the logo
var logoStyle = lipgloss.NewStyle().
	Foreground(emerald)

// Tagline style (muted)
var taglineStyle = lipgloss.NewStyle().
	Foreground(zinc500)

// RenderLogo returns the styled logo with tagline
func RenderLogo() string {
	return logoStyle.Render(Logo) + "\n\n" + taglineStyle.Render(Tagline) + "\n"
}

// RenderLogoWithVersion returns the styled logo with version
func RenderLogoWithVersion() string {
	versionText := taglineStyle.Render("v" + Version)
	return logoStyle.Render(Logo) + "\n\n" + taglineStyle.Render(Tagline) + "  " + versionText + "\n"
}
