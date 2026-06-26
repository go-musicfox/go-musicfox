package model

import (
	tea "charm.land/bubbletea/v2"
)

type Component interface {
	Update(msg tea.Msg, a *App)
	View(a *App, main *Main) (view string, lines int)
}
