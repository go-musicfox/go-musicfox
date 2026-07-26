package model

import (
	"fmt"
	"math"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/layout"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

type Main struct {
	options *Options

	app *App

	isDualColumn bool

	menuTitle *MenuItem

	menuStartRow     int
	menuStartColumn  int
	menuBottomRow    int
	menuListStartRow int // actual row where the first menu item renders

	menuCurPage  int
	menuPageSize int

	menuList      []MenuItem
	menuStack     *util.Stack
	selectedIndex int

	// local search
	inSearching bool
	searchInput textinput.Model

	loadingTips string // transient: set by MenuTips.DisplayTips, cleared by Recover

	// Deferred menu entry: instead of running the BeforeEnterMenuHook
	// synchronously (which blocks the Update cycle and prevents the
	// loadingTips from being rendered), the hook is deferred to the
	// next tick. This allows the current View() cycle to render the
	// loading text before the hook executes.
	pendingEnterMenu   *enterMenuDeferred
	pendingRefreshMenu *refreshMenuDeferred
	pendingMenuAction  *menuActionDeferred
	pendingAction      *actionDeferred // 通用延迟动作

	menu Menu // current menu

	components []Component

	kbCtrls    []KeyboardController
	mouseCtrls []MouseController

	statusBar StatusBar

	// Mouse click tracking for double-click detection
	lastClickTime time.Time
	lastClickX    int
	lastClickY    int

	// Mouse hover tracking for breadcrumb segments
	hoveredBreadcrumbIdx int // -1 = none, 0+ = display index in breadcrumbSegments

	// Mouse hover tracking for menu list items
	hoveredMenuItemIdx int // -1 = none, 0+ = index in menuList

	// Mouse hover tracking for tabs
	hoveredTabIdx int // -1 = none, 0+ = tab index

	// hoveredBackButton tracks whether the mouse is hovering over the back
	// button shown before the menu title when inside a submenu.
	hoveredBackButton bool

	// hoverPointerActive tracks whether the mouse is currently over a clickable
	// element. When true, the terminal mouse pointer is set to "pointer" (hand
	// cursor) via OSC 22. When false, it's reset to "default".
	hoverPointerActive bool

	// Multi-tab navigation (when Options.EnableTabs is true)
	tabs      *Tabs      // nil when EnableTabs is false
	activeTab int        // current tab index (0-based)
	tabStates []tabState // per-tab isolated state (parallel to Options.TabConfigs)
}

// tabState holds the per-tab navigation state that is saved/restored on tab switch.
// This allows each tab to maintain its own menu hierarchy, scroll position, and submenu stack.
type tabState struct {
	menu          Menu
	menuTitle     *MenuItem
	menuList      []MenuItem
	selectedIndex int
	menuCurPage   int
	menuStack     *util.Stack
}

type tickMainMsg struct{}

// enterMenuDeferred holds the state for a deferred submenu entry.
// Instead of executing BeforeEnterMenuHook synchronously (which blocks the
// Update cycle), the hook is deferred to the next tick. This gives the
// View() cycle a chance to render loadingTips before the hook runs.
type enterMenuDeferred struct {
	newMenu   Menu
	newTitle  *MenuItem
	loading   *Loading
	stackItem *menuStackItem
}

// refreshMenuDeferred 保存延迟刷新当前菜单所需的状态。
type refreshMenuDeferred struct {
	menu    Menu
	loading *Loading
}

// actionDeferred 保存通用延迟动作的状态。
type actionDeferred struct {
	action  func(*Main) (bool, Page) // 延迟执行的逻辑，返回（是否继续、目标页面）
	loading *Loading                 // 加载提示控制器
}

// menuActionDeferred holds an activated menu item until its loading state has rendered.
type menuActionDeferred struct {
	menu    Menu
	index   int
	item    MenuItem
	loading *Loading
}

func NewMain(app *App, options *Options) (m *Main) {
	var mainMenuTitle *MenuItem
	if options.MainMenuTitle != nil {
		mainMenuTitle = options.MainMenuTitle
	} else {
		mainMenuTitle = &MenuItem{Title: options.AppName}
	}

	m = &Main{
		app:                  app,
		options:              options,
		menuTitle:            mainMenuTitle,
		menu:                 options.MainMenu,
		menuStack:            &util.Stack{},
		menuCurPage:          1,
		menuPageSize:         10,
		searchInput:          textinput.New(),
		components:           options.Components,
		kbCtrls:              options.KBControllers,
		mouseCtrls:           options.MouseControllers,
		statusBar:            options.StatusBar,
		hoveredBreadcrumbIdx: -1,
		hoveredMenuItemIdx:   -1,
		hoveredTabIdx:        -1,
		hoveredBackButton:    false,
		hoverPointerActive:   false,
	}

	// Initialize multi-tab navigation if enabled
	if options.EnableTabs && len(options.TabConfigs) > 0 {
		// Create Tabs widget with titles extracted from TabConfigs
		titles := make([]string, len(options.TabConfigs))
		for i, cfg := range options.TabConfigs {
			titles[i] = cfg.Title
		}
		m.tabs = NewTabs(titles)
		m.tabs.SetBorder(true)
		m.tabs.Focus()

		// Initialize per-tab state snapshots
		m.tabStates = make([]tabState, len(options.TabConfigs))
		for i, cfg := range options.TabConfigs {
			m.tabStates[i] = tabState{
				menu:          cfg.Menu,
				menuTitle:     cfg.MenuTitle,
				menuList:      cfg.Menu.MenuViews(),
				selectedIndex: 0,
				menuCurPage:   1,
				menuStack:     &util.Stack{},
			}
		}

		// Load initial tab (tab 0)
		m.activeTab = 0
		firstState := m.tabStates[0]
		m.menu = firstState.menu
		m.menuTitle = firstState.menuTitle
		m.menuList = firstState.menuList
		m.menuStack = firstState.menuStack
		// selectedIndex, menuCurPage already default to 0/1
	} else {
		// Standard single-menu mode (EnableTabs=false or no TabConfigs)
		m.menuList = m.menu.MenuViews()
	}

	m.searchInput.Placeholder = " " + SearchPlaceholder
	m.searchInput.Prompt = util.GetFocusedPrompt()
	s := textinput.DefaultStyles(true)
	s.Focused.Text = util.GetPrimaryFontStyle(true)
	m.searchInput.SetStyles(s)
	m.searchInput.CharLimit = 32

	return
}

func (m *Main) RefreshMenuList() {
	m.menuList = m.menu.MenuViews()
}

func (m *Main) RefreshMenuTitle() {
	m.menu.FormatMenuItem(m.menuTitle)
}

func (m *Main) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return m.inSearching
}

func (m *Main) Type() PageType {
	return PtMain
}

func (m *Main) Msg() tea.Msg {
	return tickMainMsg{}
}

func (m *Main) Init(a *App) tea.Cmd {
	return a.Tick(time.Nanosecond)
}

func (m *Main) computeTitleStartRow() int {
	titleStartRow := 0
	if m.options.WhetherDisplayTitle && m.menuStartRow > 2 {
		if m.menuStartRow > 4 {
			titleStartRow = m.menuStartRow - 3
		} else {
			titleStartRow = 2
		}
	} else if !m.options.WhetherDisplayTitle && m.menuStartRow > 1 {
		if m.menuStartRow > 3 {
			titleStartRow = m.menuStartRow - 3
		} else {
			titleStartRow = 2
		}
	}
	return titleStartRow
}

