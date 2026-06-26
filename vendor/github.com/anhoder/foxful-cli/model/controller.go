package model

import (
	tea "charm.land/bubbletea/v2"
)

type KeyboardController interface {
	KeyMsgHandle(msg tea.KeyMsg, a *App) (stopPropagation bool, newPage Page, cmd tea.Cmd)
}

type MouseController interface {
	MouseMsgHandle(msg tea.MouseMsg, a *App) (stopPropagation bool, newPage Page, cmd tea.Cmd)
}
