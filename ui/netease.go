package ui

import (
	tea "github.com/anhoder/bubbletea"
	"time"
)

type neteaseModel struct {
	WindowWidth  int
	WindowHeight int

	// startup
	startupModel

	// main ui
}

// NewNeteaseModel get netease model
func NewNeteaseModel(loadingDuration time.Duration) (m *neteaseModel) {
	m = new(neteaseModel)
	m.TotalDuration = loadingDuration

	return
}

func (m neteaseModel) Init() tea.Cmd {
	return tickStartup(time.Nanosecond)
}

func (m neteaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.WindowHeight = msg.Height
		m.WindowWidth  = msg.Width
	}

	// Hand off the message and model to the approprate update function for the
	// appropriate view based on the current state.
	if !m.loaded {
		return updateStartup(msg, m)
	}

	return updateMainUI(msg, m)
}

func (m neteaseModel) View() string {
	if m.quitting {
		return ""
	}
	if !m.loaded {
		return startupView(m)
	}

	return mainUIView(m)
}

