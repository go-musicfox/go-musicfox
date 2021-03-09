package ui

import (
	tea "github.com/anhoder/bubbletea"
	"time"
)

type NeteaseModel struct {
	TotalDuration  time.Duration
	loadedDuration time.Duration
	loadedPercent  float64
	Loaded         bool
	Quitting       bool
}

func (m NeteaseModel) Init() tea.Cmd {
	return tickStartup(time.Nanosecond)
}

func (m NeteaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		WindowHeight = msg.Height
		WindowWidth  = msg.Width
	}

	// Hand off the message and model to the approprate update function for the
	// appropriate view based on the current state.
	if !m.Loaded {
		return updateStartup(msg, m)
	}

	return updateMainUI(msg, m)
}

func (m NeteaseModel) View() string {
	if m.Quitting {
		return ""
	}
	if !m.Loaded {
		return startupView(m)
	}

	return mainUIView(m)
}

