package model

import (
	"time"

	"github.com/anhoder/foxful-cli/util"
	"github.com/muesli/termenv"
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
	return item.Title + " " + util.SetFgStyle(item.Subtitle, termenv.ANSIBrightBlack)
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
}

type LocalSearchMenu interface {
	Menu
	Search(menu Menu, search string)
}

type DefaultMenu struct {
}

func (e *DefaultMenu) IsSearchable() bool {
	return false
}

func (e *DefaultMenu) RealDataIndex(index int) int {
	return index
}

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
	// ignore data race at d.t
	return d.t.Sub(d.startTime)
}

func (d *defaultTicker) Close() error {
	close(d.stop)
	d.ticker.Stop()
	return nil
}
