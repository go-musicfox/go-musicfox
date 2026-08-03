package model

import (
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/layout"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

type backgroundRect struct {
	x int
	y int
	w int
	h int
}

type App struct {
	windowWidth  int
	windowHeight int
	options      *Options
	quiting      bool
	themeIndex   int // current position in Options.ThemeList

	program *tea.Program

	startup *StartupPage
	main    *Main

	page       Page    // current page
	modalStack []Modal // stack of active modals (popups, context menus); topmost is last

	notifications      []*Notification // active notifications (newest at end)
	nextNotificationID NotificationID

	// styleSet is the app-scoped theme. When nil, StyleSet() falls back to the
	// global style.CurrentStyleSet(). Set via SetStyleSet to isolate this app's
	// theme from the global state (e.g. for multi-app or parallel-test scenarios).
	styleSet *style.StyleSet

	appBackgroundExclusion backgroundRect

	listeningKBEventL    sync.Mutex
	listeningMouseEventL sync.Mutex

	// pageMu 保护 page 字段。page 只可由主事件循环 goroutine（App.Update
	// 与 App.Run 启动阶段）写入，命令或 ticker goroutine 不得写入。写入必须
	// 使用 setPage（Lock），并发读取（Rerender、RerenderCmd、Tick、CurPage）
	// 必须使用 getPage（RLock）。主循环读取可直接访问 page，因为它们与唯一
	// 写入者串行；从 goroutine 写入 page 前必须重新检查所有直接读取点。
	pageMu sync.RWMutex
}

// StyleSet returns the app-scoped StyleSet if one was set via SetStyleSet,
// otherwise the global style.CurrentStyleSet(). Internal render paths should
// prefer this over style.CurrentStyleSet() so per-app themes take effect.
func (a *App) StyleSet() style.StyleSet {
	if a.styleSet != nil {
		return *a.styleSet
	}
	return style.CurrentStyleSet()
}

// SetStyleSet sets an app-scoped StyleSet, isolating this app's theme from the
// global style state. Pass a StyleSet built via style.NewStyleSet. Once set,
// this app's rendering uses the given StyleSet regardless of global changes.
func (a *App) SetStyleSet(s style.StyleSet) {
	a.styleSet = &s
}

// SetThemePair updates the dark and light variants used for future system
// appearance changes.
func (a *App) SetThemePair(dark, light style.Theme) {
	a.options.DarkTheme = dark
	a.options.LightTheme = light
}

// SetAppBackgroundExclusion excludes a terminal-cell rectangle from the app
// background applied by Main.View. Invalid dimensions clear the exclusion.
func (a *App) SetAppBackgroundExclusion(x, y, width, height int) {
	if width <= 0 || height <= 0 {
		a.appBackgroundExclusion = backgroundRect{}
		return
	}
	a.appBackgroundExclusion = backgroundRect{x: x, y: y, w: width, h: height}
}

// ClearAppBackgroundExclusion restores application background rendering everywhere.
func (a *App) ClearAppBackgroundExclusion() {
	a.appBackgroundExclusion = backgroundRect{}
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

func (a *App) With(w ...WithOption) *App {
	for _, item := range w {
		if item != nil {
			item(a.options)
		}
	}
	return a
}

func (a *App) Init() tea.Cmd {
	if a.options.InitHook != nil {
		a.options.InitHook(a)
	}

	var cmds []tea.Cmd
	if a.options.Ticker != nil {
		go func() {
			for range a.options.Ticker.Ticker() {
				a.Rerender(false)
			}
		}()
		if err := a.options.Ticker.Start(); err != nil {
			// Ticker start failure is rare (system resource exhaustion).
			// Degrade gracefully: the goroutine blocks harmlessly, the app
			// continues without ticks. Surface the error via notification
			// so it's not silently swallowed.
			cmds = append(cmds, func() tea.Msg {
				return ShowNotificationMsg{
					Spec: NotificationSpec{
						Title:   "Ticker Start Failed",
						Message: err.Error(),
						Level:   NotificationError,
					},
				}
			})
		}
	}

	// Request initial terminal background color to seed light/dark detection.
	// Handled in Update via BackgroundColorMsg.
	cmds = append(cmds, func() tea.Msg {
		return tea.RequestBackgroundColor()
	})
	// Enable DEC 2031 mode so the terminal proactively reports system
	// light/dark scheme changes. When enabled, the terminal sends a
	// ModeReportMsg with Mode=DECMode(2031) on every color scheme switch.
	// Supported by kitty, WezTerm, Ghostty, iTerm2, and others.
	cmds = append(cmds, tea.Raw(ansi.SetModeLightDark))

	if initPage, ok := a.page.(InitPage); ok {
		cmds = append(cmds, initPage.Init(a))
	}
	return tea.Batch(cmds...)
}

func (a *App) Close() {
	// Reset terminal mouse pointer to default on exit
	resetMousePointer()

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

// SetMousePointer returns a tea.Cmd that sends an OSC 22 escape sequence to
// change the terminal mouse pointer shape when hovering over the terminal
// window. Supported by Kitty, WezTerm, Ghostty, iTerm2, Terminal.app, and
// others. Unsupported terminals silently ignore the sequence.
//
// Shape names follow CSS cursor conventions. Common values:
//   - "default"  — normal arrow pointer
//   - "pointer"  — pointing hand (for clickable elements)
//   - "text"     — I-beam (for editable text)
//   - "wait"     — wait spinner (for busy state)
//   - "crosshair" — crosshair (for grid selection)
func (a *App) SetMousePointer(shape string) tea.Cmd {
	return setMousePointer(shape)
}

// resetMousePointer writes the OSC 22 reset sequence directly to stdout.
// Used synchronously during shutdown where tea.Cmd cannot be returned.
func resetMousePointer() {
	writeMousePointer("")
}

func (a *App) Update(msg tea.Msg) (model tea.Model, returnCmd tea.Cmd) {
	var notificationHoverCmd tea.Cmd
	defer func() {
		if notificationHoverCmd != nil {
			returnCmd = tea.Batch(notificationHoverCmd, returnCmd)
		}
	}()
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

	// Modal input interception must happen BEFORE quit key check,
	// so modals can handle keys like 'q' as close keys.
	// Notification messages are handled before modal interception so they work
	// regardless of any open modal. These are never Key/Mouse messages, so the
	// locks at the top of Update do not apply.
	switch msgWithType := msg.(type) {
	case ShowNotificationMsg:
		return a, a.handleShowNotification(msgWithType.Spec)
	case notificationExpireMsg:
		a.handleExpire(msgWithType.id)
		return a, a.RerenderCmd(true)
	case updateNotificationMsg:
		cmd := a.updateNotificationContent(msgWithType.id, msgWithType.spec)
		if cmd != nil {
			return a, tea.Batch(a.RerenderCmd(true), cmd)
		}
		return a, a.RerenderCmd(true)
	case dismissNotificationMsg:
		a.removeNotification(msgWithType.id)
		return a, a.RerenderCmd(true)
	case clearAllNotificationsMsg:
		a.notifications = nil
		return a, a.RerenderCmd(true)
	}

	// Theme switch shortcut: cycle to the next theme in ThemeList.
	// Works globally regardless of modals.
	{
		if k, ok := msg.(tea.KeyPressMsg); ok {
			key := k.String()
			if a.options.ThemeSwitchKey != "" && len(a.options.ThemeList) > 0 && key == a.options.ThemeSwitchKey {
				a.themeIndex = (a.themeIndex + 1) % len(a.options.ThemeList)
				style.SetStyleSet(style.NewStyleSet(a.resolveTheme()))
				return a, a.RerenderCmd(true)
			}
		}
	}

	// Make sure these keys always quit (but only if no modal is handling them)
	if len(a.modalStack) == 0 {
		switch msgWithType := msg.(type) {
		case tea.KeyPressMsg:
			k := msgWithType.String()
			if k == "q" || k == "Q" || k == "ctrl+c" {
				if a.page == nil || !a.page.IgnoreQuitKeyMsg(msgWithType) {
					a.Close()
					a.quiting = true
					return a, tea.Quit
				}
			}
		}
	}

	switch msgWithType := msg.(type) {
	case tea.WindowSizeMsg:
		a.windowHeight = msgWithType.Height
		a.windowWidth = msgWithType.Width
	case tea.BackgroundColorMsg:
		// Response to RequestBackgroundColor() issued in Init(). Seeds the
		// initial light/dark detection before the first user interaction.
		a.onBackgroundChanged(msgWithType.IsDark())
		return a, a.RerenderCmd(true)
	case uv.LightColorSchemeEvent:
		a.onBackgroundChanged(false)
		return a, a.RerenderCmd(true)
	case uv.DarkColorSchemeEvent:
		a.onBackgroundChanged(true)
		return a, a.RerenderCmd(true)
	}

	// Notification mouse handling is checked before modal/page routing. Events
	// inside a notification are consumed; leaving an action clears its hover
	// state while allowing the event to continue to the modal or page.
	if mouseMsg, ok := msg.(tea.MouseMsg); ok {
		if notif := a.notificationAt(mouseMsg.Mouse()); notif != nil {
			result := notif.handleMouse(mouseMsg)
			if result.consumed {
				var cmds []tea.Cmd
				onAction := notif.spec.OnAction
				if result.dismiss {
					a.removeNotification(notif.id)
				}
				if result.dismiss || result.rerender {
					cmds = append(cmds, a.RerenderCmd(true))
				}
				if result.cmd != nil {
					cmds = append(cmds, result.cmd)
				}
				if result.action != nil && onAction != nil {
					onAction(*result.action)
				}
				if len(cmds) > 0 {
					return a, tea.Batch(cmds...)
				}
				return a, nil
			}
		} else if _, isMotion := mouseMsg.(tea.MouseMotionMsg); isMotion {
			hoverChanged := false
			for _, notification := range a.notifications {
				if notification.hoveredAction >= 0 {
					notification.hoveredAction = -1
					hoverChanged = true
				}
			}
			if hoverChanged {
				notificationHoverCmd = a.RerenderCmd(true)
			}
		}
	}

	// Modal input interception — only the topmost modal receives input.
	if len(a.modalStack) > 0 {
		top := a.modalStack[len(a.modalStack)-1]
		switch msg := msg.(type) {
		case tea.KeyMsg:
			top.update(msg)
			if top.dismissed() {
				page, cmd := a.completeTopModal()
				if page != nil {
					a.setPage(page)
				}
				cmds := []tea.Cmd{a.RerenderCmd(true)}
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return a, tea.Batch(cmds...)
			}
			return a, a.RerenderCmd(true)
		case tea.MouseMsg:
			handled, mouseCmd := top.handleMouse(msg)
			if handled {
				if top.dismissed() {
					page, actionCmd := a.completeTopModal()
					if page != nil {
						a.setPage(page)
					}
					cmds := []tea.Cmd{a.RerenderCmd(true)}
					if mouseCmd != nil {
						cmds = append(cmds, mouseCmd)
					}
					if actionCmd != nil {
						cmds = append(cmds, actionCmd)
					}
					return a, tea.Batch(cmds...)
				}
				cmds := []tea.Cmd{a.RerenderCmd(true)}
				if mouseCmd != nil {
					cmds = append(cmds, mouseCmd)
				}
				return a, tea.Batch(cmds...)
			}
			// Click outside the topmost modal: dismiss behavior depends on modal type.
			if _, isClick := msg.(tea.MouseClickMsg); !isClick {
				// Non-click (e.g. motion) outside the modal is not consumed, but a
				// returned pointer-reset command must still be honored.
				if mouseCmd != nil {
					return a, mouseCmd
				}
				return a, nil
			}
			mouse := msg.Mouse()
			if mouse.Button == tea.MouseLeft {
				// Left-click outside dismisses (and consumes the click) unless the
				// modal declines dismissal (e.g. DisableOutsideClick). Either way
				// the click is consumed so it never reaches the underlying page.
				if !top.dismissOutside() {
					return a, a.RerenderCmd(true)
				}
				page, cmd := a.completeTopModal()
				if page != nil {
					a.setPage(page)
				}
				cmds := []tea.Cmd{a.RerenderCmd(true)}
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return a, tea.Batch(cmds...)
			} else if mouse.Button == tea.MouseRight {
				// Right-click outside: dismiss if modal allows passthrough, then forward to page
				if top.allowsRightClickPassthrough() {
					top.dismissOutside()
					page, modalCmd := a.completeTopModal()
					if page != nil {
						a.setPage(page)
					}
					// Forward the right-click to the page
					newPage, pageCmd := a.page.Update(msg, a)
					if newPage != nil {
						a.setPage(newPage)
					}
					cmds := []tea.Cmd{a.RerenderCmd(true)}
					if modalCmd != nil {
						cmds = append(cmds, modalCmd)
					}
					if pageCmd != nil {
						cmds = append(cmds, pageCmd)
					}
					return a, tea.Batch(cmds...)
				}
			}
			return a, nil
		}
		// Forward non-input messages (ticks, etc.) to the page so it continues
		// updating while a modal is open.
	}

	page, cmd := a.page.Update(msg, a)
	if page != nil {
		a.setPage(page)
	}
	return a, cmd
}

func (a *App) View() tea.View {
	var v tea.View
	v.AltScreen = a.options.AltScreen
	v.MouseMode = a.options.MouseMode

	if a.quiting || a.WindowHeight() <= 0 || a.WindowWidth() <= 0 || a.page == nil {
		return v
	}

	baseContent := a.page.View(a)

	// Composite modals on top of the page content (if any).
	if len(a.modalStack) > 0 {
		baseContent = a.compositeModals(baseContent)
	}

	// Composite notifications on top of everything (page + modals).
	if len(a.notifications) > 0 {
		baseContent = a.compositeNotifications(baseContent)
	}

	v.SetContent(baseContent)
	return v
}

// resolveTheme selects the appropriate theme based on the configured options
// and current terminal background detection.
//
// Priority:
//  1. Both DarkTheme and LightTheme configured → auto-select based on detectedBg.
//  2. Neither configured → use DefaultTheme() (auto-adaptive).
func (a *App) resolveTheme() style.Theme {
	if len(a.options.ThemeList) > 0 {
		if a.themeIndex >= len(a.options.ThemeList) {
			a.themeIndex = 0
		}
		return a.options.ThemeList[a.themeIndex]
	}
	if a.options.DarkTheme.Primary != nil && a.options.LightTheme.Primary != nil {
		if style.HasDarkBackground() {
			return a.options.DarkTheme
		}
		return a.options.LightTheme
	}
	return style.DefaultTheme()
}

// onBackgroundChanged handles a detected change in terminal background
// color (light/dark). Updates the cached detection and rebuilds the
// StyleSet. Updates both the app-scoped StyleSet (if set) and the global
// StyleSet, because all current rendering code reads the global via
// style.CurrentStyleSet() (30+ call sites across model/).
//
// Also re-renders markdown popups to pick up the new auto-detected glamour style.
func (a *App) onBackgroundChanged(isDark bool) {
	style.SetDarkBackground(isDark)
	newStyleSet := style.NewStyleSet(a.resolveTheme())

	// Update app-scoped StyleSet if one was set via SetStyleSet
	if a.styleSet != nil {
		a.SetStyleSet(newStyleSet)
	}
	// Always update global StyleSet so existing rendering code sees the change
	style.SetStyleSet(newStyleSet)

	// Invalidate popup render cache: markdown popups rerender entirely;
	// plain-text popups clear contentLines so next render picks up new theme colors
	for _, modal := range a.modalStack {
		if popup, ok := modal.(*Popup); ok {
			if !popup.rerenderMarkdown() {
				// Plain-text popup: invalidate cached contentLines
				popup.contentLines = nil
			}
		}
	}
}

func (a *App) Run() error {
	util.PrimaryColor = a.options.PrimaryColor

	// Skip synchronous background detection — it can block for up to 2 seconds
	// and defaults to dark on failure, which incorrectly affects light terminals
	// with slow or unsupported OSC 11 queries.
	// Instead, rely entirely on the asynchronous BackgroundColorMsg issued in
	// Init(), which arrives within the first few frames and triggers a rerender.
	// The default detectedDarkBg=true acts as a safe fallback for the 1-2 frames
	// before BackgroundColorMsg arrives.

	// Initialize the StyleSet from the configured theme. Set both app-scoped
	// and global StyleSet because all current rendering code reads the global
	// via style.CurrentStyleSet() (30+ call sites across model/).
	themeStyleSet := style.NewStyleSet(a.resolveTheme())
	a.SetStyleSet(themeStyleSet)
	style.SetStyleSet(themeStyleSet)

	if a.page == nil {
		a.main = NewMain(a, a.options)
		a.startup = NewStartup(&a.options.StartupOptions, a.main)
		if a.options.InitPage == nil {
			a.options.InitPage = a.main
			if a.options.EnableStartup {
				a.options.InitPage = a.startup
			}
		}
		a.setPage(a.options.InitPage)
	}

	if len(a.options.GlobalKeyHandlers) > 0 {
		ListenGlobalKeys(a, a.options.GlobalKeyHandlers)
	}

	a.options.TeaOptions = append(a.options.TeaOptions, tea.WithHardTabs(false), tea.WithFoxfulRenderer())
	a.program = tea.NewProgram(a, a.options.TeaOptions...)
	_, err := a.program.Run()
	return err
}

func (a *App) Rerender(cleanScreen bool) {
	if a.program == nil {
		return
	}
	// Send in a goroutine to avoid blocking on the unbuffered msgs channel.
	// This is called from goroutines (e.g., ticker) and must not deadlock
	// when the event loop is busy or hasn't started yet.
	go func() {
		if cleanScreen {
			a.program.Send(tea.ClearScreen())
		}
		if p := a.getPage(); p != nil {
			a.program.Send(p.Msg())
		}
	}()
}

func (a *App) RerenderCmd(cleanScreen bool) tea.Cmd {
	return func() tea.Msg {
		if cleanScreen {
			a.program.Send(tea.ClearScreen())
		}
		p := a.getPage()
		if p == nil {
			return nil
		}
		return p.Msg()
	}
}

func (a *App) Tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		p := a.getPage()
		if p == nil {
			return nil
		}
		return p.Msg()
	})
}

