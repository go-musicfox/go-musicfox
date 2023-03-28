package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// startup tick
type tickStartupMsg struct{}

func tickStartup(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return tickStartupMsg{}
	})
}

// main ui tick
type tickMainUIMsg struct{}

func tickMainUI(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return tickMainUIMsg{}
	})
}

// login tick
type tickLoginMsg struct{}

func tickLogin(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return tickLoginMsg{}
	})
}

// search tick
type tickSearchMsg struct{}

func tickSearch(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return tickSearchMsg{}
	})
}
