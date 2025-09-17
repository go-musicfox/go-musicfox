package commands

import (
	"github.com/gookit/gcli/v2"

	"github.com/charmbracelet/bubbles/table"

	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/clipboard"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func NewConfigCommand() *gcli.Command {
	cmd := &gcli.Command{
		Name:   "config",
		UseFor: "Print configuration information",
		Func: func(_ *gcli.Command, _ []string) error {
			columns := []table.Column{
				{Title: "Item", Width: 30},
				{Title: "Value", Width: 60},
			}

			rows := []table.Row{
				{"Config Dir", app.ConfigDir()},
				{"Data Dir", app.DataDir()},
				{"State Dir", app.StateDir()},
				{"Log Dir", app.LogDir()},
				{"Cache Dir", app.CacheDir()},
				{"Music Cache Dir", app.MusicCacheDir()},
				{"Download Dir", app.DownloadDir()},
				{"Download Lyric Dir", app.DownloadLyricDir()},
				{"Runtime Dir", app.RuntimeDir()},
				{"Loaded Configuration File", app.ConfigFilePath()},
			}

			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithFocused(true),
				table.WithHeight(len(rows)),
			)

			s := table.DefaultStyles()
			s.Header = s.Header.
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderBottom(true).
				Bold(false)
			s.Selected = s.Selected.
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("160")).
				Bold(true)
			t.SetStyles(s)

			m := configModel{t}
			_, err := tea.NewProgram(m).Run()
			return err
		},
	}
	return cmd
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type configModel struct {
	table table.Model
}

func (m configModel) Init() tea.Cmd { return nil }

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				func() tea.Msg {
					clipboard.Write(m.table.SelectedRow()[1])
					return nil
				},
				tea.Printf("Copied to clipboard: %s", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m configModel) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}
