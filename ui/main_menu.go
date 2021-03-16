package ui

import (
	tea "github.com/anhoder/bubbletea"
	"github.com/anhoder/go-musicfox/constants"
	"math"
	"strings"
	"time"
	"unicode/utf8"
)

type mainMenuModel struct {
	doubleColumn  bool // 是否双列显示
	startRow      int  // 开始行
	startColumn   int  // 开始列
	menuBottomRow int  // 菜单最底部所在行

	menuCurPage	  int // 菜单当前页
	menuPageSize  int // 菜单每页大小
}

// update main ui
func updateMainUI(msg tea.Msg, m neteaseModel) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		return keyMsgHandle(msg, m)

	case tickMainUIMsg:
		// every second update ui
		return m, tickMainUI(time.Second)

	case tea.WindowSizeMsg:
		m.doubleColumn = msg.Width >= 80
		m.startRow = msg.Height/3

		if m.doubleColumn {
			m.startColumn = (msg.Width-60)/2
			m.menuBottomRow = m.startRow+int(math.Ceil(float64(m.menuPageSize)/2))-1
		} else {
			m.startColumn = (msg.Width-20)/2
			m.menuBottomRow = m.startRow+m.menuPageSize-1
		}

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
		titleBuilder.WriteString(strings.ToUpper(constants.AppName))
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