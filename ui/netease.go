package ui

import (
	tea "github.com/anhoder/bubbletea"
	"github.com/anhoder/go-musicfox/constants"
	"time"
)

type neteaseModel struct {
	WindowWidth  int
	WindowHeight int

	// startup
	startupModel

	// main ui
	mainMenuModel
}

// NewNeteaseModel get netease model
func NewNeteaseModel(loadingDuration time.Duration) (m *neteaseModel) {
	m = new(neteaseModel)
	m.TotalDuration = loadingDuration
	m.menuTitle = "网易云音乐"

	return
}

func (m *neteaseModel) Init() tea.Cmd {
	if constants.AppShowStartup {
		return tickStartup(time.Nanosecond)
	}

	return tickMainUI(time.Nanosecond)
}

func (m *neteaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	switch msgWithType := msg.(type) {
	case tea.KeyMsg:
		k := msgWithType.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.WindowHeight = msgWithType.Height
		m.WindowWidth  = msgWithType.Width
	}

	// Hand off the message and model to the approprate update function for the
	// appropriate view based on the current state.
	if constants.AppShowStartup && !m.loaded {
		if _, ok := msg.(tea.WindowSizeMsg); ok {
			updateMainUI(msg, m)
		}
		return updateStartup(msg, m)
	}

	return updateMainUI(msg, m)
}

func (m *neteaseModel) View() string {
	if m.quitting {
		return ""
	}
	if constants.AppShowStartup && !m.loaded {
		return startupView(m)
	}

	return mainUIView(m)
}