func (a *App) WindowWidth() int {
	return a.windowWidth
}

func (a *App) WindowHeight() int {
	return a.windowHeight
}

func (a *App) CurPage() Page {
	return a.getPage()
}

// getPage 在读锁保护下返回当前页面，可从任意 goroutine 调用。
// 调用者必须在本方法返回后再调用页面方法（例如 Msg），不可持锁调用。
func (a *App) getPage() Page {
	a.pageMu.RLock()
	defer a.pageMu.RUnlock()
	return a.page
}

// setPage 在写锁保护下存储当前页面，只可由主事件循环 goroutine 调用。
func (a *App) setPage(p Page) {
	a.pageMu.Lock()
	a.page = p
	a.pageMu.Unlock()
}

func (a *App) Startup() *StartupPage {
	return a.startup
}

func (a *App) Main() *Main {
	return a.main
}

// MustMain returns the main page, panicking if it has not been initialized.
// It follows the Go Must* convention (cf. regexp.MustCompile): use it only
// when the main page is guaranteed to exist (i.e. after App.Run has set it up).
// Prefer the non-panicking Main() when nil is a possible, recoverable state.
func (a *App) MustMain() *Main {
	if a.main != nil {
		return a.main
	}
	panic("main page is empty")
}

// MustStartup returns the startup page, panicking if it has not been
// initialized. It follows the Go Must* convention (cf. regexp.MustCompile):
// use it only when the startup page is guaranteed to exist. Prefer the
// non-panicking Startup() when nil is a possible, recoverable state.
func (a *App) MustStartup() *StartupPage {
	if a.startup != nil {
		return a.startup
	}
	panic("startup page is empty")
}

