package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudboy-jh/pact/internal/config"
	"github.com/cloudboy-jh/pact/internal/ui"
	"github.com/spf13/cobra"
)

var versionFlag bool

var rootCmd = &cobra.Command{
	Use:   "pact",
	Short: "Your portable dev identity",
	Long:  ui.RenderLogo() + "\nYour portable dev identity. Shell, editor, AI prefs, themes — one kit, any machine.",
	Run: func(cmd *cobra.Command, args []string) {
		// Handle --version flag
		if versionFlag {
			fmt.Println(ui.RenderLogoWithVersion())
			return
		}

		// Check if pact is initialized
		if !config.Exists() {
			fmt.Println(ui.RenderLogo())
			fmt.Println("Pact is not initialized. Run 'pact init' to get started.")
			os.Exit(1)
		}

		// Run interactive TUI
		p := tea.NewProgram(initialModel())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version information")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(secretCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(nukeCmd)
}

// TUI Model
type model struct {
	cfg      *config.PactConfig
	quitting bool
	err      error
}

type keyMap struct {
	Sync key.Binding
	Edit key.Binding
	Quit key.Binding
}

var keys = keyMap{
	Sync: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sync"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func initialModel() model {
	cfg, err := config.Load()
	return model{
		cfg: cfg,
		err: err,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

type syncDoneMsg struct{ err error }
type editDoneMsg struct{ err error }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case syncDoneMsg:
		// Reload config after sync
		cfg, err := config.Load()
		m.cfg = cfg
		m.err = err
		return m, nil
	case editDoneMsg:
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, keys.Sync):
			// Run sync command
			c := exec.Command(os.Args[0], "sync")
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return syncDoneMsg{err}
			})
		case key.Matches(msg, keys.Edit):
			// Run edit command
			c := exec.Command(os.Args[0], "edit")
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return editDoneMsg{err}
			})
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return fmt.Sprintf("Error loading config: %v\n\nPress q to quit.", m.err)
	}

	if m.cfg == nil {
		return "Loading...\n"
	}

	return ui.RenderStatus(m.cfg)
}
