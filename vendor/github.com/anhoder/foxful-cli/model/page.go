package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Page interface {
	IgnoreQuitKeyMsg(msg tea.KeyMsg) bool
	Type() PageType
	Update(msg tea.Msg, a *App) (Page, tea.Cmd)
	View(a *App) string
	Msg() tea.Msg
}

type InitPage interface {
	Page
	Init(a *App) tea.Cmd
}

type PageType string

const (
	PtStartup PageType = "startup"
	PtMain    PageType = "main"
)