func (a *App) Options() *Options {
	return a.options
}

// Quit gracefully shuts down the application. It calls Close() to run any
// registered close hooks, then sends the quit signal to the bubbletea program.
func (a *App) Quit() {
	a.Close()
	a.quiting = true
	if a.program != nil {
		a.program.Send(tea.Quit())
	}
}

// pushModal appends a modal to the stack. A nil modal is ignored rather than
// panicking, so a mis-wired caller degrades to a no-op instead of crashing the
// host application.
func (a *App) pushModal(m Modal) {
	if m == nil {
		return
	}
	a.modalStack = append(a.modalStack, m)
}

// ShowPopup pushes a validated popup onto the modal stack.
// The topmost modal receives input first. A nil popup is ignored, so callers
// can safely pass the result of a fallible NewPopup without a separate guard.
func (a *App) ShowPopup(p *Popup) {
	if p == nil {
		return
	}
	a.pushModal(p)
}

// DismissPopup dismisses the topmost modal on the stack.
// Does nothing if the stack is empty.
func (a *App) DismissPopup() {
	if len(a.modalStack) > 0 {
		a.modalStack = a.modalStack[:len(a.modalStack)-1]
	}
}

// completeTopModal pops the topmost modal, calls its complete() method,
// and returns the resulting (Page, tea.Cmd).
func (a *App) completeTopModal() (Page, tea.Cmd) {
	if len(a.modalStack) == 0 {
		return nil, nil
	}
	topIndex := len(a.modalStack) - 1
	top := a.modalStack[topIndex]
	a.modalStack = a.modalStack[:topIndex]
	return top.complete(a)
}

