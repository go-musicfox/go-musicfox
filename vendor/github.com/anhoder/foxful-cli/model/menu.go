package model

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/style"
)

type Hook func(main *Main) (bool, Page)

type MenuItem struct {
	Title    string
	Subtitle string
}

func (item *MenuItem) OriginString() string {
	if item.Subtitle == "" {
		return item.Title
	}
	return item.Title + " " + item.Subtitle
}

func (item *MenuItem) String() string {
	if item.Subtitle == "" {
		return item.Title
	}
	return item.Title + " " + style.CurrentStyleSet().Subtitle.Render(item.Subtitle)
}

// HelpHint describes a single keyboard shortcut displayed in the help bar
// below the menu list. Each menu can customize the hints shown for its context.
type HelpHint struct {
	Key  string // e.g. "↑↓/jk", "enter", "/"
	Desc string // e.g. "navigate", "confirm", "search"
}

// Menu menu interface
type Menu interface {
	// IsSearchable is the current menu searchable
	IsSearchable() bool

	// RealDataIndex index of real data
	RealDataIndex(index int) int

	// GetMenuKey Menu unique key
	GetMenuKey() string

	// MenuViews get submenu View
	MenuViews() []MenuItem

	// FormatMenuItem format before entering the menu
	FormatMenuItem(item *MenuItem)

	// SubMenu obtain menu by index
	SubMenu(app *App, index int) Menu

	// Action is called when the user activates a menu item (Enter/double-click).
	// If it returns a non-nil Page or Cmd, the action is executed and submenu
	// navigation is skipped. Return (nil, nil) to fall through to SubMenu.
	//
	// Use cases: show a popup, write a log, trigger a side effect, or any
	// arbitrary action that should not navigate to a submenu.
	Action(app *App, index int) (Page, tea.Cmd)

	// HelpHints returns the keyboard shortcuts to display in the help bar
	// below the menu list. Return nil to hide the help bar for this menu.
	HelpHints() []HelpHint

	// BeforePrePageHook Hook before turn to previous page
	BeforePrePageHook() Hook

	// BeforeNextPageHook Hook before turn to next page
	BeforeNextPageHook() Hook

	// BeforeEnterMenuHook Hook before enter menu
	BeforeEnterMenuHook() Hook

	// BeforeBackMenuHook Hook before back menu
	BeforeBackMenuHook() Hook

	// BottomOutHook Hook while bottom out
	BottomOutHook() Hook

	// TopOutHook Hook while top out
	TopOutHook() Hook

	// ContextMenuItems returns the context menu items for a right-clicked menu item.
	// Return nil or empty slice to show no context menu for this item.
	ContextMenuItems(app *App, index int) []ContextMenuItem

	// ContextMenuAction is called when the user selects a context menu item.
	// Similar to Action, it can return a Page and/or Cmd to perform navigation or side effects.
	ContextMenuAction(app *App, index int, item ContextMenuItem) (Page, tea.Cmd)
}

type LocalSearchMenu interface {
	Menu
	Search(menu Menu, search string)
}

type DefaultMenu struct{}

func (e *DefaultMenu) IsSearchable() bool {
	return false
}

func (e *DefaultMenu) RealDataIndex(index int) int {
	return index
}

// GetMenuKey returns a stable identifier for the menu, derived from its
// concrete type. Override this to provide an explicit, stable key when your
// app persists or routes by menu key.
//
// Caveat: because Go method embedding does not give the embedded DefaultMenu
// access to the outer concrete type, this default returns the same value
// ("*model.DefaultMenu") for every menu that embeds DefaultMenu. Any menu
// that requires a unique key must override this method.
func (e *DefaultMenu) GetMenuKey() string {
	panic("implement me")
}

func (e *DefaultMenu) MenuViews() []MenuItem {
	return nil
}

func (e *DefaultMenu) FormatMenuItem(_ *MenuItem) {
}

func (e *DefaultMenu) SubMenu(_ *App, _ int) Menu {
	return nil
}

func (e *DefaultMenu) Action(_ *App, _ int) (Page, tea.Cmd) {
	return nil, nil
}

// HelpHints returns a default set of keyboard shortcuts.
// Individual menus can override this to provide context-specific hints.
func (e *DefaultMenu) HelpHints() []HelpHint {
	return []HelpHint{
		{Key: "↑↓/jk", Desc: "navigate"},
		{Key: "n/enter", Desc: "confirm"},
		{Key: "/", Desc: "search"},
		{Key: "b/esc", Desc: "back"},
		{Key: "q", Desc: "quit"},
	}
}

func (e *DefaultMenu) BeforePrePageHook() Hook {
	return nil
}

func (e *DefaultMenu) BeforeNextPageHook() Hook {
	return nil
}

func (e *DefaultMenu) BeforeEnterMenuHook() Hook {
	return nil
}

func (e *DefaultMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (e *DefaultMenu) BottomOutHook() Hook {
	return nil
}

func (e *DefaultMenu) TopOutHook() Hook {
	return nil
}

func (e *DefaultMenu) ContextMenuItems(_ *App, _ int) []ContextMenuItem {
	return nil
}

func (e *DefaultMenu) ContextMenuAction(_ *App, _ int, _ ContextMenuItem) (Page, tea.Cmd) {
	return nil, nil
}

type Closer interface {
	Close() error
}

type Ticker interface {
	Closer
	Start() error
	Ticker() <-chan time.Time
	PassedTime() time.Duration
}

type defaultTicker struct {
	startTime time.Time
	t         time.Time
	ticker    *time.Ticker
	stop      chan struct{}
	pipeline  chan time.Time
	closed    bool
}

func DefaultTicker(duration time.Duration) Ticker {
	return &defaultTicker{
		ticker:   time.NewTicker(duration),
		stop:     make(chan struct{}),
		pipeline: make(chan time.Time),
	}
}

func (d *defaultTicker) Start() error {
	d.startTime = time.Now()
	go func() {
		for {
			select {
			case <-d.stop:
				return
			case d.t = <-d.ticker.C:
				// ignore data race at d.t
				select {
				case d.pipeline <- d.t:
				default:
				}
			}
		}
	}()
	return nil
}

func (d *defaultTicker) Ticker() <-chan time.Time {
	return d.pipeline
}

func (d *defaultTicker) PassedTime() time.Duration {
	// Before the first tick arrives, d.t is zero-valued.
	// Return 0 to prevent downstream code from computing negative indices.
	if d.t.IsZero() {
		return 0
	}
	return d.t.Sub(d.startTime)
}
func (d *defaultTicker) Close() error {
	if d.closed {
		return nil
	}
	d.closed = true
	close(d.stop)
	d.ticker.Stop()
	return nil
}