func (m *Main) Update(msg tea.Msg, a *App) (Page, tea.Cmd) {
	// Deliver every message to registered components after primary handling.
	// Components read from the original tea.Msg and mutate their own state.
	defer m.updateComponents(msg, a)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.keyMsgHandle(msg, a)
	case tea.MouseMsg:
		return m.mouseMsgHandle(msg, a)
	case tickMainMsg:
		// Priority 1: invoke the selected item's Action after rendering its loading state.
		if m.pendingMenuAction != nil {
			p := m.pendingMenuAction
			m.pendingMenuAction = nil

			actionPage, actionCmd := p.menu.Action(m.app, p.index)
			if actionPage != nil || actionCmd != nil {
				p.loading.Complete()
				return actionPage, actionCmd
			}

			submenu := p.menu.SubMenu(m.app, p.index)
			if submenu == nil {
				p.loading.Complete()
				return m, a.RerenderCmd(true)
			}

			p.loading.Complete()
			newPage := m.enterMenuWithLoading(submenu, &p.item)
			if m.pendingEnterMenu != nil {
				return m, a.RerenderCmd(true)
			}
			if newPage != nil {
				return newPage, func() tea.Msg { return newPage.Msg() }
			}
			return m, a.RerenderCmd(true)
		}

		// Priority 2: run a generic deferred action.
		if m.pendingAction != nil {
			p := m.pendingAction
			m.pendingAction = nil

			shouldContinue, newPage := p.action(m)
			p.loading.Complete()

			if !shouldContinue && newPage != nil {
				return newPage, func() tea.Msg { return newPage.Msg() }
			}
			return m, a.RerenderCmd(true)
		}
		if m.pendingRefreshMenu != nil {
			p := m.pendingRefreshMenu
			m.pendingRefreshMenu = nil

			if hook := p.menu.BeforeEnterMenuHook(); hook != nil {
				if res, newPage := hook(m); !res {
					p.loading.Complete()
					if newPage != nil {
						return newPage, func() tea.Msg { return newPage.Msg() }
					}
					return m, a.RerenderCmd(true)
				}
			}
			m.menuList = p.menu.MenuViews()
			p.loading.Complete()
			return m, a.RerenderCmd(true)
		}
		if m.pendingEnterMenu != nil {
			p := m.pendingEnterMenu
			m.pendingEnterMenu = nil

			var res bool
			var newPage Page
			if hook := p.newMenu.BeforeEnterMenuHook(); hook != nil {
				if res, newPage = hook(m); !res {
					p.loading.Complete()
					m.menuStack.Pop()
					if newPage != nil {
						return newPage, func() tea.Msg { return newPage.Msg() }
					}
					return m, nil
				}
			}
			p.loading.Complete()

			if p.newMenu != nil {
				p.newMenu.FormatMenuItem(p.newTitle)
			}
			m.hoveredMenuItemIdx = -1
			menuList := p.newMenu.MenuViews()
			m.menu = p.newMenu
			m.menuList = menuList
			m.menuTitle = p.newTitle
			m.selectedIndex = 0
			m.menuCurPage = 1

			if newPage != nil {
				return newPage, func() tea.Msg { return newPage.Msg() }
			}
			return m, a.RerenderCmd(true)
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.isDualColumn = msg.Width >= 75 && m.options.DualColumn
		m.menuStartRow = msg.Height / 3
		if m.options.MaxMenuStartRow > 0 {
			if m.menuStartRow > m.options.MaxMenuStartRow {
				m.menuStartRow = m.options.MaxMenuStartRow
			}
		}
		if !m.options.WhetherDisplayTitle && m.menuStartRow > 1 {
			m.menuStartRow--
		}
		// When status bar is at top, it replaces the title bar (forced override).
		if m.statusBar != nil && m.options.StatusBarPosition == StatusBarTop {
			m.options.WhetherDisplayTitle = false
		}
		if m.isDualColumn {
			switch {
			case msg.Width < 100:
				m.menuStartColumn = msg.Width / 5
			case msg.Width < 150:
				m.menuStartColumn = msg.Width / 4
			default:
				m.menuStartColumn = msg.Width / 3
			}
		} else {
			if msg.Width < 100 {
				m.menuStartColumn = msg.Width / 3
			} else {
				m.menuStartColumn = msg.Width * 2 / 5
			}
		}
		if m.menuStartColumn < 5 {
			m.menuStartColumn = 5
		}

		bottomHeight := 13
		if m.options.BottomHeight > 0 {
			bottomHeight = m.options.BottomHeight
		}
		// Reserve space for status bar (DefaultStatusBar = 1 row).
		if m.statusBar != nil {
			bottomHeight++
		}
		if m.options.DynamicRowCount {
			maxEntries := (msg.Height - m.menuStartRow - bottomHeight) * m.getNumColumns()
			if maxEntries > 10 {
				m.menuPageSize = maxEntries
			} else {
				m.menuPageSize = 10
			}
		} else {
			m.menuPageSize = 10
		}

		// Compute actual menu list start row and bottom row for hit-testing
		titleStartRow := m.computeTitleStartRow()

		// Leading rows before the menu list:
		// - Title bar (if displayed): 1 row
		// - Status bar at top (if configured): 1 row (replaces title bar, same layout effect)
		// - Filler blank lines: strings.Repeat("\n", titleStartRow-1) → getLines produces titleStartRow parts
		// - Menu title: 1 row
		// - Gap "": empty string renders as 1 visual row in JoinVertical
		leadingRows := 0
		if m.options.WhetherDisplayTitle {
			leadingRows++
		}
		// When status bar is at top, it occupies row 0 like the title bar does,
		// so leadingRows must account for it even though WhetherDisplayTitle is forced false.
		if m.statusBar != nil && m.options.StatusBarPosition == StatusBarTop {
			leadingRows++
		}
		if titleStartRow > 1 {
			leadingRows += titleStartRow
		}
		leadingRows += 2 // menu title (1) + gap "" (1 visual row, see line 429)

		m.menuListStartRow = leadingRows

		menuDisplayLines := m.menuPageSize
		if m.isDualColumn {
			menuDisplayLines = int(math.Ceil(float64(m.menuPageSize) / 2))
		}
		m.menuBottomRow = m.menuListStartRow + menuDisplayLines

		if m.menuCurPage > 0 {
			maxPage := int(math.Ceil(float64(len(m.menuList)) / float64(m.menuPageSize)))
			if m.menuCurPage > maxPage {
				m.menuCurPage = maxPage
			}
		}
		return m, a.RerenderCmd(true)
	}

	return m, nil
}

// updateComponents dispatches a message to every registered component so they
// can maintain their own internal state. Components render below the menu list
// (e.g. spectrum, lyrics). Nil entries are skipped.
func (m *Main) updateComponents(msg tea.Msg, a *App) {
	for _, component := range m.components {
		if component == nil {
			continue
		}
		component.Update(msg, a)
	}
}

func (m *Main) View(a *App) string {
	w, h := a.WindowWidth(), a.WindowHeight()
	if w <= 0 || h <= 0 {
		return ""
	}

	var sections []string

	// ── 1. Top bar: status bar (when position=top) OR title bar ──
	if m.statusBar != nil && m.options.StatusBarPosition == StatusBarTop {
		statusBarView := m.statusBar.View(a, m)
		sections = append(sections, statusBarView)
	} else if m.options.WhetherDisplayTitle {
		sections = append(sections, m.TitleView(a))
	}

	// ── 2. Tab bar (if multi-tab mode enabled) ──
	if m.options.EnableTabs && m.tabs != nil {
		// Update tabs widget size to match window width
		m.tabs.SetSize(w, 0) // height auto-calculated by tabs widget
		m.tabs.SetHovered(m.hoveredTabIdx)
		sections = append(sections, m.tabs.View())
	}

	// ── 3. Menu sections ──
	if !m.options.HideMenu {
		titleStartRow := m.computeTitleStartRow()

		// Inject loading tips into a copy of the menu title
		mt := m.menuTitle
		if m.loadingTips != "" {
			tmp := *mt
			if tmp.Subtitle != "" {
				tmp.Subtitle = tmp.Subtitle + " " + m.loadingTips
			} else {
				tmp.Subtitle = m.loadingTips
			}
			mt = &tmp
		}

		// When tabs are enabled and in a submenu, show breadcrumb trail
		if m.options.EnableTabs && m.menuStack.Len() > 0 {
			mt = m.buildBreadcrumb()
			// Preserve loading tips if present
			if m.loadingTips != "" {
				tmp := *mt
				if tmp.Subtitle != "" {
					tmp.Subtitle = tmp.Subtitle + " " + m.loadingTips
				} else {
					tmp.Subtitle = m.loadingTips
				}
				mt = &tmp
			}
		}

		// Vertical gap to menu title row.
		if titleStartRow > 1 {
			sections = append(sections, strings.Repeat("\n", max(0, titleStartRow-1)))
		}
		sections = append(sections, m.menuTitleViewContent(a, mt))

		// Vertical gap: title row → menu start row (empty string = 1 visual row in JoinVertical)
		sections = append(sections, "")
		sections = append(sections, m.menuListView(a))

		// Only append searchInput if non-empty (matches component loop pattern at line 442).
		// When not searching and menu has no HelpHints, searchInputView returns "" which
		// JoinVertical counts as 1 phantom row, causing +1 overflow and title truncation.
		if searchView := m.searchInputView(a); searchView != "" {
			sections = append(sections, searchView)
		}
	} else {
		sections = append(sections, "\n\n\n")
	}

	// ── 3. Components (natural flow) ──
	for _, component := range m.components {
		if component == nil {
			continue
		}
		view, _ := component.View(a, m)
		if view != "" {
			sections = append(sections, view)
		}
	}

	// ── 4. Compose vertically ──
	body := layout.JoinVertical(lipgloss.Left, sections...)

	// ── 5. Status bar at bottom ──
	statusBarView := ""
	statusBarH := 0
	if m.statusBar != nil && m.options.StatusBarPosition == StatusBarBottom {
		statusBarView = m.statusBar.View(a, m)
		statusBarH = lipgloss.Height(statusBarView)
	}

	// ── 6. Adjust body height for status bar ──
	// Components use a.WindowHeight() which doesn't account for the status bar.
	// Trim body to targetHeight to prevent overflow when status bar is present.
	targetHeight := h - statusBarH
	bodyHeight := lipgloss.Height(body)
	if bodyHeight > targetHeight {
		lines := strings.Split(body, "\n")
		if len(lines) > targetHeight {
			body = strings.Join(lines[:targetHeight], "\n")
		}
	} else if bodyHeight < targetHeight {
		body = lipgloss.NewStyle().Height(targetHeight).Render(body)
	}

	// Combine body + status bar, then wrap with AppBackground.
	// AppBackground is transparent by default (terminal bg shows through).
	ss := style.CurrentStyleSet()
	var content string
	if statusBarView != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, body, statusBarView)
	} else {
		content = body
	}
	return ss.AppBackground.Width(w).Render(content)
}

// MenuTitleStartColumn returns the horizontal column where the menu title starts.
func (m *Main) MenuTitleStartColumn() int {
	return m.menuStartColumn
}

// MenuTitleStartRow returns the row where the menu title starts.
// Computed dynamically in View() since it was removed as a struct field.
// Returns 0 as a sensible default; downstream should use lipgloss layout instead.
func (m *Main) MenuTitleStartRow() int {
	return 0
}

// MenuStartColumn returns the horizontal column where menu items start.
func (m *Main) MenuStartColumn() int {
	return m.menuStartColumn
}

func (m *Main) MenuStartRow() int {
	return m.menuListStartRow
}

func (m *Main) MenuBottomRow() int {
	return m.menuBottomRow
}

func (m *Main) StatusBar() StatusBar {
	return m.statusBar
}

func (m *Main) StatusBarPosition() StatusBarPosition {
	return m.options.StatusBarPosition
}

// EffectiveWindowHeight returns the available height for page content
// excluding the status bar (if present). Components should use this for
// layout calculations instead of a.WindowHeight() when positioning content
// that should not overlap the status bar.
func (m *Main) EffectiveWindowHeight(a *App) int {
	h := a.WindowHeight()
	if m.statusBar == nil {
		return h
	}

	// Status bar at bottom always consumes 1 extra row
	if m.options.StatusBarPosition == StatusBarBottom {
		return h - 1
	}

	// Status bar at top occupies row 0 but leaves the bottom row free.
	// Whether it replaces a title bar or sits alone doesn't affect the
	// bottom boundary — content can always extend to row h-1.
	return h
}