// HasPopup returns whether a modal (popup or context menu) is currently active.
func (a *App) HasPopup() bool {
	return len(a.modalStack) > 0
}

// TopModalBounds returns the absolute screen rectangle (x, y, width, height) of
// the topmost modal in the stack. The second return value is false when there is
// no active modal. Callers should treat a true-but-zero-area result as an unplaced
// modal whose bounds haven't been computed yet (first-frame race).
func (a *App) TopModalBounds() (x, y, w, h int, ok bool) {
	if len(a.modalStack) == 0 {
		return 0, 0, 0, 0, false
	}
	top := a.modalStack[len(a.modalStack)-1]
	switch m := top.(type) {
	case *Popup:
		x, y, w, h = m.Bounds()
		return x, y, w, h, true
	case *ContextMenu:
		x, y, w, h = m.Bounds()
		return x, y, w, h, true
	}
	return 0, 0, 0, 0, false
}

// compositeModals renders the base page content with all modals in the stack
// overlaid using lipgloss Compositor layers. Modals are rendered in stack
// order (bottom of stack = back layer, top of stack = front layer).
func (a *App) compositeModals(baseContent string) string {
	w, h := a.WindowWidth(), a.WindowHeight()
	ss := style.CurrentStyleSet()

	layers := []*layout.Layer{layout.NewLayer(baseContent)}
	for _, modal := range a.modalStack {
		// Type-switch to render Popup vs ContextMenu
		switch m := modal.(type) {
		case *Popup:
			rendered := m.render(ss.Popup)
			popupH := lipgloss.Height(rendered.content)
			popupW := layout.Width(rendered.content)
			x, y := m.computePosition(w, h, popupW, popupH)
			m.setBounds(x, y, popupW, popupH, rendered.actionBounds)
			m.SetTermSize(w, h)
			layers = append(layers, layout.NewLayer(rendered.content).X(x).Y(y))
		case *ContextMenu:
			rendered := m.renderModal(ss, w, h)
			menuH := lipgloss.Height(rendered.content)
			menuW := layout.Width(rendered.content)
			x, y := m.computePosition(w, h, menuW, menuH)
			m.setModalBounds(x, y, menuW, menuH, rendered.itemBounds)
			layers = append(layers, layout.NewLayer(rendered.content).X(x).Y(y))
		}
	}
	return layout.NewCompositor(layers...).Render()
}

