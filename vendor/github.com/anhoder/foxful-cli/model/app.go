package model

import (
	"sync"
	"time"

	"github.com/anhoder/foxful-cli/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

type App struct {
	windowWidth  int
	windowHeight int
	options      *Options
	quiting      bool

	program *tea.Program

	startup *StartupPage
	main    *Main

	page Page // current page

	listeningKBEventL    sync.Mutex
	listeningMouseEventL sync.Mutex
}

// NewApp create application
func NewApp(options *Options) (a *App) {
	a = &App{
		options: options,
		page:    options.InitPage,
	}

	runewidth.DefaultCondition.EastAsianWidth = false

	return
}

type WithOption func(options *Options)

func (a *App) With(w ...WithOption) *App {
	for _, item := range w {
		if item != nil {
			item(a.options)
		}
	}
	return a
}

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

func (a *App) Init() tea.Cmd {
	if a.options.InitHook != nil {
		a.options.InitHook(a)
	}
	if a.options.Ticker != nil {
		go func() {
			for range a.options.Ticker.Ticker() {
				a.Rerender(false)
			}
		}()
		if err := a.options.Ticker.Start(); err != nil {
			panic("Fail to start ticker: " + err.Error())
		}
	}
	if initPage, ok := a.page.(InitPage); ok {
		return initPage.Init(a)
	}
	return nil
}

func (a *App) Close() {
	if a.options.CloseHook != nil {
		a.options.CloseHook(a)
	}
	if closer, ok := a.page.(Closer); ok {
		_ = closer.Close()
	}
	if a.options.Ticker != nil {
		_ = a.options.Ticker.Close()
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		if !a.listeningKBEventL.TryLock() {
			return a, nil
		}
		defer a.listeningKBEventL.Unlock()
	} else if _, ok := msg.(tea.MouseMsg); ok {
		if !a.listeningMouseEventL.TryLock() {
			return a, nil
		}
		defer a.listeningMouseEventL.Unlock()
	}

	// Make sure these keys always quit
	switch msgWithType := msg.(type) {
	case tea.KeyMsg:
		var k = msgWithType.String()
		if k != "q" && k != "Q" && k != "ctrl+c" {
			break
		}
		if a.page != nil && a.page.IgnoreQuitKeyMsg(msgWithType) {
			break
		}
		a.Close()
		a.quiting = true
		return a, tea.Quit
	case tea.WindowSizeMsg:
		a.windowHeight = msgWithType.Height
		a.windowWidth = msgWithType.Width
	}

	page, cmd := a.page.Update(msg, a)
	a.page = page
	return a, cmd
}

func (a *App) View() string {
	if a.quiting || a.WindowHeight() <= 0 || a.WindowWidth() <= 0 {
		return ""
	}

	return a.page.View(a)
}

func (a *App) Run() error {
	util.PrimaryColor = a.options.PrimaryColor

	if a.page == nil {
		a.main = NewMain(a, a.options)
		a.startup = NewStartup(&a.options.StartupOptions, a.main)
		if a.options.InitPage == nil {
			a.options.InitPage = a.main
			if a.options.EnableStartup {
				a.options.InitPage = a.startup
			}
		}
		a.page = a.options.InitPage
	}
	a.program = tea.ReplaceWithFoxfulRenderer(tea.NewProgram(a, a.options.TeaOptions...))
	_, err := a.program.Run()
	return err
}

func (a *App) Rerender(cleanScreen bool) {
	if a.program == nil {
		return
	}
	a.program.Send(a.RerenderCmd(cleanScreen))
}

func (a *App) RerenderCmd(cleanScreen bool) tea.Cmd {
	return func() tea.Msg {
		if cleanScreen {
			a.program.Send(tea.ClearScreen())
		}
		return a.page.Msg()
	}
}

func (a *App) Tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return a.page.Msg()
	})
}

func (a *App) WindowWidth() int {
	return a.windowWidth
}

func (a *App) WindowHeight() int {
	return a.windowHeight
}

func (a *App) CurPage() Page {
	return a.page
}

func (a *App) Startup() *StartupPage {
	return a.startup
}

func (a *App) Main() *Main {
	return a.main
}

func (a *App) MustMain() *Main {
	if a.main != nil {
		return a.main
	}
	panic("main page is empty")
}

func (a *App) MustStartup() *StartupPage {
	if a.startup != nil {
		return a.startup
	}
	panic("startup page is empty")
}

func (a *App) Options() *Options {
	return a.options
}