// statusBarRowY returns the 0-based screen row where the status bar is rendered.
// Returns -1 if no status bar is configured.
func (m *Main) statusBarRowY(a *App) int {
	if m.statusBar == nil {
		return -1
	}

	switch m.options.StatusBarPosition {
	case StatusBarTop:
		return 0 // Top row
	case StatusBarBottom:
		return a.WindowHeight() - 1 // Last row
	default:
		return -1
	}
}

func (m *Main) IsDualColumn() bool {
	return m.isDualColumn
}

func (m *Main) CenterEverything() bool {
	return m.options.CenterEverything
}

func (m *Main) MenuTitle() *MenuItem {
	return m.menuTitle
}

func (m *Main) CurMenu() Menu {
	return m.menu
}

func (m *Main) CurPage() int {
	return m.menuCurPage
}

func (m *Main) PageSize() int {
	return m.menuPageSize
}

func (m *Main) SelectedIndex() int {
	return m.selectedIndex
}

func (m *Main) SetSelectedIndex(i int) {
	m.selectedIndex = i
}

// switchTab switches to the specified tab index, saving current tab state and
// restoring the target tab's state. Calls TabConfig.OnActivate hook if defined.
func (m *Main) switchTab(newIndex int) {
	if newIndex == m.activeTab || newIndex < 0 || newIndex >= len(m.tabStates) {
		return
	}

	prevIndex := m.activeTab

	// 1. Save current tab's state
	m.tabStates[m.activeTab] = tabState{
		menu:          m.menu,
		menuTitle:     m.menuTitle,
		menuList:      m.menuList,
		selectedIndex: m.selectedIndex,
		menuCurPage:   m.menuCurPage,
		menuStack:     m.menuStack.DeepCopy(),
	}

	// 2. Call OnActivate hook (if defined) — can veto the switch
	if m.options.TabConfigs[newIndex].OnActivate != nil {
		if !m.options.TabConfigs[newIndex].OnActivate(m, prevIndex) {
			return // Hook vetoed the switch
		}
	}

	// 3. Restore new tab's state
	m.activeTab = newIndex
	state := m.tabStates[newIndex]
	m.menu = state.menu
	m.menuTitle = state.menuTitle
	m.menuList = state.menuList
	m.selectedIndex = state.selectedIndex
	m.menuCurPage = state.menuCurPage
	m.menuStack = state.menuStack

	// 4. Update Tabs widget active index
	m.tabs.SetActive(newIndex)

	// 5. Clear transient state
	m.inSearching = false
	m.searchInput.Reset()
	m.searchInput.Blur()
	m.hoveredMenuItemIdx = -1
	m.hoveredBreadcrumbIdx = -1
}

// buildBreadcrumb constructs a breadcrumb trail from the current tab title
// and menu stack hierarchy. Used when EnableTabs=true to show navigation context.
func (m *Main) buildBreadcrumb() *MenuItem {
	if !m.options.EnableTabs || m.tabs == nil {
		return m.menuTitle
	}

	var parts []string

	// Start with active tab title
	if m.activeTab >= 0 && m.activeTab < len(m.options.TabConfigs) {
		parts = append(parts, m.options.TabConfigs[m.activeTab].Title)
	}

	// Add menu stack path (each menuStackItem's title)
	if m.menuStack.Len() > 0 {
		stackItems := m.menuStack.ToSlice()
		for _, item := range stackItems {
			if si, ok := item.(*menuStackItem); ok && si.menuTitle != nil {
				parts = append(parts, si.menuTitle.Title)
			}
		}
	}

	// Add current menu title
	if m.menuTitle != nil && m.menuTitle.Title != "" {
		parts = append(parts, m.menuTitle.Title)
	}

	if len(parts) == 0 {
		return &MenuItem{Title: ""}
	}

	// Join with " > " separator
	breadcrumb := strings.Join(parts, " > ")

	// Truncate if too long (optional: could implement smart truncation here)
	return &MenuItem{Title: breadcrumb, Subtitle: ""}
}

// tabIndexAt returns the tab index at the given mouse coordinates, or -1 if not over any tab.
// Calculates the horizontal layout of tabs (with borders and padding) and checks if the click
// falls within a tab's rendered region.
func (m *Main) tabIndexAt(x, y int, a *App) int {
	if !m.options.EnableTabs || m.tabs == nil || len(m.tabStates) == 0 {
		return -1
	}

	// Calculate tab bar starting row in the view
	// Rendering order from View(): [Title bar (optional)] + [Tab bar] + [vertical gap] + [menu title] + ...
	tabBarStartRow := 0
	if m.options.WhetherDisplayTitle {
		tabBarStartRow++ // Title bar takes 1 row
	}
	if m.statusBar != nil && m.options.StatusBarPosition == StatusBarTop {
		tabBarStartRow++
	}

	// Tab bar with borders occupies ~3 rows (top border + content + bottom border)
	// Check if mouse Y is within tab bar region
	tabBarHeight := 3 // conservative estimate for bordered tabs
	if y < tabBarStartRow || y >= tabBarStartRow+tabBarHeight {
		return -1
	}

	// Calculate horizontal position of each tab to determine which was clicked
	// Match the rendering logic from tabs.go renderTabBar()
	ss := style.CurrentStyleSet()

	activeTabBorder := lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	borderColor := ss.SelectedItem.GetForeground()
	if !m.tabs.Focused() {
		borderColor = ss.Border.GetForeground()
	}

	tab := lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(borderColor).
		Padding(0, 1)

	activeTab := lipgloss.NewStyle().
		Border(activeTabBorder, true).
		BorderForeground(borderColor).
		Padding(0, 1)

	// Calculate cumulative width for each tab to determine click boundaries
	currentX := 0
	for i := 0; i < len(m.options.TabConfigs); i++ {
		title := m.options.TabConfigs[i].Title

		var tabContent string
		var renderedTab string
		if i == m.activeTab {
			tabContent = ss.SelectedItem.Render(title)
			renderedTab = activeTab.Render(tabContent)
		} else {
			tabContent = ss.MenuItem.Render(title)
			renderedTab = tab.Render(tabContent)
		}

		tabWidth := lipgloss.Width(renderedTab)

		// Check if click falls within this tab's horizontal range
		if x >= currentX && x < currentX+tabWidth {
			return i
		}

		currentX += tabWidth
	}

	return -1
}

// TitleView renders the app name as a decorative bar with dashes on both sides.
func (m *Main) TitleView(a *App) string {
	appName := " " + m.options.AppName + " "
	w := a.WindowWidth()
	titleLen := layout.Width(appName)
	prefixLen := (w - titleLen) / 2
	suffixLen := w - prefixLen - titleLen

	var b strings.Builder
	if prefixLen > 0 {
		b.WriteString(strings.Repeat("─", prefixLen))
	}
	b.WriteString(appName)
	if suffixLen > 0 {
		b.WriteString(strings.Repeat("─", suffixLen))
	}
	return style.CurrentStyleSet().Title.Render(b.String())
}

// backButtonIcon returns the styled back button icon suitable for prepending
// to the menu title when inside a submenu.
func (m *Main) backButtonIcon() string {
	ss := style.CurrentStyleSet()
	if m.hoveredBackButton {
		return ss.BackButtonHover.Render("←")
	}
	return ss.BackButton.Render("←")
}

// menuTitleViewContent renders the menu title content string, left-aligned
// at menuStartColumn to match the menu items' horizontal position.
// The loading tips should be injected by the caller (see View).
func (m *Main) menuTitleViewContent(a *App, menuTitle *MenuItem) string {
	if menuTitle == nil {
		menuTitle = m.menuTitle
	}
	windowWidth := a.WindowWidth()
	startCol := m.menuStartColumn
	ss := style.CurrentStyleSet()

	// When in a submenu, show a back button to the left of the title.
	// The back button is positioned at startCol - backButtonWidth so the
	// title itself remains at its original startCol position.
	showBack := m.menuStack.Len() > 0

	maxLen := windowWidth - startCol
	realString := menuTitle.OriginString()
	formatString := menuTitle.String()

	var titleText string
	if lipgloss.Width(realString) > maxLen {
		// Truncate long titles: prioritize title, clip subtitle if needed
		tmp := *menuTitle
		titleLen := lipgloss.Width(tmp.Title)
		subTitleLen := lipgloss.Width(tmp.Subtitle)
		if titleLen >= maxLen-1 {
			tmp.Title = lipgloss.NewStyle().Width(maxLen - 1).MaxWidth(maxLen - 1).Render(tmp.Title)
			tmp.Subtitle = ""
		} else if subTitleLen >= maxLen-titleLen-1 {
			tmp.Subtitle = lipgloss.NewStyle().Width(maxLen - titleLen - 1).MaxWidth(maxLen - titleLen - 1).Render(tmp.Subtitle)
		}
		titleText = tmp.String()
	} else {
		titleText = lipgloss.NewStyle().Width(maxLen).Render(formatString)
	}

	// Style the title independently — back button must NOT affect its color.
	styledTitle := ss.MenuTitle.Render(titleText)

	if showBack {
		// Back button at startCol - backButtonWidth, title unchanged at startCol.
		// Layout: [padding]←[space][title...]
		backIcon := m.backButtonIcon()
		padding := startCol - backButtonWidth
		if padding < 0 {
			padding = 0
		}
		return strings.Repeat(" ", padding) + backIcon + " " + styledTitle
	}

	// No back button: original padding + title
	if startCol > 0 {
		styledTitle = lipgloss.NewStyle().PaddingLeft(startCol).Render(styledTitle)
	}
	return styledTitle
}

// MenuTitleView menu title
func (m *Main) MenuTitleView(a *App) string {
	return m.menuTitleViewContent(a, m.menuTitle)
}

func (m *Main) MenuList() []MenuItem {
	return m.menuList
}

func (m *Main) getNumColumns() int {
	if m.isDualColumn {
		return 2
	}
	return 1
}