// ---- Notification API ----

// Notify displays a notification and returns its ID. Interactive notifications
// with actions and a zero Timeout remain visible for user interaction. Other
// Info/Success notifications use the configured default timeout; Warning/Error
// notifications persist until dismissed manually.
//
// Safe to call from goroutines, including during Init(); internally sends a
// message to the Update loop via a non-blocking goroutine to avoid deadlocks
// when called before the event loop starts.
func (a *App) Notify(spec NotificationSpec) NotificationID {
	if a.program == nil {
		return 0
	}
	spec = cloneNotificationSpec(spec)
	// Assign ID optimistically for return (actual assignment happens in Update).
	// This is a heuristic; for guaranteed ID tracking, use the returned ID.
	nextID := a.nextNotificationID + 1
	go a.program.Send(ShowNotificationMsg{Spec: spec})
	return nextID
}

// UpdateNotification updates the content of an existing notification.
// Does nothing if the ID does not exist.
//
// Safe to call from goroutines, including during Init().
func (a *App) UpdateNotification(id NotificationID, spec NotificationSpec) {
	if a.program == nil {
		return
	}
	spec = cloneNotificationSpec(spec)
	go a.program.Send(updateNotificationMsg{id: id, spec: spec})
}

// DismissNotification dismisses a specific notification by ID.
// Does nothing if the ID does not exist.
//
// Safe to call from goroutines, including during Init().
func (a *App) DismissNotification(id NotificationID) {
	if a.program == nil {
		return
	}
	go a.program.Send(dismissNotificationMsg{id: id})
}

