package model

import (
	tea "charm.land/bubbletea/v2"
)

// Component is a supplementary UI element that renders below the menu list
// (e.g., audio spectrum visualizer, lyrics panel, progress bar).
//
// Components are registered via Options.Components and rendered in order
// at the bottom of the Main page view.
type Component interface {
	// Update is called from Main.Update for every message while Main is the
	// active page, after the message's primary handling. Components should
	// filter messages themselves and mutate their own state as needed.
	// They do not return commands (trigger redraws via the app's ticker or
	// call app.Rerender if needed).
	//
	// Note: Messages consumed by an open modal (popup/context menu) are not
	// delivered to components.
	Update(msg tea.Msg, a *App)

	// View renders the component's current state. Return empty string to
	// render nothing. The lines return value is currently unused.
	View(a *App, main *Main) (view string, lines int)
}
