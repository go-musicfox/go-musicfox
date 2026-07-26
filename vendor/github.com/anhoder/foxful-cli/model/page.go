package model

import (
	tea "charm.land/bubbletea/v2"
)

// Page is a full-screen view managed by App. Exactly one Page is active at a
// time; App dispatches messages to it and renders its View. Implement this
// interface to create custom screens, then activate a Page by returning it
// from a Menu.Action, a controller, or Options.InitPage.
//
// The framework provides two built-in pages: StartupPage (PtStartup) and
// Main (PtMain).
type Page interface {
	// IgnoreQuitKeyMsg returns true if the given key should NOT trigger the
	// global quit behavior while this page is active (e.g. to allow typing
	// "q" into a text field).
	IgnoreQuitKeyMsg(msg tea.KeyMsg) bool

	// Type returns the page's type identifier (PtStartup, PtMain, or a custom
	// value for user-defined pages).
	Type() PageType

	// Update handles a message and returns the next page to display (return
	// the receiver to stay on the current page) plus an optional command.
	Update(msg tea.Msg, a *App) (Page, tea.Cmd)

	// View renders the page to a string for the current terminal size.
	View(a *App) string

	// Msg returns an optional message to emit when this page becomes active,
	// used to trigger initial data loading or animations. Return nil for none.
	Msg() tea.Msg
}

// InitPage is an optional extension of Page for pages that need a one-time
// initialization command when first shown (e.g. to start a timer or kick off
// an async load). App calls Init during startup if the initial page implements
// this interface.
type InitPage interface {
	Page
	// Init returns a command executed once when the page is first activated.
	Init(a *App) tea.Cmd
}

// PageType identifies the kind of a Page. Built-in values are PtStartup and
// PtMain; custom pages may define their own values.
type PageType string

const (
	PtStartup PageType = "startup" // the startup/splash page
	PtMain    PageType = "main"    // the default main menu page
)