// ClearAllNotifications dismisses all visible notifications immediately.
//
// Safe to call from goroutines, including during Init().
func (a *App) ClearAllNotifications() {
	if a.program == nil {
		return
	}
	go a.program.Send(clearAllNotificationsMsg{})
}

// handleShowNotification creates a notification and returns a timeout Cmd if needed.
func (a *App) handleShowNotification(spec NotificationSpec) tea.Cmd {
	a.nextNotificationID++
	id := a.nextNotificationID
	spec = cloneNotificationSpec(spec)

	notif := &Notification{
		id:            id,
		spec:          spec,
		createdAt:     time.Now(),
		hoveredAction: -1,
	}
	a.notifications = append(a.notifications, notif)

	cmds := []tea.Cmd{a.RerenderCmd(true)}

	// Interactive notifications remain visible when no timeout is explicit.
	// Otherwise Info/Success use the configured default and Warning/Error persist.
	timeout := spec.Timeout
	if timeout == 0 && len(spec.Actions) == 0 {
		if spec.Level == NotificationInfo || spec.Level == NotificationSuccess {
			timeout = a.options.NotificationOptions.DefaultTimeout
			if timeout == 0 {
				timeout = 4 * time.Second
			}
		}
	}

	if timeout > 0 {
		notif.expireAt = time.Now().Add(timeout)
		cmds = append(cmds, tea.Tick(timeout, func(time.Time) tea.Msg {
			return notificationExpireMsg{id: id}
		}))
	}

	return tea.Batch(cmds...)
}

// removeNotification removes a notification by ID.
func (a *App) removeNotification(id NotificationID) {
	for i, n := range a.notifications {
		if n.id == id {
			a.notifications = append(a.notifications[:i], a.notifications[i+1:]...)
			return
		}
	}
}

// handleExpire checks expireAt before removing. When UpdateNotification clears
// or extends the timeout, stale ticks are ignored because expireAt was updated.
func (a *App) handleExpire(id NotificationID) {
	for _, n := range a.notifications {
		if n.id == id {
			if !n.expireAt.IsZero() && time.Now().After(n.expireAt) {
				a.removeNotification(id)
			}
			return
		}
	}
}

// updateNotificationContent updates the spec of an existing notification and
// manages the expiration timeout. For updates, spec.Timeout == 0 means "no
// auto-expire" (unlike initial creation which falls back to DefaultTimeout for
// Info/Success). Returns a tea.Tick if a new timeout should be set.
func (a *App) updateNotificationContent(id NotificationID, spec NotificationSpec) tea.Cmd {
	for _, n := range a.notifications {
		if n.id == id {
			n.spec = cloneNotificationSpec(spec)
			n.resetInteraction()
			// Update expiration: 0 means no timeout for updates.
			if spec.Timeout > 0 {
				n.expireAt = time.Now().Add(spec.Timeout)
				return tea.Tick(spec.Timeout, func(time.Time) tea.Msg {
					return notificationExpireMsg{id: id}
				})
			}
			n.expireAt = time.Time{} // clear expiration
			return nil
		}
	}
	return nil
}

