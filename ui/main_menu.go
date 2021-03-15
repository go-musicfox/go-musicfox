package ui

import (
	tea "github.com/anhoder/bubbletea"
	"github.com/anhoder/go-musicfox/constants"
	"strings"
	"time"
	"unicode/utf8"
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
	if m.WindowWidth <= 0 || m.WindowHeight <= 0 {
		return ""
	}

	var builder strings.Builder

	// title
	if constants.MainShowTitle {
		var titleBuilder strings.Builder
		titleLen := utf8.RuneCountInString(constants.AppName)+2
		prefixLen := (m.WindowWidth-titleLen)/2
		suffixLen := m.WindowWidth-prefixLen-titleLen
		titleBuilder.WriteString(strings.Repeat("─", prefixLen))
		titleBuilder.WriteString(" ")
		titleBuilder.WriteString(strings.ToTitle(constants.AppName))
		titleBuilder.WriteString(" ")
		titleBuilder.WriteString(strings.Repeat("─", suffixLen))

		builder.WriteString(SetFgStyle(titleBuilder.String(), primaryColor))
	}

	//

	return builder.String()
}

func keyMsgHandle(msg tea.KeyMsg, m neteaseModel) (tea.Model, tea.Cmd) {
	return m, nil
}