func (m *Main) forceEntryLength(item *MenuItem, targetLength int) string {
	// Case 1:
	// Only enough space for the main title. Not enough width for subtitle.
	titleWidth := layout.Width(item.Title)
	minSubtitleWidth := 5
	if titleWidth >= targetLength-minSubtitleWidth {
		return lipgloss.NewStyle().
			Width(targetLength).
			Render(item.Title)
	}
	// Case 2:
	// Enough space for everything.
	fullWidth := layout.Width(item.OriginString())
	if fullWidth <= targetLength {
		return lipgloss.NewStyle().Width(targetLength).Render(item.OriginString())
	}
	// Case 3:
	// Enough space for main title. Need to scroll subtitle.
	subtitleSpace := targetLength - titleWidth - 1
	// Need 2 extra spaces for visual separation between end of subtitle and beginning.
	r := []rune(item.Subtitle + "  ")
	s := make([]rune, 0, subtitleSpace)
	indexStart := 0
	if m.options.Ticker != nil {
		indexStart = int(m.options.Ticker.PassedTime().Milliseconds() / 500 % int64(len(r)))
	}
	currentWidth := 0
	for i := indexStart; currentWidth < subtitleSpace; i = (i + 1) % len(r) {
		rw := layout.Width(string(r[i]))
		if currentWidth+rw > subtitleSpace {
			break
		}
		s = append(s, r[i])
		currentWidth += rw
	}
	subtitle := lipgloss.NewStyle().Width(subtitleSpace).MaxWidth(subtitleSpace).Render(string(s))
	return item.Title + " " + style.CurrentStyleSet().Subtitle.Render(subtitle)
}

func (m *Main) formatEntry(item *MenuItem, index int, targetLength int) string {
	if item == nil {
		return lipgloss.NewStyle().Width(targetLength).Render("")
	}
	var fmtStart string
	if !m.inSearching && index == m.selectedIndex {
		fmtStart = " => "
	} else {
		fmtStart = "    "
	}
	titleLength := targetLength - m.getMaxIndexWidth() - 6
	songEntry := fmt.Sprintf(
		fmt.Sprintf("%s%%%dd. %%s", fmtStart, m.getMaxIndexWidth()),
		index,
		m.forceEntryLength(item, titleLength))
	if m.isSelected(index) {
		return style.CurrentStyleSet().SelectedItem.Render(songEntry)
	}
	return songEntry
}

func (m *Main) centeredMenuView(a *App, lines int) string {
	var allSongs []*MenuItem
	startIndex := m.getPageStartIndex()
	endIndex := startIndex + lines
	if m.isDualColumn {
		endIndex = startIndex + lines*2
	}
	var titleLengths []int
	for i := startIndex; i < endIndex; i++ {
		if i < len(m.menuList) {
			menuItem := m.menuList[i]
			length := layout.Width(menuItem.OriginString())
			titleLengths = append(titleLengths, length)
			allSongs = append(allSongs, &menuItem)
		} else {
			allSongs = append(allSongs, nil)
		}
	}
	allSongs = append(allSongs, nil)

	entryLength := m.centeredEntryLength(a, titleLengths)

	var rows []string
	for i := 0; i < lines; i++ {
		index := i * m.getNumColumns()
		menuIndex := m.getPageStartIndex() + index
		left := m.formatEntry(allSongs[index], menuIndex, entryLength)
		if m.isDualColumn {
			right := m.formatEntry(allSongs[index+1], menuIndex+1, entryLength)
			row := layout.JoinHorizontal(lipgloss.Center, left, right)
			rows = append(rows, lipgloss.NewStyle().Width(a.windowWidth).Align(lipgloss.Center).Render(row))
		} else {
			rows = append(rows, lipgloss.NewStyle().Width(a.windowWidth).Align(lipgloss.Center).Render(left))
		}
	}
	return layout.JoinVertical(lipgloss.Left, rows...)
}

func (m *Main) centeredEntryLength(a *App, titleLengths []int) int {
	slices.Sort(titleLengths)

	maxTitleWidth := 0
	if len(titleLengths) > 0 {
		maxTitleWidth = titleLengths[len(titleLengths)-1]
	}
	if len(titleLengths) >= 6 && maxTitleWidth >= 30 {
		maxTitleWidth = titleLengths[int32(0.7*float32(len(titleLengths)))]
		if maxTitleWidth < 30 {
			maxTitleWidth = 30
		}
	}

	remainingWidth := a.windowWidth - 4
	extraPadding := (a.windowWidth - 40) / 5
	if extraPadding > 0 {
		remainingWidth -= extraPadding
	}

	itemMaxLength := remainingWidth / m.getNumColumns()
	entryLength := maxTitleWidth + 6 + m.getMaxIndexWidth()
	if entryLength > itemMaxLength {
		return itemMaxLength
	}
	return entryLength
}

func (m *Main) menuListView(a *App) string {
	var menuListBuilder strings.Builder
	if m.options.DynamicRowCount {
		m.menuCurPage = m.selectedIndex/m.menuPageSize + 1
	}
	menus := m.getCurPageMenus()
	var lines, maxLines int
	if m.isDualColumn {
		lines = int(math.Ceil(float64(len(menus)) / 2))
		maxLines = int(math.Ceil(float64(m.menuPageSize) / 2))
	} else {
		lines = len(menus)
		maxLines = m.menuPageSize
	}

	if m.options.CenterEverything {
		menuListBuilder.WriteString(m.centeredMenuView(a, lines))
	} else {
		var menuLines []string
		for i := 0; i < lines; i++ {
			menuLines = append(menuLines, m.menuLineView(a, i))
		}
		menuListBuilder.WriteString(lipgloss.JoinVertical(lipgloss.Left, menuLines...))
	}

	// fill blanks to maintain fixed page size
	if maxLines > lines {
		var fillLines []string
		blankLine := lipgloss.NewStyle().Width(a.WindowWidth() - m.menuStartColumn).Render("")
		for i := lines; i < maxLines; i++ {
			fillLines = append(fillLines, blankLine)
		}
		if lines > 0 {
			menuListBuilder.WriteByte('\n')
		}
		menuListBuilder.WriteString(lipgloss.JoinVertical(lipgloss.Left, fillLines...))
	}

	return menuListBuilder.String()
}

// truncateVisualWidth truncates s to fit within maxWidth visual cells,
// handling CJK characters (width 2) and other wide runes correctly.
// Returns the original string if it already fits.
func truncateVisualWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(s)
	var width int
	for i, r := range runes {
		rw := runewidth.RuneWidth(r)
		if width+rw > maxWidth {
			return string(runes[:i])
		}
		width += rw
	}
	return s
}

func (m *Main) menuItemView(a *App, index int) (string, int) {
	var (
		menuItemBuilder strings.Builder
		menuTitle       string
		itemMaxLen      int
		menuName        string
		windowWidth     = a.WindowWidth()
		maxIndexWidth   = m.getMaxIndexWidth()
	)

	isSelected := m.isSelected(index)
	isHovered := !m.inSearching && index == m.hoveredMenuItemIdx

	// Resolve title style based on selection + hover state
	ss := style.CurrentStyleSet()
	titleStyle := ss.MenuItem
	switch {
	case isHovered && isSelected:
		titleStyle = ss.SelectedItemHover
	case isHovered:
		titleStyle = ss.MenuItemHover
	case isSelected:
		titleStyle = ss.SelectedItem
	}

	if isSelected {
		menuTitle = fmt.Sprintf(fmt.Sprintf(" => %%%dd. %%s", maxIndexWidth), index, m.menuList[index].Title)
	} else {
		menuTitle = fmt.Sprintf(fmt.Sprintf("    %%%dd. %%s", maxIndexWidth), index, m.menuList[index].Title)
	}
	// if len(m.menuList[index].Subtitle) != 0 {
	menuTitle += " "
	// }

	if m.isDualColumn {
		if windowWidth <= 88 {
			itemMaxLen = (windowWidth - m.menuStartColumn - 4) / 2
		} else {
			if index%2 == 0 {
				itemMaxLen = 44
			} else {
				itemMaxLen = windowWidth - m.menuStartColumn - 44
			}
		}
	} else {
		itemMaxLen = windowWidth - m.menuStartColumn
	}

	menuTitleLen := lipgloss.Width(menuTitle)
	menuSubtitleLen := lipgloss.Width(m.menuList[index].Subtitle)

	var tmp string
	if menuTitleLen > itemMaxLen {
		// Title too long — manually truncate to visual width to avoid
		// line wrapping that would break the fixed-row UI layout.
		truncated := truncateVisualWidth(menuTitle, itemMaxLen)
		menuName = titleStyle.Render(truncated)
	} else if menuTitleLen+menuSubtitleLen > itemMaxLen {
		r := []rune(m.menuList[index].Subtitle + "   ")
		s := make([]rune, 0, itemMaxLen-menuTitleLen)
		indexStart := 0
		if m.options.Ticker != nil {
			indexStart = int(m.options.Ticker.PassedTime().Milliseconds() / 500 % int64(len(r)))
		}
		currentWidth := 0
		for i := indexStart; currentWidth < itemMaxLen-menuTitleLen; i = (i + 1) % len(r) {
			rw := lipgloss.Width(string(r[i]))
			if currentWidth+rw > itemMaxLen-menuTitleLen {
				break
			}
			s = append(s, r[i])
			currentWidth += rw
		}
		tmp = lipgloss.NewStyle().Width(itemMaxLen - menuTitleLen).MaxWidth(itemMaxLen - menuTitleLen).Render(string(s))
		menuName = titleStyle.Render(menuTitle) + ss.Subtitle.Render(tmp)
	} else {
		tmp = lipgloss.NewStyle().
			Width(itemMaxLen - menuTitleLen).
			Render(m.menuList[index].Subtitle)
		menuName = titleStyle.Render(menuTitle) + ss.Subtitle.Render(tmp)
	}

	menuItemBuilder.WriteString(menuName)

	return menuItemBuilder.String(), itemMaxLen
}

