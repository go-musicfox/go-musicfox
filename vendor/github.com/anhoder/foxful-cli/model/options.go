package model

import (
	"time"

	"github.com/anhoder/foxful-cli/util"
	tea "github.com/charmbracelet/bubbletea"
)

type Options struct {
	StartupOptions
	ProgressOptions

	AppName             string
	WhetherDisplayTitle bool
	LoadingText         string
	PrimaryColor        string
	DualColumn          bool // The menu list is displayed as a dual column
	DynamicRowCount     bool // If true, the number of entries per page can be greater than 10
	MaxMenuStartRow     int  // Max number of rows occupied by the title section before the menu. Works only when DynamicRowCount is on.
	CenterEverything    bool // If true, everything will be centered. Otherwise, use default layout.
	HideMenu            bool

	TeaOptions []tea.ProgramOption // Tea program options

	InitPage        InitPage
	MainMenuTitle   *MenuItem
	Ticker          Ticker          // Ticker for render
	MainMenu        Menu            // Entry menu of app
	LocalSearchMenu LocalSearchMenu // Local search result menu
	Components      []Component     // Custom Extra components

	GlobalKeyHandlers map[string]GlobalKeyHandler
	KBControllers     []KeyboardController
	MouseControllers  []MouseController

	InitHook  func(a *App)
	CloseHook func(a *App)
}

type StartupOptions struct {
	EnableStartup     bool
	LoadingDuration   time.Duration
	TickDuration      time.Duration
	ProgressOutBounce bool
	Welcome           string
}

func DefaultOptions() *Options {
	return &Options{
		StartupOptions: StartupOptions{
			EnableStartup:     true,
			LoadingDuration:   time.Second * 2,
			TickDuration:      time.Millisecond * 16,
			ProgressOutBounce: true,
			Welcome:           util.PkgName,
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
		WhetherDisplayTitle: true,
		DualColumn:          true,
		DynamicRowCount:     false,
		MaxMenuStartRow:     0,
		CenterEverything:    false,
		AppName:             util.PkgName,
		LoadingText:         util.LoadingText,
		PrimaryColor:        util.RandomColor,
		MainMenu:            &DefaultMenu{},
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
