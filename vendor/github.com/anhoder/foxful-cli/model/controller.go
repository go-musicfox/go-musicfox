package model

import (
	tea "charm.land/bubbletea/v2"
)

// KeyboardController intercepts keyboard input before it reaches the default
// page handlers. Register via Options.KeyboardControllers. Controllers run in
// order; the first controller to return stopPropagation=true halts the chain
// and the key event does not reach subsequent controllers or the active page.
//
// Use cases: global hotkeys, debug overlays, custom navigation shortcuts.
type KeyboardController interface {
	// KeyMsgHandle processes a keyboard event. Return stopPropagation=true to
	// consume the event and prevent it from reaching the active page. Return
	// a non-nil newPage to trigger page navigation, or a non-nil cmd to execute
	// a bubbletea command.
	KeyMsgHandle(msg tea.KeyMsg, a *App) (stopPropagation bool, newPage Page, cmd tea.Cmd)
}

// MouseController intercepts mouse input before it reaches the default page
// handlers. Register via Options.MouseControllers. Controllers run in order;
// the first controller to return stopPropagation=true halts the chain and the
// mouse event does not reach subsequent controllers or the active page.
//
// Use cases: custom click handlers, drag-and-drop, gesture recognition.
type MouseController interface {
	// MouseMsgHandle processes a mouse event. Return stopPropagation=true to
	// consume the event and prevent it from reaching the active page. Return
	// a non-nil newPage to trigger page navigation, or a non-nil cmd to execute
	// a bubbletea command.
	MouseMsgHandle(msg tea.MouseMsg, a *App) (stopPropagation bool, newPage Page, cmd tea.Cmd)
}