func (m *Main) menuLineView(a *App, line int) string {
	var index int
	if m.isDualColumn {
		index = line*2 + m.getPageStartIndex()
	} else {
		index = line + m.getPageStartIndex()
	}
	if index >= len(m.menuList) {
		return "" // beyond menu bounds — empty row
	}

	menuItemStr, _ := m.menuItemView(a, index)

	var row string
	if m.isDualColumn {
		var secondMenuItemStr string
		if index+1 < len(m.menuList) {
			secondMenuItemStr, _ = m.menuItemView(a, index+1)
		} else {
			secondMenuItemStr = "" // last item has no second column
		}
		// Fixed 4-space gap between columns
		row = menuItemStr + "    " + secondMenuItemStr
	} else {
		row = menuItemStr
	}
	// Left-align row at menuStartColumn (offset by -4 to account for " => "/"    " prefix)
	if m.menuStartColumn > 4 {
		row = lipgloss.NewStyle().PaddingLeft(m.menuStartColumn - 4).Render(row)
	}
	return row
}

func (m *Main) getPageStartIndex() int {
	return (m.menuCurPage - 1) * m.menuPageSize
}

func (m *Main) getMaxIndexWidth() int {
	return int(math.Log10(float64((m.menuPageSize*m.menuCurPage)-1))) + 1
}

// backButtonWidth is the display width of the back button icon including padding.
const backButtonWidth = 2 // "←"

// menuTitleY returns the 0-indexed Y position of the menu title row in the
// rendered output. Used for mouse hit-testing the back button.
func (m *Main) menuTitleY() int {
	y := m.computeTitleStartRow()
	if m.options.WhetherDisplayTitle {
		y++
	}
	// When status bar is at top, it occupies row 0 like the title bar does,
	// so menu title row must account for it even though WhetherDisplayTitle is forced false.
	if m.statusBar != nil && m.options.StatusBarPosition == StatusBarTop {
		y++
	}
	return y
}

// isOverBackButton checks if the given screen position falls within the back
// button area shown before the menu title when inside a submenu.
func (m *Main) isOverBackButton(x, y int, _ *App) bool {
	if m.menuStack.Len() <= 0 {
		return false
	}
	if y != m.menuTitleY() {
		return false
	}
	// The back icon "←" is rendered at column (menuStartColumn - backButtonWidth).
	iconCol := m.menuStartColumn - backButtonWidth
	return iconCol >= 0 && x >= iconCol && x < iconCol+backButtonWidth
}

// mouseInMenuArea checks if the given Y coordinate falls within the menu list bounds.
func (m *Main) mouseInMenuArea(y int) bool {
	return y >= m.menuListStartRow && y < m.menuBottomRow
}

// menuItemAt maps a screen (x, y) coordinate to a menuList index.
// Returns -1 if the coordinate is outside any menu item.
func (m *Main) menuItemAt(x, y int) int {
	row := y - m.menuListStartRow
	numCols := m.getNumColumns()

	maxLines := m.menuPageSize
	if m.isDualColumn {
		maxLines = int(math.Ceil(float64(m.menuPageSize) / 2))
	}
	if row < 0 || row >= maxLines {
		return -1
	}

	col := 0
	if m.isDualColumn {
		if m.options.CenterEverything {
			// In centered mode, columns are centered — split at midpoint
			if x > m.app.WindowWidth()/2 {
				col = 1
			}
		} else {
			// In left-aligned mode, columns start at menuStartColumn-4.
			// The left column occupies up to leftItemWidth, then 4-space gap,
			// then the right column.
			leftItemWidth := 44
			if m.app.WindowWidth() <= 88 {
				leftItemWidth = (m.app.WindowWidth() - m.menuStartColumn - 4) / 2
			}
			splitX := m.menuStartColumn - 4 + leftItemWidth + 2
			if x >= splitX {
				col = 1
			}
		}
	}

	idx := m.getPageStartIndex() + row*numCols + col
	if idx < 0 || idx >= len(m.menuList) || !m.menuItemHasTextAt(x, idx) {
		return -1
	}
	return idx
}

func (m *Main) menuItemHasTextAt(x, index int) bool {
	start, end, ok := m.menuItemTextBounds(index)
	return ok && x >= start && x < end
}

func (m *Main) menuItemTextBounds(index int) (start, end int, ok bool) {
	if index < 0 || index >= len(m.menuList) || m.app == nil {
		return 0, 0, false
	}
	if m.options.CenterEverything {
		return m.centeredMenuItemTextBounds(index)
	}

	item, _ := m.menuItemView(m.app, index)
	start, end, ok = visibleTextBounds(item)
	if !ok {
		return 0, 0, false
	}

	start += m.menuStartColumn - 4
	end += m.menuStartColumn - 4
	if m.isDualColumn && (index-m.getPageStartIndex())%2 == 1 {
		_, leftWidth := m.menuItemView(m.app, index-1)
		start += leftWidth + 4
		end += leftWidth + 4
	}
	return start, end, true
}

func (m *Main) centeredMenuItemTextBounds(index int) (start, end int, ok bool) {
	menus := m.getCurPageMenus()
	titleLengths := make([]int, len(menus))
	for i, item := range menus {
		titleLengths[i] = layout.Width(item.OriginString())
	}
	entryLength := m.centeredEntryLength(m.app, titleLengths)
	entry := m.formatEntry(&m.menuList[index], index, entryLength)
	start, end, ok = visibleTextBounds(entry)
	if !ok {
		return 0, 0, false
	}

	column := (index - m.getPageStartIndex()) % m.getNumColumns()
	if m.isDualColumn {
		start += (m.app.WindowWidth()-entryLength*2)/2 + column*entryLength
		end += (m.app.WindowWidth()-entryLength*2)/2 + column*entryLength
	} else {
		start += (m.app.WindowWidth() - entryLength) / 2
		end += (m.app.WindowWidth() - entryLength) / 2
	}
	return start, end, true
}

func visibleTextBounds(view string) (start, end int, ok bool) {
	column := 0
	start = -1
	for _, r := range ansi.Strip(view) {
		width := runewidth.RuneWidth(r)
		if !unicode.IsSpace(r) {
			if start < 0 {
				start = column
			}
			end = column + width
		}
		column += width
	}
	return start, end, start >= 0
}

func (m *Main) isSelected(index int) bool {
	return !m.inSearching && index == m.selectedIndex
}

func (m *Main) searchInputView(app *App) string {
	var (
		windowWidth = app.WindowWidth()
		ss          = style.CurrentStyleSet()
	)

	if !m.inSearching {
		// Help hint bar: shows per-menu keyboard shortcuts when search is inactive.
		// Each Menu can override HelpHints() to customize the displayed shortcuts.
		hints := m.menu.HelpHints()
		if len(hints) == 0 {
			return "" // menu opted out of help bar
		}
		var parts []string
		for _, h := range hints {
			parts = append(parts,
				ss.HintKey.Render("  "+h.Key)+ss.Muted.Render(" "+h.Desc),
			)
		}
		hint := layout.JoinHorizontal(layout.Top, parts...)
		return lipgloss.NewStyle().
			Width(windowWidth).
			Align(lipgloss.Center).
			PaddingTop(1).
			Render(hint)
	}

	// Search input: left-aligned with menu, same row as the help bar.
	inputView := m.searchInput.View()
	inputView = lipgloss.NewStyle().
		Width(windowWidth).
		PaddingLeft(m.menuStartColumn).
		PaddingTop(1).
		Render(inputView)

	return inputView
}

func (m *Main) getCurPageMenus() []MenuItem {
	start := m.getPageStartIndex()
	end := int(math.Min(float64(len(m.menuList)), float64(m.menuCurPage*m.menuPageSize)))

	return m.menuList[start:end]
}

// key handle
func (m *Main) keyMsgHandle(msg tea.KeyMsg, a *App) (Page, tea.Cmd) {
	if m.inSearching {
		switch msg.String() {
		case "esc":
			m.inSearching = false
			m.searchInput.Blur()
			m.searchInput.Reset()
			return m, a.RerenderCmd(true)
		case "enter":
			m.searchMenuHandle()
			return m, a.RerenderCmd(true)
		}
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, tea.Batch(cmd)
	}

	// Tab switching (when tabs enabled and not in search mode)
	if m.options.EnableTabs && m.tabs != nil && len(m.tabStates) > 0 {
		key := msg.String()
		switch key {
		case "ctrl+tab":
			m.switchTab((m.activeTab + 1) % len(m.tabStates))
			return m, a.RerenderCmd(true)
		case "ctrl+shift+tab":
			newIndex := m.activeTab - 1
			if newIndex < 0 {
				newIndex = len(m.tabStates) - 1
			}
			m.switchTab(newIndex)
			return m, a.RerenderCmd(true)
		case "ctrl+right":
			m.switchTab((m.activeTab + 1) % len(m.tabStates))
			return m, a.RerenderCmd(true)
		case "ctrl+left":
			newIndex := m.activeTab - 1
			if newIndex < 0 {
				newIndex = len(m.tabStates) - 1
			}
			m.switchTab(newIndex)
			return m, a.RerenderCmd(true)
		}
	}

	var (
		key             = msg.String()
		newPage         Page
		lastCmd         tea.Cmd
		stopPropagation bool
	)
	for _, c := range m.kbCtrls {
		stopPropagation, newPage, lastCmd = c.KeyMsgHandle(msg, a)
		if stopPropagation {
			if newPage != nil {
				return newPage, func() tea.Msg { return newPage.Msg() }
			}
			if lastCmd == nil {
				lastCmd = a.Tick(time.Nanosecond)
			}
			return m, lastCmd
		}
	}

	switch key {
	case "j", "J", "down":
		newPage = m.MoveDown()
	case "k", "K", "up":
		newPage = m.MoveUp()
	case "h", "H", "left":
		newPage = m.MoveLeft()
	case "l", "L", "right":
		newPage = m.MoveRight()
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		num, _ := strconv.Atoi(key)
		start := m.getPageStartIndex()
		if start+num >= len(m.menuList) {
			break
		}
		target := start + num
		if m.selectedIndex == target {
			newPage = m.enterMenuWithLoading(nil, nil)
			if m.pendingEnterMenu != nil {
				return m, a.RerenderCmd(true)
			}
		} else {
			m.selectedIndex = target
		}
	case "g":
		newPage = m.MoveTop()
	case "G":
		newPage = m.MoveBottom()
	case "n", "N", "enter":
		if m.selectedIndex < 0 {
			break
		}
		return m.activateSelectedItemWithLoading(a)
	case "b", "B", "esc":
		newPage = m.BackMenu()
	case "r", "R":
		return m, a.RerenderCmd(true)
	case "/", "／", "、":
		if m.menu.IsSearchable() {
			m.inSearching = true
			m.searchInput.Focus()
		}
	}

	if newPage != nil {
		return newPage, func() tea.Msg { return newPage.Msg() }
	}
	return m, a.Tick(time.Nanosecond)
}

