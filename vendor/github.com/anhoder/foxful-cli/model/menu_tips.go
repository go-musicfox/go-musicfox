package model

type MenuTips struct {
	main       *Main
	originMenu *MenuItem
}

func NewMenuTips(m *Main, originMenu *MenuItem) *MenuTips {
	return &MenuTips{
		main:       m,
		originMenu: originMenu,
	}
}

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

func (t *MenuTips) Recover() {
	t.main.loadingTips = ""
}
