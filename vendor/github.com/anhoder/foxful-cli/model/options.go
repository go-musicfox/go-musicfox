package model

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
)

// TabConfig defines a single tab in multi-tab main menu mode.
// Each tab has an isolated menu hierarchy, scroll position, and navigation stack.
type TabConfig struct {
	Title      string
	Menu       Menu
	MenuTitle  *MenuItem
	OnActivate func(m *Main, prevTabIndex int) bool // Called when tab becomes active. Return false to veto the switch.
}

// ContextMenuOptions configures the size limits of right-click context menus.
// MaxWidth and MaxHeight include the border; zero leaves that dimension unlimited.
type ContextMenuOptions struct {
	MaxWidth  int
	MaxHeight int
}

type Options struct {
	StartupOptions
	ProgressOptions
	NotificationOptions

	ContextMenuOptions  ContextMenuOptions
	AppName             string
	WhetherDisplayTitle bool
	LoadingText         string
	PrimaryColor        string
	DualColumn          bool // The menu list is displayed as a dual column
	DynamicRowCount     bool // If true, the number of entries per page can be greater than 10
	MaxMenuStartRow     int  // Max number of rows occupied by the title section before the menu. 0 means no limit.
	BottomHeight        int  // Height of the bottom area reserved for components (e.g. spectrum, lyrics, progress bar). Only effective when DynamicRowCount is true. 0 means use the default.
	CenterEverything    bool // If true, everything will be centered. Otherwise, use default layout.
	HideMenu            bool
	DarkTheme           style.Theme // Dark variant for adaptive theme pair. If zero-valued, DefaultTheme is used.
	LightTheme          style.Theme // Light variant for adaptive theme pair. If zero-valued, DefaultTheme is used.

	TeaOptions []tea.ProgramOption // Tea program options

	InitPage          InitPage
	MainMenuTitle     *MenuItem
	Ticker            Ticker            // Ticker for render
	MainMenu          Menu              // Entry menu when EnableTabs is false. When EnableTabs is true, TabConfigs[0].Menu becomes the entry menu and this field is ignored.
	LocalSearchMenu   LocalSearchMenu   // Local search result menu
	Components        []Component       // Custom Extra components
	StatusBar         StatusBar         // Custom status bar, nil = no status bar
	StatusBarPosition StatusBarPosition // Position of status bar: StatusBarBottom (default) or StatusBarTop

	// EnableTabs activates multi-tab navigation in the Main page. When true,
	// TabConfigs defines the available tabs; when false (default), MainMenu and
	// MainMenuTitle are used. Tab switching keys: Ctrl+Tab, Ctrl+Shift+Tab,
	// Ctrl+Left, Ctrl+Right. Each tab maintains isolated menu stack and scroll position.
	EnableTabs bool
	// TabConfigs defines the tabs when EnableTabs is true. Each tab has an
	// isolated menu stack and scroll position. Tabs are static (no runtime
	// add/remove). Keep count ≤8 for usability.
	TabConfigs []TabConfig

	GlobalKeyHandlers map[string]GlobalKeyHandler
	KBControllers     []KeyboardController
	MouseControllers  []MouseController

	InitHook  func(a *App)
	CloseHook func(a *App)

	AltScreen bool
	MouseMode tea.MouseMode
}

type StartupOptions struct {
	EnableStartup     bool
	LoadingDuration   time.Duration
	TickDuration      time.Duration
	ProgressOutBounce bool
	Welcome           string

	// Animation selects the startup visual. The default Sequence combines fade,
	// rainbow sweep, glitch, spinner, and staged status text. See the
	// StartupAnimation constants for all available effects.
	Animation StartupAnimation
	// ReducedMotion renders a static, readable final frame. It is useful for
	// accessibility, automated environments, and slow remote terminals.
	ReducedMotion bool
}

// NotificationOptions configures the global notification system behavior.
type NotificationOptions struct {
	Anchor         PopupAnchor   // Screen anchor. Default AnchorTopRight.
	DefaultTimeout time.Duration // Auto-dismiss timeout for Info/Success. Default 4s.
	MaxWidth       int           // Whole-notification max width. 0 = min(termWidth/3, 60).
	MaxLines       int           // Max message body lines before truncation. Default 5.
	Gap            int           // Vertical gap between stacked notifications. Default 1.
}

func DefaultOptions() *Options {
	return &Options{
		StartupOptions: StartupOptions{
			EnableStartup:   true,
			LoadingDuration: time.Second * 2,
			// The startup renderer updates complete lines. 20 FPS keeps effects
			// smooth while remaining suitable for SSH and low-power terminals.
			TickDuration:      time.Millisecond * 50,
			ProgressOutBounce: true,
			Welcome:           util.PkgName,
			Animation:         StartupAnimationSequence,
		},
		ProgressOptions: ProgressOptions{
			EmptyCharWhenFirst: '.',
			EmptyChar:          '.',
			EmptyCharWhenLast:  '.',
			FirstEmptyChar:     '.',
			FullCharWhenFirst:  '#',
			FullChar:           '#',
			FullCharWhenLast:   '#',
			LastFullChar:       '#',
		},
		NotificationOptions: NotificationOptions{
			Anchor:         AnchorTopRight,
			DefaultTimeout: 4 * time.Second,
			MaxWidth:       0,
			MaxLines:       5,
			Gap:            1,
		},
		ContextMenuOptions: ContextMenuOptions{
			MaxWidth:  0,
			MaxHeight: 0,
		},
		WhetherDisplayTitle: true,
		StatusBarPosition:   StatusBarBottom,
		DualColumn:          true,
		DynamicRowCount:     false,
		MaxMenuStartRow:     0,
		CenterEverything:    false,
		AppName:             util.PkgName,
		LoadingText:         util.LoadingText,
		PrimaryColor:        util.RandomColor,
		MainMenu:            &DefaultMenu{},
		AltScreen:           true,
		MouseMode:           tea.MouseModeAllMotion,
	}
}

type WithOption func(options *Options)

func WithHook(init, close func(a *App)) WithOption {
	return func(opts *Options) {
		opts.InitHook = init
		opts.CloseHook = close
	}
}

func WithMainMenu(mainMenu Menu, mainMenuTitle *MenuItem) WithOption {
	return func(opts *Options) {
		opts.MainMenu = mainMenu
		opts.MainMenuTitle = mainMenuTitle
	}
}

func WithGlobalKeyHandlers(m map[string]GlobalKeyHandler) WithOption {
	return func(options *Options) {
		options.GlobalKeyHandlers = m
	}
}

// WithThemePair sets a dark/light theme pair for adaptive appearance.
// The appropriate theme is selected automatically based on terminal
// background detection. Usage:
//
//	app.With(model.WithThemePair(style.DefaultDarkTheme(), style.DefaultLightTheme()))
func WithThemePair(dark, light style.Theme) WithOption {
	return func(options *Options) {
		options.DarkTheme = dark
		options.LightTheme = light
	}
}

// WithNotificationOptions overrides the notification system configuration.
func WithNotificationOptions(opts NotificationOptions) WithOption {
	return func(o *Options) {
		o.NotificationOptions = opts
	}
}

// WithContextMenuOptions overrides the right-click context menu configuration.
func WithContextMenuOptions(opts ContextMenuOptions) WithOption {
	return func(o *Options) {
		o.ContextMenuOptions = opts
	}
}