// mouse handle
func (m *Main) mouseMsgHandle(msg tea.MouseMsg, a *App) (Page, tea.Cmd) {
	mouse := msg.Mouse()

	// External controllers — run first, they have priority
	var (
		newPage         Page
		lastCmd         tea.Cmd
		stopPropagation bool
	)
	for _, c := range m.mouseCtrls {
		stopPropagation, newPage, lastCmd = c.MouseMsgHandle(msg, a)
		if stopPropagation {
			break
		}
	}
	if stopPropagation {
		if newPage != nil {
			return newPage, func() tea.Msg { return newPage.Msg() }
		}
		if lastCmd == nil {
			lastCmd = a.Tick(time.Nanosecond)
		}
		return m, lastCmd
	}

	// --- DEBUG: log all mouse click events ---
	// if f, err := os.OpenFile("/tmp/foxful-mouse-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
	// 	inMenu := m.mouseInMenuArea(mouse.Y)
	// 	log.New(f, "", log.LstdFlags|log.Lmicroseconds).Printf(
	// 		"button=%-10s x=%-3d y=%-3d inMenu=%v searching=%v",
	// 		mouse.Button.String(), mouse.X, mouse.Y, inMenu, m.inSearching,
	// 	)
	// 	f.Close()
	// }
	// --- END DEBUG ---

	// Only process concrete message types for built-in menu handling.
	// Bubbletea v2 sends both MouseClickMsg + MouseReleaseMsg per physical click;
	// we must ignore MouseReleaseMsg to avoid false double-click detection.
	switch msg.(type) {
	case tea.MouseClickMsg:
		return m.mouseClickHandle(mouse, a)
	case tea.MouseMotionMsg:
		return m.mouseMotionHandle(mouse, a)
	case tea.MouseReleaseMsg:
		// Ignore — hover and pointer state are driven by mouseMotionHandle.
		// Clearing them here would flicker the pointer when clicking a menu
		// item (mouse still over clickable area after release).
		return m, a.Tick(time.Nanosecond)
	case tea.MouseWheelMsg:
		return m.mouseWheelHandle(mouse, a)
	}

	// Nothing handled — tick to keep the event loop alive
	return m, a.Tick(time.Nanosecond)
}

// mouseClickHandle processes mouse click events for the built-in menu.
func (m *Main) mouseClickHandle(mouse tea.Mouse, a *App) (Page, tea.Cmd) {
	if m.inSearching {
		return m, a.Tick(time.Nanosecond)
	}

	switch mouse.Button {
	case tea.MouseLeft:
		// Check tab bar click (when multi-tab mode enabled)
		if m.options.EnableTabs && m.tabs != nil {
			if tabIdx := m.tabIndexAt(mouse.X, mouse.Y, a); tabIdx >= 0 {
				if tabIdx != m.activeTab {
					m.switchTab(tabIdx)
					return m, a.RerenderCmd(true)
				}
				return m, a.Tick(time.Nanosecond)
			}
		}

		// Check back button click (navigate back to parent menu)
		if m.isOverBackButton(mouse.X, mouse.Y, a) {
			newPage := m.BackMenu()
			if newPage != nil {
				return newPage, a.RerenderCmd(true)
			}
			return m, a.RerenderCmd(true)
		}

		// Check menu area click (existing behavior)
		if m.mouseInMenuArea(mouse.Y) {
			idx := m.menuItemAt(mouse.X, mouse.Y)
			if idx < 0 || idx >= len(m.menuList) {
				break
			}

			now := time.Now()
			doubleClickInterval := m.doubleClickInterval()

			// Position tolerance (±2px)
			deltaX := mouse.X - m.lastClickX
			if deltaX < 0 {
				deltaX = -deltaX
			}
			deltaY := mouse.Y - m.lastClickY
			if deltaY < 0 {
				deltaY = -deltaY
			}

			// Double-click: within interval AND close position
			if now.Sub(m.lastClickTime) <= doubleClickInterval &&
				deltaX <= 2 && deltaY <= 2 {
				// Double-click → enter submenu or execute custom action
				m.selectedIndex = idx
				m.lastClickTime = time.Time{} // reset

				return m.activateSelectedItemWithLoading(a)
			}

			// Single click → just focus/select, never enter
			m.selectedIndex = idx
			m.lastClickTime = now
			m.lastClickX = mouse.X
			m.lastClickY = mouse.Y
			return m, a.RerenderCmd(true)
		}

		// Check status bar breadcrumb click
		if newPage := m.handleBreadcrumbClick(mouse.X, mouse.Y, a); newPage != nil {
			return newPage, a.RerenderCmd(true)
		}

	case tea.MouseBackward:
		if !m.mouseInMenuArea(mouse.Y) {
			break
		}
		// Back button: return to parent menu
		newPage := m.BackMenu()
		if newPage != nil {
			return newPage, a.RerenderCmd(true)
		}
		return m, a.RerenderCmd(true)

	case tea.MouseForward:
		if !m.mouseInMenuArea(mouse.Y) {
			break
		}
		// Forward button: enter selected item's submenu
		newPage := m.enterMenuWithLoading(nil, nil)
		if m.pendingEnterMenu != nil {
			return m, a.RerenderCmd(true)
		}
		if newPage != nil {
			return newPage, nil
		}
		return m, a.RerenderCmd(true)

	case tea.MouseMiddle:
		if !m.mouseInMenuArea(mouse.Y) {
			break
		}
		newPage := m.BackMenu()
		if newPage != nil {
			return newPage, a.RerenderCmd(true)
		}
		return m, a.RerenderCmd(true)

	case tea.MouseRight:
		idx := m.menuItemAt(mouse.X, mouse.Y)
		if idx < 0 || idx >= len(m.menuList) {
			idx = -1 // 空白区域
		}
		items := m.menu.ContextMenuItems(a, idx)
		if len(items) == 0 {
			break
		}
		if idx >= 0 {
			m.selectedIndex = idx
		}
		contextMenu := newContextMenu(m.menu, idx, items, mouse.X, mouse.Y, m.options.ContextMenuOptions)
		a.pushModal(contextMenu)
		return m, a.RerenderCmd(true)

	}

	return m, a.Tick(time.Nanosecond)
}

// mouseWheelHandle processes mouse wheel events for the built-in menu.
func (m *Main) mouseWheelHandle(mouse tea.Mouse, a *App) (Page, tea.Cmd) {
	if m.inSearching {
		return m, a.Tick(time.Nanosecond)
	}

	switch mouse.Button {
	case tea.MouseWheelUp:
		if !m.mouseInMenuArea(mouse.Y) {
			break
		}
		newPage := m.MoveUp()
		if newPage != nil {
			return newPage, a.RerenderCmd(true)
		}
		return m, a.RerenderCmd(true)
	case tea.MouseWheelDown:
		if !m.mouseInMenuArea(mouse.Y) {
			break
		}
		newPage := m.MoveDown()
		if newPage != nil {
			return newPage, a.RerenderCmd(true)
		}
		return m, a.RerenderCmd(true)
	}

	return m, a.Tick(time.Nanosecond)
}

// doubleClickInterval returns the OS-specific double-click interval threshold.
func (m *Main) doubleClickInterval() time.Duration {
	switch runtime.GOOS {
	case "darwin":
		return 400 * time.Millisecond
	case "windows":
		return 500 * time.Millisecond
	default:
		return 300 * time.Millisecond
	}
}

func (m *Main) searchMenuHandle() {
	m.inSearching = false
	searchMenu := m.options.LocalSearchMenu
	if m.options.LocalSearchMenu == nil {
		searchMenu = DefaultSearchMenu()
	}
	searchMenu.Search(m.menu, m.searchInput.Value())
	m.EnterMenu(searchMenu, &MenuItem{Title: SearchResult, Subtitle: m.searchInput.Value()})
	m.searchInput.Blur()
	m.searchInput.Reset()
}

type menuStackItem struct {
	menuList      []MenuItem
	selectedIndex int
	menuCurPage   int
	menuTitle     *MenuItem
	menu          Menu
}

func (m *Main) MoveUp() Page {
	var (
		topHook = m.menu.TopOutHook()
		newPage Page
		res     bool
	)
	if m.isDualColumn {
		if m.selectedIndex-2 < 0 && topHook != nil {
			loading := NewLoading(m)
			loading.Start()
			if res, newPage = topHook(m); !res {
				loading.Complete()
				return newPage
			}
			// update menu ui
			m.menuList = m.menu.MenuViews()
			loading.Complete()
		}
		if m.selectedIndex-2 < 0 {
			return nil
		}
		m.selectedIndex -= 2
	} else {
		if m.selectedIndex-1 < 0 && topHook != nil {
			loading := NewLoading(m)
			loading.Start()
			if res, newPage = topHook(m); !res {
				loading.Complete()
				return newPage
			}
			m.menuList = m.menu.MenuViews()
			loading.Complete()
		}
		if m.selectedIndex-1 < 0 {
			return nil
		}
		m.selectedIndex--
	}
	if m.selectedIndex < m.getPageStartIndex() {
		newPage = m.PrePage()
	}
	return newPage
}

