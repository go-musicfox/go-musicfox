package ui

import (
	tea "github.com/anhoder/bubbletea"
	"time"
)

type mainMenuModel struct {

}

// update main ui
func updateMainUI(msg tea.Msg, m neteaseModel) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		return keyMsgHandle(msg, m)

	case tickMainUIMsg:
		// every second update ui
		return m, tickMainUI(time.Second)

	}

	return m, nil
}

// get main ui view
func mainUIView(m neteaseModel) string {
	//var builder strings.Builder
	//builder.WriteString()
	return "test"
}

func keyMsgHandle(msg tea.KeyMsg, m neteaseModel) (tea.Model, tea.Cmd) {
	return m, nil
}