// notificationAt returns the notification at the given mouse position,
// or nil if no notification is under the mouse.
func (a *App) notificationAt(mouse tea.Mouse) *Notification {
	for i := len(a.notifications) - 1; i >= 0; i-- {
		n := a.notifications[i]
		if n.boundsSet && n.bounds.contains(mouse.X, mouse.Y) {
			return n
		}
	}
	return nil
}

// compositeNotifications overlays all active notifications on top of the base content.
func (a *App) compositeNotifications(baseContent string) string {
	w, h := a.WindowWidth(), a.WindowHeight()
	ss := a.StyleSet()
	opts := a.options.NotificationOptions

	// Calculate effective max width (0 = min(termWidth/3, 60)).
	maxWidth := opts.MaxWidth
	if maxWidth == 0 {
		maxWidth = min(w/3, 60)
	}
	maxWidth = max(maxWidth, 20) // minimum sanity bound

	// Dynamic height limit: notifications occupy at most half the screen.
	maxTotalHeight := h / 2
	gap := opts.Gap

	layers := []*layout.Layer{layout.NewLayer(baseContent)}
	currentHeight := 0
	for _, n := range a.notifications {
		n.clearBounds()
	}

	// Render from newest (end of slice) to oldest, accumulating height.
	// Stop when we exceed the screen height limit.
	for i := len(a.notifications) - 1; i >= 0; i-- {
		n := a.notifications[i]
		rendered := a.renderNotification(n, ss.Notification, maxWidth, opts.MaxLines, maxTotalHeight-currentHeight)
		notifH := lipgloss.Height(rendered.content)
		notifW := layout.Width(rendered.content)

		if currentHeight+notifH > maxTotalHeight {
			for j := i; j >= 0; j-- {
				a.notifications[j].hoveredAction = -1
			}
			break // Oldest notifications are pushed out of view.
		}

		x, y := a.computeNotificationPosition(opts.Anchor, w, h, notifW, notifH, currentHeight, gap)
		n.setBounds(x, y, notifW, notifH, rendered.actionBounds, rendered.actionArea)

		layers = append(layers, layout.NewLayer(rendered.content).X(x).Y(y))
		currentHeight += notifH
	}

	return layout.NewCompositor(layers...).Render()
}

type notificationRender struct {
	content      string
	actionBounds []notificationRect
	actionArea   notificationRect
}

type notificationActionsRender struct {
	content string
	height  int
	bounds  []notificationRect
}

func renderNotificationActions(n *Notification, styles style.NotificationStyleSet, maxWidth int) notificationActionsRender {
	if len(n.spec.Actions) == 0 || maxWidth <= 0 {
		return notificationActionsRender{}
	}

	type actionRow struct {
		content strings.Builder
		width   int
	}

	rows := []actionRow{{}}
	bounds := make([]notificationRect, len(n.spec.Actions))
	for i, action := range n.spec.Actions {
		buttonStyle := styles.Action
		if i == n.hoveredAction {
			buttonStyle = styles.ActionHover
		}
		button := buttonStyle.Render(action.Label)
		buttonWidth := lipgloss.Width(button)
		if buttonWidth > maxWidth {
			button = lipgloss.NewStyle().MaxWidth(maxWidth).Render(button)
			buttonWidth = lipgloss.Width(button)
		}

		row := &rows[len(rows)-1]
		gap := 0
		if row.width > 0 {
			gap = 1
		}
		if row.width > 0 && row.width+gap+buttonWidth > maxWidth {
			rows = append(rows, actionRow{})
			row = &rows[len(rows)-1]
			gap = 0
		}
		if gap > 0 {
			row.content.WriteString(lipgloss.NewStyle().Background(styles.Surface).Render(" "))
		}
		bounds[i] = notificationRect{x: row.width + gap, y: len(rows) - 1, w: buttonWidth, h: 1}
		row.content.WriteString(button)
		row.width += gap + buttonWidth
	}

	rowTexts := make([]string, len(rows))
	for i, row := range rows {
		xOffset := maxWidth - row.width
		for j := range bounds {
			if bounds[j].y == i {
				bounds[j].x += xOffset
			}
		}
		rowTexts[i] = lipgloss.NewStyle().
			Width(maxWidth).
			Align(lipgloss.Right).
			Background(styles.Surface).
			Render(row.content.String())
	}

	return notificationActionsRender{
		content: strings.Join(rowTexts, "\n"),
		height:  len(rows),
		bounds:  bounds,
	}
}