func (m *Main) MoveDown() Page {
	var (
		bottomHook = m.menu.BottomOutHook()
		newPage    Page
		res        bool
	)
	// Initial state: no item selected — select first item
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
		return nil
	}
	if m.isDualColumn {
		if m.selectedIndex+2 > len(m.menuList)-1 && bottomHook != nil {
			loading := NewLoading(m)
			loading.Start()
			if res, newPage = bottomHook(m); !res {
				loading.Complete()
				return newPage
			}
			m.menuList = m.menu.MenuViews()
			loading.Complete()
		}
		if m.selectedIndex+2 > len(m.menuList)-1 {
			return nil
		}
		m.selectedIndex += 2
	} else {
		if m.selectedIndex+1 > len(m.menuList)-1 && bottomHook != nil {
			loading := NewLoading(m)
			loading.Start()
			if res, newPage = bottomHook(m); !res {
				loading.Complete()
				return newPage
			}
			m.menuList = m.menu.MenuViews()
			loading.Complete()
		}
		if m.selectedIndex+1 > len(m.menuList)-1 {
			return nil
		}
		m.selectedIndex++
	}
	if m.selectedIndex >= m.menuCurPage*m.menuPageSize {
		newPage = m.NextPage()
	}
	return newPage
}

func (m *Main) MoveLeft() Page {
	if !m.isDualColumn || m.selectedIndex%2 == 0 || m.selectedIndex-1 < 0 {
		return nil
	}
	m.selectedIndex--
	return nil
}

func (m *Main) MoveRight() Page {
	if !m.isDualColumn || m.selectedIndex%2 != 0 {
		return nil
	}
	var (
		newPage Page
		res     bool
	)
	if bottomHook := m.menu.BottomOutHook(); m.selectedIndex >= len(m.menuList)-1 && bottomHook != nil {
		loading := NewLoading(m)
		loading.Start()
		if res, newPage = bottomHook(m); !res {
			loading.Complete()
			return newPage
		}
		m.menuList = m.menu.MenuViews()
		loading.Complete()
	}
	if m.selectedIndex >= len(m.menuList)-1 {
		return nil
	}
	m.selectedIndex++
	return newPage
}

func (m *Main) MoveTop() Page {
	if m.isDualColumn {
		m.selectedIndex = m.selectedIndex % 2
	} else {
		m.selectedIndex = 0
	}
	m.menuCurPage = 1
	return nil
}

func (m *Main) MoveBottom() Page {
	if m.isDualColumn && len(m.menuList)%2 == 0 {
		m.selectedIndex = len(m.menuList) + (m.selectedIndex%2 - 2)
	} else if m.isDualColumn && m.selectedIndex%2 != 0 {
		m.selectedIndex = len(m.menuList) - 2
	} else {
		m.selectedIndex = len(m.menuList) - 1
	}
	m.menuCurPage = int(math.Ceil(float64(len(m.menuList)) / float64(m.menuPageSize)))
	if m.isDualColumn && m.selectedIndex%2 != 0 && len(m.menuList)%m.menuPageSize == 1 {
		m.menuCurPage -= 1
	}
	return nil
}

func (m *Main) PrePage() Page {
	var (
		newPage Page
		res     bool
	)
	if prePageHook := m.menu.BeforePrePageHook(); prePageHook != nil {
		loading := NewLoading(m)
		loading.Start()
		if res, newPage = prePageHook(m); !res {
			loading.Complete()
			return newPage
		}
		loading.Complete()
	}
	if m.menuCurPage <= 1 {
		return nil
	}
	m.menuCurPage--
	m.hoveredMenuItemIdx = -1
	return newPage
}

func (m *Main) NextPage() Page {
	var (
		res     bool
		newPage Page
	)
	if nextPageHook := m.menu.BeforeNextPageHook(); nextPageHook != nil {
		loading := NewLoading(m)
		loading.Start()
		if res, newPage = nextPageHook(m); !res {
			loading.Complete()
			return newPage
		}
		loading.Complete()
	}
	if m.menuCurPage >= int(math.Ceil(float64(len(m.menuList))/float64(m.menuPageSize))) {
		return nil
	}

	m.menuCurPage++
	m.hoveredMenuItemIdx = -1
	return newPage
}

// enterMenuWithLoading initiates a deferred submenu entry. Unlike EnterMenu
// which runs the BeforeEnterMenuHook synchronously, this method defers the hook
// to the next tickMainMsg cycle. This gives the current View() cycle a chance
// to render the loadingTips text before the (potentially slow) API call in the
// hook blocks the event loop.
//
// Returns nil on success (the actual menu transition happens in the tick handler).
// Returns a non-nil Page on immediate failure (e.g., login required).
func (m *Main) enterMenuWithLoading(newMenu Menu, newTitle *MenuItem) Page {
	if m.pendingEnterMenu != nil {
		return nil // already pending, wait for completion
	}

	if newMenu == nil {
		newMenu = m.menu.SubMenu(m.app, m.selectedIndex)
	}
	if newTitle == nil && m.selectedIndex >= 0 && m.selectedIndex < len(m.menuList) {
		newTitle = &m.menuList[m.selectedIndex]
	}

	if newMenu == nil || newTitle == nil {
		return nil
	}

	m.hoveredBreadcrumbIdx = -1
	m.hoveredMenuItemIdx = -1
	m.hoveredBackButton = false

	stackItem := &menuStackItem{
		menuList:      m.menuList,
		selectedIndex: m.selectedIndex,
		menuCurPage:   m.menuCurPage,
		menuTitle:     m.menuTitle,
		menu:          m.menu,
	}
	m.menuStack.Push(stackItem)

	loading := NewLoading(m)
	loading.Start() // sets m.loadingTips so the next View() shows progress

	m.pendingEnterMenu = &enterMenuDeferred{
		newMenu:   newMenu,
		newTitle:  newTitle,
		loading:   loading,
		stackItem: stackItem,
	}

	return nil
}

// activateSelectedItemWithLoading defers Action until the loading state has rendered.
// A nil Action result retains the existing fallback to the selected item's submenu.
func (m *Main) activateSelectedItemWithLoading(a *App) (Page, tea.Cmd) {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.menuList) {
		return m, a.Tick(time.Nanosecond)
	}
	if m.pendingMenuAction != nil || m.pendingEnterMenu != nil || m.pendingRefreshMenu != nil || m.pendingAction != nil {
		return m, nil
	}

	loading := NewLoading(m)
	loading.Start()
	m.pendingMenuAction = &menuActionDeferred{
		menu:    m.menu,
		index:   m.selectedIndex,
		item:    m.menuList[m.selectedIndex],
		loading: loading,
	}
	return m, a.RerenderCmd(true)
}

// DeferWithLoading 在下一次 tick 执行延迟动作，并先显示加载提示。
//
// action 的返回值语义：
//   - (true, nil)：正常完成，结束加载并重绘
//   - (false, nil)：中途取消或失败，结束加载并重绘
//   - (false, newPage)：结束加载并跳转到 newPage
//
// 如果当前已有其他待执行操作，静默返回 false。
// 所有状态写入均在 Update 内完成，避免与 View 并发访问。
func (m *Main) DeferWithLoading(action func(*Main) (bool, Page)) bool {
	if m.pendingMenuAction != nil || m.pendingEnterMenu != nil || m.pendingRefreshMenu != nil || m.pendingAction != nil {
		return false
	}

	loading := NewLoading(m)
	loading.Start()
	m.pendingAction = &actionDeferred{
		action:  action,
		loading: loading,
	}
	return true
}

// RefreshMenuWithLoading 在下一次 tick 刷新当前菜单，并先显示加载提示。
// 所有状态写入均在 Update 内完成，避免与 View 并发访问。
func (m *Main) RefreshMenuWithLoading() {
	m.DeferWithLoading(func(m *Main) (bool, Page) {
		if hook := m.menu.BeforeEnterMenuHook(); hook != nil {
			if res, newPage := hook(m); !res {
				return false, newPage
			}
		}
		m.menuList = m.menu.MenuViews()
		return true, nil
	})
}

func (m *Main) EnterMenu(newMenu Menu, newTitle *MenuItem) Page {
	if (newMenu == nil || newTitle == nil) && m.selectedIndex >= len(m.menuList) {
		return nil
	}

	if newMenu == nil {
		newMenu = m.menu.SubMenu(m.app, m.selectedIndex)
	}
	if newTitle == nil {
		if m.selectedIndex >= 0 {
			newTitle = &m.menuList[m.selectedIndex]
		}
	}

	m.hoveredBreadcrumbIdx = -1
	m.hoveredMenuItemIdx = -1
	m.hoveredBackButton = false

	stackItem := &menuStackItem{
		menuList:      m.menuList,
		selectedIndex: m.selectedIndex,
		menuCurPage:   m.menuCurPage,
		menuTitle:     m.menuTitle,
		menu:          m.menu,
	}
	m.menuStack.Push(stackItem)

	if newMenu == nil {
		m.menuStack.Pop()
		return nil
	}

	var (
		res     bool
		newPage Page
	)
	if enterMenuHook := newMenu.BeforeEnterMenuHook(); enterMenuHook != nil {
		loading := NewLoading(m)
		loading.Start()
		if res, newPage = enterMenuHook(m); !res {
			loading.Complete()
			m.menuStack.Pop()
			return newPage
		}
		loading.Complete()
	}
	if newMenu != nil {
		newMenu.FormatMenuItem(newTitle)
	}

	menuList := newMenu.MenuViews()

	m.menu = newMenu
	m.menuList = menuList
	m.menuTitle = newTitle
	m.selectedIndex = 0
	m.menuCurPage = 1

	return newPage
}

func (m *Main) BackMenu() Page {
	if m.menuStack.Len() <= 0 {
		return nil
	}

	m.hoveredBreadcrumbIdx = -1
	m.hoveredMenuItemIdx = -1
	m.hoveredBackButton = false

	var (
		stackItem = m.menuStack.Pop()
		newPage   Page
		res       bool
	)
	if backMenuHook := m.menu.BeforeBackMenuHook(); backMenuHook != nil {
		loading := NewLoading(m)
		loading.Start()
		if res, newPage = backMenuHook(m); !res {
			loading.Complete()
			m.menuStack.Push(stackItem)
			return newPage
		}
		loading.Complete()
	}
	m.menu.FormatMenuItem(m.menuTitle)

	stackMenu, ok := stackItem.(*menuStackItem)
	if !ok {
		return nil
	}

	m.menuList = stackMenu.menuList
	m.menu = stackMenu.menu
	m.menuTitle = stackMenu.menuTitle
	m.menu.FormatMenuItem(m.menuTitle)
	m.selectedIndex = stackMenu.selectedIndex
	m.menuCurPage = stackMenu.menuCurPage

	return newPage
}

