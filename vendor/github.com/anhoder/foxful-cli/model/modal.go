package model

import tea "charm.land/bubbletea/v2"

// Modal is an internal protocol for popup dialogs and context menus.
//
// Modal defines the lifecycle contract for items on the App's modal stack,
// including event interception, dismissal logic, and rendering coordination.
// It is deliberately unexported (all methods lowercase) and not intended as
// a public extension point. Rendering is dispatched by concrete type switch
// in app.go rather than through the interface.
//
// To show custom modal content, use Popup (see NewPopup, NewMarkdownPopup)
// rather than implementing Modal directly. Popup already supports arbitrary
// content including markdown, scrollable text, forms, and action buttons.
//
// Only Popup and ContextMenu implement this interface.
type Modal interface {
	// update handles keyboard input
	update(msg tea.Msg)

	// handleMouse handles mouse events. Returns (handled, cmd).
	// If handled=true, the modal consumed the event.
	// If handled=false, the event should pass through to underlying UI.
	handleMouse(msg tea.MouseMsg) (bool, tea.Cmd)

	// dismissed returns true if the modal should be removed from the stack
	dismissed() bool

	// dismissOutside is called when the user clicks outside the modal.
	// Returns true if the click dismissed the modal, false if it was ignored
	// (e.g. a popup configured with DisableOutsideClick).
	dismissOutside() bool

	// complete is called after dismissal to execute any result callbacks.
	// Returns (Page, tea.Cmd) for navigation/actions, or (nil, nil) if none.
	complete(app *App) (Page, tea.Cmd)

	// allowsRightClickPassthrough returns true if right-clicks outside this modal
	// should dismiss it and pass through to open a new context menu.
	// Popup returns false (traditional modal behavior); ContextMenu returns true.
	allowsRightClickPassthrough() bool
}
