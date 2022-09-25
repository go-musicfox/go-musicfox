package ui

import tea "github.com/anhoder/bubbletea"

type Page interface {
	update(msg tea.Msg, m *NeteaseModel) (tea.Model, tea.Cmd)
	view(m *NeteaseModel) string
}
