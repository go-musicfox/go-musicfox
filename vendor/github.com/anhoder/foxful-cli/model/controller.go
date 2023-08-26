package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

type KeyboardController interface {
	KeyMsgHandle(msg tea.KeyMsg, a *App) (stopPropagation bool, newPage Page, cmd tea.Cmd)
}

type MouseController interface {
	MouseMsgHandle(msg tea.MouseMsg, a *App) (stopPropagation bool, newPage Page, cmd tea.Cmd)
}
