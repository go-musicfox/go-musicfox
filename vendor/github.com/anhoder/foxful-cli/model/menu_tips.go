package model

// MenuTips displays transient text beside the current menu title (e.g. loading
// indicators, status messages). Create via NewMenuTips and use DisplayTips to
// show text; Recover to clear it. The text is rendered through bubbletea's
// normal diff-based pipeline (not direct terminal writes) to avoid layout
// corruption when entering/exiting submenus.
type MenuTips struct {
	main       *Main
	originMenu *MenuItem
}

// NewMenuTips creates a MenuTips instance for the given Main page.
// originMenu is reserved for future use and currently unused.
func NewMenuTips(m *Main, originMenu *MenuItem) *MenuTips {
	return &MenuTips{
		main:       m,
		originMenu: originMenu,
	}
}

// DisplayTips sets transient text to display beside the menu title in the next
// View cycle. The tips string appears styled with the Subtitle style, adjacent
// to the current menu title. Use this for loading indicators, progress messages,
// or status updates that should not persist across menu navigation.
func (t *MenuTips) DisplayTips(tips string) {
	// Set a transient loading text on Main so that the next View() cycle
	// renders it through bubbletea's normal diff-based pipeline.
	// Direct terminal writes (fmt.Print, terrmenv.MoveCursor) are NOT used
	// here because they bypass bubbletea's renderer and cause the
	// diff algorithm to produce incorrect output, resulting in submenu
	// layout corruption (title bar overwritten, items at wrong rows,
	// stale main menu content ghosting through).
	t.main.loadingTips = tips
}

// Recover clears the transient tips text so the next View renders without it.
func (t *MenuTips) Recover() {
	t.main.loadingTips = ""
}