// BackToMenu pops count levels from the menu stack (or until the stack is
// empty). The current menu's BeforeBackMenuHook is called first; intermediate
// menus that are skipped over do NOT get their hooks called. Must only be
// called from the mouse click handler after hit-testing breadcrumb segments.
func (m *Main) BackToMenu(count int) Page {
	if count <= 0 {
		return nil
	}
	if m.menuStack.Len() <= 0 {
		return nil
	}

	m.hoveredBreadcrumbIdx = -1
	m.hoveredMenuItemIdx = -1
	m.hoveredBackButton = false

	// Call hook on the current (deepest) menu once
	var newPage Page
	if backMenuHook := m.menu.BeforeBackMenuHook(); backMenuHook != nil {
		loading := NewLoading(m)
		loading.Start()
		var res bool
		if res, newPage = backMenuHook(m); !res {
			loading.Complete()
			// Hook refused — don't pop anything
			if newPage == nil {
				return nil
			}
			return newPage
		}
		loading.Complete()
	}
	m.menu.FormatMenuItem(m.menuTitle)

	// Pop count levels, keeping the last popped item as the target state.
	var targetStackItem *menuStackItem
	for i := 0; i < count; i++ {
		if m.menuStack.Len() <= 0 {
			break
		}
		item := m.menuStack.Pop()
		if si, ok := item.(*menuStackItem); ok {
			targetStackItem = si
		}
	}
	if targetStackItem == nil {
		return nil
	}

	// Restore the target state
	m.menuList = targetStackItem.menuList
	m.menu = targetStackItem.menu
	m.menuTitle = targetStackItem.menuTitle
	m.menu.FormatMenuItem(m.menuTitle)
	m.selectedIndex = targetStackItem.selectedIndex
	m.menuCurPage = targetStackItem.menuCurPage

	return newPage
}

// breadcrumbSegmentAt returns the breadcrumb segment display index and depth
// index at the given screen position (x, y). Returns (-1, 0, false) when no
// clickable ancestor segment is at that position. Only works for
// DefaultStatusBar layout — returns false for other status bars.
func (m *Main) breadcrumbSegmentAt(x, y int, a *App) (segIdx int, depthIdx int, ok bool) {
	if m.menuStack.Len() <= 0 {
		return -1, 0, false
	}
	if m.statusBar == nil {
		return -1, 0, false
	}
	if _, ok := m.statusBar.(*DefaultStatusBar); !ok {
		return -1, 0, false
	}

	// Status bar occupies a specific row based on position.
	statusBarRow := m.statusBarRowY(a)
	if statusBarRow < 0 || y != statusBarRow {
		return -1, 0, false
	}

	segments := computeBreadcrumbSegments(m)
	if len(segments) == 0 {
		return -1, 0, false
	}

	ss := style.CurrentStyleSet()
	pathLabel := ss.StatusBarNuggetLabel.Render(" » ")
	labelW := lipgloss.Width(pathLabel)
	segStartX := labelW + 1

	for i, seg := range segments {
		if seg.IsEllipsis {
			segStartX += seg.DisplayWidth + 3
			continue
		}
		if seg.IsLast {
			break
		}

		segEndX := segStartX + seg.DisplayWidth
		if x >= segStartX && x < segEndX {
			return i, seg.DepthIndex, true
		}

		segStartX = segEndX + 3 // " / " = 3 chars
	}

	return -1, 0, false
}

// isOverClickableElement returns true if the given screen position is over
// an interactive/clickable element that should show a pointer cursor.
func (m *Main) isOverClickableElement(x, y int, a *App) bool {
	// 1. Back button (clickable to navigate back to parent menu)
	if m.isOverBackButton(x, y, a) {
		return true
	}

	// 2. Breadcrumb ancestor segment (clickable to navigate back)
	if _, _, ok := m.breadcrumbSegmentAt(x, y, a); ok {
		return true
	}

	// 3. Tab bar (when multi-tab mode enabled)
	if m.options.EnableTabs && m.tabs != nil {
		if tabIdx := m.tabIndexAt(x, y, a); tabIdx >= 0 {
			return true
		}
	}

	// 4. Menu list area (single-click selects, double-click enters)
	if !m.inSearching && m.mouseInMenuArea(y) {
		idx := m.menuItemAt(x, y)
		if idx >= 0 && idx < len(m.menuList) {
			return true
		}
	}

	return false
}

// mouseMotionHandle processes mouse motion events for hover effects and
// terminal mouse pointer shape changes. It updates both the breadcrumb
// hover rendering state and the global pointer cursor.
func (m *Main) mouseMotionHandle(mouse tea.Mouse, a *App) (Page, tea.Cmd) {
	oldBreadcrumbHover := m.hoveredBreadcrumbIdx
	oldTabHover := m.hoveredTabIdx
	oldPointerActive := m.hoverPointerActive
	oldBackButtonHover := m.hoveredBackButton

	if m.inSearching {
		stateChanged := false
		if oldBreadcrumbHover != -1 {
			m.hoveredBreadcrumbIdx = -1
			stateChanged = true
		}
		if m.hoveredMenuItemIdx != -1 {
			m.hoveredMenuItemIdx = -1
			stateChanged = true
		}
		if m.hoveredTabIdx != -1 {
			m.hoveredTabIdx = -1
			stateChanged = true
		}
		if m.hoveredBackButton {
			m.hoveredBackButton = false
			stateChanged = true
		}
		if oldPointerActive {
			m.hoverPointerActive = false
			stateChanged = true
		}
		if stateChanged {
			return m, tea.Sequence(a.RerenderCmd(true), a.SetMousePointer("default"))
		}
		return m, nil
	}

	// Update breadcrumb hover (visual underline effect)
	segIdx, _, bcOk := m.breadcrumbSegmentAt(mouse.X, mouse.Y, a)
	if bcOk {
		m.hoveredBreadcrumbIdx = segIdx
	} else {
		m.hoveredBreadcrumbIdx = -1
	}

	// Update tab hover (when multi-tab mode enabled)
	if m.options.EnableTabs && m.tabs != nil {
		m.hoveredTabIdx = m.tabIndexAt(mouse.X, mouse.Y, a)
	} else {
		m.hoveredTabIdx = -1
	}

	// Update menu item hover
	oldMenuItemHover := m.hoveredMenuItemIdx
	if !m.inSearching && m.mouseInMenuArea(mouse.Y) {
		idx := m.menuItemAt(mouse.X, mouse.Y)
		if idx >= 0 && idx < len(m.menuList) {
			m.hoveredMenuItemIdx = idx
		} else {
			m.hoveredMenuItemIdx = -1
		}
	} else {
		m.hoveredMenuItemIdx = -1
	}

	// Update back button hover
	m.hoveredBackButton = m.isOverBackButton(mouse.X, mouse.Y, a)

	// Update global pointer state
	m.hoverPointerActive = m.isOverClickableElement(mouse.X, mouse.Y, a)

	// Compute commands for state changes
	var cmds []tea.Cmd
	if m.hoveredBreadcrumbIdx != oldBreadcrumbHover || m.hoveredMenuItemIdx != oldMenuItemHover || m.hoveredBackButton != oldBackButtonHover || m.hoveredTabIdx != oldTabHover {
		cmds = append(cmds, a.RerenderCmd(true))
	}
	if m.hoverPointerActive != oldPointerActive {
		if m.hoverPointerActive {
			cmds = append(cmds, a.SetMousePointer("pointer"))
		} else {
			cmds = append(cmds, a.SetMousePointer("default"))
		}
	}

	if len(cmds) > 0 {
		return m, tea.Sequence(cmds...)
	}
	return m, nil
}

// handleBreadcrumbClick handles a left-click on the status bar breadcrumb.
// Navigates back to the clicked menu level. Returns nil if no clickable
// breadcrumb segment was hit.
// UpdateBreadcrumbHover updates the breadcrumb hover state based on the given
// screen position and returns whether the hovered segment changed (so the
// caller should re-render) and whether the position is over a clickable
// breadcrumb segment (so the caller should show a pointer cursor).
//
// This is exported for custom pages (search/login/QR) that render the top
// status bar via pageTitleView but handle their own mouse events instead of
// going through Main.MouseMsgHandle. Callers must only invoke this when the
// top status bar is actually rendered on their page.
func (m *Main) UpdateBreadcrumbHover(x, y int, a *App) (changed bool, over bool) {
	old := m.hoveredBreadcrumbIdx
	segIdx, _, ok := m.breadcrumbSegmentAt(x, y, a)
	if ok {
		m.hoveredBreadcrumbIdx = segIdx
	} else {
		m.hoveredBreadcrumbIdx = -1
	}
	return m.hoveredBreadcrumbIdx != old, ok
}

// HandleBreadcrumbClick is the exported entry point for custom pages to
// delegate a breadcrumb left-click to Main. Returns the target page (Main
// navigated back to the clicked level) or nil when no clickable segment was
// hit.
func (m *Main) HandleBreadcrumbClick(x, y int, a *App) Page {
	return m.handleBreadcrumbClick(x, y, a)
}

func (m *Main) handleBreadcrumbClick(x, y int, a *App) Page {
	_, depthIdx, ok := m.breadcrumbSegmentAt(x, y, a)
	if !ok {
		return nil
	}

	// Compute full path length for pop count
	fullPathLen := m.menuStack.Len()
	if m.menuStack.Len() > 0 {
		stackItems := m.menuStack.ToSlice()
		lastItem := stackItems[len(stackItems)-1].(*menuStackItem)
		if lastItem.menuTitle.Title != m.menuTitle.Title {
			fullPathLen++
		}
	} else {
		fullPathLen = 1
	}

	popCount := fullPathLen - 1 - depthIdx
	return m.BackToMenu(popCount)
}

func TickMain(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return tickMainMsg{}
	})
}