// renderNotification renders a single notification with the given constraints.
func (a *App) renderNotification(n *Notification, styles style.NotificationStyleSet, maxWidth, maxLines, maxHeight int) notificationRender {
	spec := n.spec

	var frameStyle lipgloss.Style
	var icon string
	switch spec.Level {
	case NotificationInfo:
		frameStyle = styles.InfoFrame
		icon = styles.InfoIcon
	case NotificationSuccess:
		frameStyle = styles.SuccessFrame
		icon = styles.SuccessIcon
	case NotificationWarning:
		frameStyle = styles.WarningFrame
		icon = styles.WarningIcon
	case NotificationError:
		frameStyle = styles.ErrorFrame
		icon = styles.ErrorIcon
	}

	// Whole width includes two border and two padding columns.
	contentWidth := max(maxWidth-4, 10)
	actions := renderNotificationActions(n, styles, contentWidth)

	titleHeight := 0
	titleText := ""
	if spec.Title != "" {
		titleHeight = 1
		titleText = icon + spec.Title
	}

	actionSpacer := 0
	if actions.height > 0 && (spec.Title != "" || spec.Message != "") {
		actionSpacer = 1
	}
	bodyLimit := max(maxLines, 1)
	if maxHeight > 0 {
		bodyLimit = min(bodyLimit, max(0, maxHeight-2-titleHeight-actionSpacer-actions.height))
	}

	var bodyLines []string
	if spec.Message != "" && bodyLimit > 0 {
		wrapped := lipgloss.NewStyle().MaxWidth(contentWidth).Render(spec.Message)
		bodyLines = strings.Split(wrapped, "\n")
		if len(bodyLines) > bodyLimit {
			bodyLines = bodyLines[:bodyLimit]
			last := len(bodyLines) - 1
			bodyLines[last] = ansi.Truncate(bodyLines[last], contentWidth-1, "…")
		}
	}
	if titleHeight == 0 && len(bodyLines) == 0 {
		actionSpacer = 0
	}

	allLines := make([]string, 0, titleHeight+len(bodyLines))
	if titleHeight > 0 {
		allLines = append(allLines, titleText)
	}
	allLines = append(allLines, bodyLines...)
	n.setContentGeometry(allLines, contentWidth, titleHeight, titleText)

	displayLines := allLines
	if n.hasSelection && len(allLines) > 0 {
		displayLines = n.applySelectionHighlight(allLines, contentWidth)
	}

	var blocks []string
	remainingLines := displayLines
	if titleHeight > 0 && len(displayLines) > 0 {
		titleRunes := []rune(displayLines[0])
		effectiveWidth := max(contentWidth-2, 1)
		if len(titleRunes) > effectiveWidth {
			titleRunes = append(titleRunes[:effectiveWidth-1], '…')
		}
		titleRendered := styles.Title.Width(contentWidth - 2).Render(string(titleRunes))
		blocks = append(blocks, titleRendered+styles.Close.Render(" ✕"))
		remainingLines = displayLines[1:]
	}
	if len(remainingLines) > 0 {
		blocks = append(blocks, styles.Message.Width(contentWidth).Render(strings.Join(remainingLines, "\n")))
	}

	rendered := notificationRender{}
	if actions.height > 0 {
		if len(blocks) > 0 {
			blocks = append(blocks, lipgloss.NewStyle().Width(contentWidth).Background(styles.Surface).Render(""))
		}
		actionY := 1 + len(allLines) + actionSpacer
		for _, bound := range actions.bounds {
			rendered.actionBounds = append(rendered.actionBounds, notificationRect{
				x: 2 + bound.x,
				y: actionY + bound.y,
				w: bound.w,
				h: bound.h,
			})
		}
		rendered.actionArea = notificationRect{x: 2, y: actionY, w: contentWidth, h: actions.height}
		blocks = append(blocks, actions.content)
	}

	inner := lipgloss.JoinVertical(lipgloss.Left, blocks...)
	rendered.content = frameStyle.Render(inner)
	return rendered
}

// computeNotificationPosition calculates the (x, y) position for a notification
// based on anchor, terminal dimensions, notification dimensions, and stack offset.
func (a *App) computeNotificationPosition(anchor PopupAnchor, termW, termH, notifW, notifH, stackOffset, gap int) (int, int) {
	const margin = 1

	var x, y int

	// Horizontal positioning.
	switch anchor {
	case AnchorTopLeft, AnchorBottomLeft:
		x = margin
	case AnchorTopRight, AnchorBottomRight:
		x = termW - notifW - margin
	case AnchorTopCenter, AnchorBottomCenter, AnchorCenter:
		x = (termW - notifW) / 2
	default:
		x = (termW - notifW) / 2
	}

	// Vertical positioning with stack offset.
	switch anchor {
	case AnchorTopLeft, AnchorTopCenter, AnchorTopRight:
		// Stack downward from the top.
		y = margin + stackOffset
	case AnchorBottomLeft, AnchorBottomCenter, AnchorBottomRight:
		// Stack upward from the bottom. y is the top-left of the notification.
		y = termH - margin - stackOffset - notifH
	default: // AnchorCenter or custom
		y = (termH / 3) + stackOffset
	}

	// Clamp to screen bounds.
	x = max(0, min(x, termW-notifW))
	y = max(0, min(y, termH-notifH))

	return x, y
}
