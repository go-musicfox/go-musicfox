package model

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

type Main struct {
	options *Options

	app *App

	isDualColumn bool

	menuTitle            *MenuItem
	menuTitleStartRow    int
	menuTitleStartColumn int

	menuStartRow    int
	menuStartColumn int
	menuBottomRow   int

	menuCurPage  int
	menuPageSize int

	menuList      []MenuItem
	menuStack     *util.Stack
	selectedIndex int

	// local search
	inSearching bool
	searchInput textinput.Model

	menu Menu // current menu

	components []Component

	kbCtrls    []KeyboardController
	mouseCtrls []MouseController
}

type tickMainMsg struct{}

func NewMain(app *App, options *Options) (m *Main) {
	var mainMenuTitle *MenuItem
	if options.MainMenuTitle != nil {
		mainMenuTitle = options.MainMenuTitle
	} else {
		mainMenuTitle = &MenuItem{Title: options.AppName}
	}

	m = &Main{
		app:          app,
		options:      options,
		menuTitle:    mainMenuTitle,
		menu:         options.MainMenu,
		menuStack:    &util.Stack{},
		menuCurPage:  1,
		menuPageSize: 10,
		searchInput:  textinput.New(),
		components:   options.Components,
		kbCtrls:      options.KBControllers,
		mouseCtrls:   options.MouseControllers,
	}
	m.menuList = m.menu.MenuViews()
	m.searchInput.Placeholder = " 搜索"
	m.searchInput.Prompt = util.GetFocusedPrompt()
	m.searchInput.TextStyle = util.GetPrimaryFontStyle()
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

func (m *Main) Update(msg tea.Msg, a *App) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.keyMsgHandle(msg, a)
	case tea.MouseMsg:
		return m.mouseMsgHandle(msg, a)
	case tickMainMsg:
		return m, nil
	case tea.WindowSizeMsg:
		m.isDualColumn = msg.Width >= 75 && m.options.DualColumn

		// 菜单开始行、列
		m.menuStartRow = msg.Height / 3
		if !m.options.WhetherDisplayTitle && m.menuStartRow > 1 {
			m.menuStartRow--
		}
		if m.isDualColumn {
			m.menuStartColumn = (msg.Width - 60) / 2
			m.menuBottomRow = m.menuStartRow + int(math.Ceil(float64(m.menuPageSize)/2)) - 1
		} else {
			m.menuStartColumn = (msg.Width - 20) / 2
			m.menuBottomRow = m.menuStartRow + m.menuPageSize - 1
		}

		// 菜单标题开始行、列
		m.menuTitleStartColumn = m.menuStartColumn
		if m.options.WhetherDisplayTitle && m.menuStartRow > 2 {
			if m.menuStartRow > 4 {
				m.menuTitleStartRow = m.menuStartRow - 3
			} else {
				m.menuTitleStartRow = 2
			}
		} else if !m.options.WhetherDisplayTitle && m.menuStartRow > 1 {
			if m.menuStartRow > 3 {
				m.menuTitleStartRow = m.menuStartRow - 3
			} else {
				m.menuTitleStartRow = 2
			}
		}

		// 组件更新
		for _, component := range m.components {
			if component == nil {
				continue
			}
			component.Update(msg, a)
		}
		return m, a.RerenderCmd(true)
	}

	return m, nil
}

func (m *Main) View(a *App) string {
	var windowHeight, windowWidth = a.WindowHeight(), a.WindowWidth()
	if windowHeight <= 0 || windowWidth <= 0 {
		return ""
	}

	var (
		builder strings.Builder
		top     int // 距离顶部的行数
	)

	// title
	if m.options.WhetherDisplayTitle {
		builder.WriteString(m.TitleView(a, &top))
	} else {
		top++
	}

	if !m.options.HideMenu {
		// menu title
		builder.WriteString(m.MenuTitleView(a, &top, nil))

		// menu list
		builder.WriteString(m.menuListView(a, &top))

		// search input
		builder.WriteString(m.searchInputView(a, &top))
	} else {
		builder.WriteString("\n\n\n")
		top += 2
	}

	// search bar
	if m.menu.IsSearchable() {
		builder.WriteString("\n\n")
		top += 2
	}

	// components view
	for _, component := range m.components {
		if component == nil {
			continue
		}
		var view, lines = component.View(a, m)
		builder.WriteString(view)
		top += lines
	}

	if top < windowHeight {
		builder.WriteString(strings.Repeat("\n", windowHeight-top-1))
	}

	return builder.String()
}

func (m *Main) MenuTitleStartColumn() int {
	return m.menuTitleStartColumn
}

func (m *Main) MenuTitleStartRow() int {
	return m.menuTitleStartRow
}

func (m *Main) MenuStartColumn() int {
	return m.menuStartColumn
}

func (m *Main) MenuStartRow() int {
	return m.menuStartRow
}

func (m *Main) SearchBarBottomRow() int {
	if m.menu.IsSearchable() {
		return m.menuBottomRow + 2
	}
	return m.menuBottomRow
}

func (m *Main) MenuBottomRow() int {
	return m.menuBottomRow
}

func (m *Main) IsDualColumn() bool {
	return m.isDualColumn
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

// TitleView title view
func (m *Main) TitleView(a *App, top *int) string {
	var (
		titleBuilder strings.Builder
		windowWidth  = a.WindowWidth()
	)
	titleLen := utf8.RuneCountInString(m.options.AppName) + 2
	prefixLen := (windowWidth - titleLen) / 2
	suffixLen := windowWidth - prefixLen - titleLen
	if prefixLen > 0 {
		titleBuilder.WriteString(strings.Repeat("─", prefixLen))
	}
	titleBuilder.WriteString(" ")
	titleBuilder.WriteString(m.options.AppName)
	titleBuilder.WriteString(" ")
	if suffixLen > 0 {
		titleBuilder.WriteString(strings.Repeat("─", suffixLen))
	}

	*top++

	return util.SetFgStyle(titleBuilder.String(), util.GetPrimaryColor())
}

// MenuTitleView menu title
func (m *Main) MenuTitleView(a *App, top *int, menuTitle *MenuItem) string {
	var (
		menuTitleBuilder strings.Builder
		title            string
		windowWidth      = a.WindowWidth()
		maxLen           = windowWidth - m.menuTitleStartColumn
	)

	if menuTitle == nil {
		menuTitle = m.menuTitle
	}

	realString := menuTitle.OriginString()
	formatString := menuTitle.String()
	if runewidth.StringWidth(realString) > maxLen {
		var menuTmp = *menuTitle
		titleLen := runewidth.StringWidth(menuTmp.Title)
		subTitleLen := runewidth.StringWidth(menuTmp.Subtitle)
		if titleLen >= maxLen-1 {
			menuTmp.Title = runewidth.Truncate(menuTmp.Title, maxLen-1, "")
			menuTmp.Subtitle = ""
		} else if subTitleLen >= maxLen-titleLen-1 {
			menuTmp.Subtitle = runewidth.Truncate(menuTmp.Subtitle, maxLen-titleLen-1, "")
		}
		title = menuTmp.String()
	} else {
		formatLen := runewidth.StringWidth(formatString)
		realLen := runewidth.StringWidth(realString)
		title = runewidth.FillRight(menuTitle.String(), maxLen+formatLen-realLen)
	}

	if top != nil && m.menuTitleStartRow-*top > 0 {
		menuTitleBuilder.WriteString(strings.Repeat("\n", m.menuTitleStartRow-*top))
	}
	if m.menuTitleStartColumn > 0 {
		menuTitleBuilder.WriteString(strings.Repeat(" ", m.menuTitleStartColumn))
	}
	menuTitleBuilder.WriteString(util.SetFgStyle(title, termenv.ANSIBrightGreen))

	if top != nil {
		*top = m.menuTitleStartRow
	}

	return menuTitleBuilder.String()
}

func (m *Main) MenuList() []MenuItem {
	return m.menuList
}

func (m *Main) menuListView(a *App, top *int) string {
	var menuListBuilder strings.Builder
	menus := m.getCurPageMenus()
	var lines, maxLines int
	if m.isDualColumn {
		lines = int(math.Ceil(float64(len(menus)) / 2))
		maxLines = int(math.Ceil(float64(m.menuPageSize) / 2))
	} else {
		lines = len(menus)
		maxLines = m.menuPageSize
	}

	if m.menuStartRow > *top {
		menuListBuilder.WriteString(strings.Repeat("\n", m.menuStartRow-*top))
	}

	var str string
	for i := 0; i < lines; i++ {
		str = m.menuLineView(a, i)
		menuListBuilder.WriteString(str)
		menuListBuilder.WriteString("\n")
	}

	// 补全空白
	if maxLines > lines {
		var windowWidth = a.WindowWidth()
		if windowWidth-m.menuStartColumn > 0 {
			menuListBuilder.WriteString(strings.Repeat(" ", windowWidth-m.menuStartColumn))
		}
		menuListBuilder.WriteString(strings.Repeat("\n", maxLines-lines))
	}

	*top = m.menuBottomRow

	return menuListBuilder.String()
}

func (m *Main) menuLineView(a *App, line int) string {
	var (
		menuLineBuilder strings.Builder
		index           int
		windowWidth     = a.WindowWidth()
	)
	if m.isDualColumn {
		index = line*2 + (m.menuCurPage-1)*m.menuPageSize
	} else {
		index = line + (m.menuCurPage-1)*m.menuPageSize
	}
	if index > len(m.menuList)-1 {
		index = len(m.menuList) - 1
	}
	if m.menuStartColumn > 4 {
		menuLineBuilder.WriteString(strings.Repeat(" ", m.menuStartColumn-4))
	}
	menuItemStr, menuItemLen := m.menuItemView(a, index)
	menuLineBuilder.WriteString(menuItemStr)
	if m.isDualColumn {
		var secondMenuItemLen int
		if index < len(m.menuList)-1 {
			var secondMenuItemStr string
			secondMenuItemStr, secondMenuItemLen = m.menuItemView(a, index+1)
			menuLineBuilder.WriteString(secondMenuItemStr)
		} else {
			menuLineBuilder.WriteString("    ")
			secondMenuItemLen = 4
		}
		if windowWidth-menuItemLen-secondMenuItemLen-m.menuStartColumn > 0 {
			menuLineBuilder.WriteString(strings.Repeat(" ", windowWidth-menuItemLen-secondMenuItemLen-m.menuStartColumn))
		}
	}

	return menuLineBuilder.String()
}

func (m *Main) menuItemView(a *App, index int) (string, int) {
	var (
		menuItemBuilder strings.Builder
		menuTitle       string
		itemMaxLen      int
		menuName        string
		windowWidth     = a.WindowWidth()
	)

	isSelected := !m.inSearching && index == m.selectedIndex

	if isSelected {
		menuTitle = fmt.Sprintf(" => %d. %s", index, m.menuList[index].Title)
	} else {
		menuTitle = fmt.Sprintf("    %d. %s", index, m.menuList[index].Title)
	}
	if len(m.menuList[index].Subtitle) != 0 {
		menuTitle += " "
	}

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

	menuTitleLen := runewidth.StringWidth(menuTitle)
	menuSubtitleLen := runewidth.StringWidth(m.menuList[index].Subtitle)

	var tmp string
	if menuTitleLen > itemMaxLen {
		tmp = runewidth.Truncate(menuTitle, itemMaxLen, "")
		tmp = runewidth.FillRight(tmp, itemMaxLen) // fix: 切割中文后缺少字符导致未对齐
		if isSelected {
			menuName = util.SetFgStyle(tmp, util.GetPrimaryColor())
		} else {
			menuName = util.SetNormalStyle(tmp)
		}
	} else if menuTitleLen+menuSubtitleLen > itemMaxLen {
		var r = []rune(m.menuList[index].Subtitle)
		r = append(r, []rune("   ")...)
		var i int
		if m.options.Ticker != nil {
			i = int(m.options.Ticker.PassedTime().Milliseconds()/500) % len(r)
		}
		var s = make([]rune, 0, itemMaxLen-menuTitleLen)
		for j := i; j < i+itemMaxLen-menuTitleLen; j++ {
			s = append(s, r[j%len(r)])
		}
		tmp = runewidth.Truncate(string(s), itemMaxLen-menuTitleLen, "")
		tmp = runewidth.FillRight(tmp, itemMaxLen-menuTitleLen)
		if isSelected {
			menuName = util.SetFgStyle(menuTitle, util.GetPrimaryColor()) + util.SetFgStyle(tmp, termenv.ANSIBrightBlack)
		} else {
			menuName = util.SetNormalStyle(menuTitle) + util.SetFgStyle(tmp, termenv.ANSIBrightBlack)
		}
	} else {
		tmp = runewidth.FillRight(m.menuList[index].Subtitle, itemMaxLen-menuTitleLen)
		if isSelected {
			menuName = util.SetFgStyle(menuTitle, util.GetPrimaryColor()) + util.SetFgStyle(tmp, termenv.ANSIBrightBlack)
		} else {
			menuName = util.SetNormalStyle(menuTitle) + util.SetFgStyle(tmp, termenv.ANSIBrightBlack)
		}
	}

	menuItemBuilder.WriteString(menuName)

	return menuItemBuilder.String(), itemMaxLen
}

func (m *Main) searchInputView(app *App, top *int) string {
	if !m.inSearching {
		*top++
		return "\n"
	}

	var (
		builder     strings.Builder
		windowWidth = app.WindowWidth()
	)
	builder.WriteString("\n")
	*top++

	inputs := []textinput.Model{
		m.searchInput,
	}

	var startColumn int
	if m.menuStartColumn > 2 {
		startColumn = m.menuStartColumn - 2
	}
	for i, input := range inputs {
		if startColumn > 0 {
			builder.WriteString(strings.Repeat(" ", startColumn))
		}

		builder.WriteString(input.View())

		var valueLen int
		if input.Value() == "" {
			valueLen = runewidth.StringWidth(input.Placeholder)
		} else {
			valueLen = runewidth.StringWidth(input.Value())
		}
		if spaceLen := windowWidth - startColumn - valueLen - 3; spaceLen > 0 {
			builder.WriteString(strings.Repeat(" ", spaceLen))
		}

		*top++

		if i < len(inputs)-1 {
			builder.WriteString("\n\n")
			*top++
		}
	}
	return builder.String()
}

func (m *Main) getCurPageMenus() []MenuItem {
	start := (m.menuCurPage - 1) * m.menuPageSize
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
		start := (m.menuCurPage - 1) * m.menuPageSize
		if start+num < len(m.menuList) {
			m.selectedIndex = start + num
		}
	case "g":
		newPage = m.MoveTop()
	case "G":
		newPage = m.MoveBottom()
	case "n", "N", "enter":
		newPage = m.EnterMenu(nil, nil)
	case "b", "B", "esc":
		newPage = m.BackMenu()
	case "r", "R":
		return m, a.RerenderCmd(true)
	case "/", "／":
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
	if newPage != nil {
		return newPage, func() tea.Msg { return newPage.Msg() }
	}
	if lastCmd == nil {
		lastCmd = a.Tick(time.Nanosecond)
	}
	return m, lastCmd
}

func (m *Main) searchMenuHandle() {
	m.inSearching = false
	var searchMenu = m.options.LocalSearchMenu
	if m.options.LocalSearchMenu == nil {
		searchMenu = DefaultSearchMenu()
	}
	searchMenu.Search(m.menu, m.searchInput.Value())
	m.EnterMenu(searchMenu, &MenuItem{Title: "搜索结果", Subtitle: m.searchInput.Value()})
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
			loading.start()
			if res, newPage = topHook(m); !res {
				loading.complete()
				return newPage
			}
			// update menu ui
			m.menuList = m.menu.MenuViews()
			loading.complete()
		}
		if m.selectedIndex-2 < 0 {
			return nil
		}
		m.selectedIndex -= 2
	} else {
		if m.selectedIndex-1 < 0 && topHook != nil {
			loading := NewLoading(m)
			loading.start()
			if res, newPage = topHook(m); !res {
				loading.complete()
				return newPage
			}
			m.menuList = m.menu.MenuViews()
			loading.complete()
		}
		if m.selectedIndex-1 < 0 {
			return nil
		}
		m.selectedIndex--
	}
	if m.selectedIndex < (m.menuCurPage-1)*m.menuPageSize {
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
	if m.isDualColumn {
		if m.selectedIndex+2 > len(m.menuList)-1 && bottomHook != nil {
			loading := NewLoading(m)
			loading.start()
			if res, newPage = bottomHook(m); !res {
				loading.complete()
				return newPage
			}
			m.menuList = m.menu.MenuViews()
			loading.complete()
		}
		if m.selectedIndex+2 > len(m.menuList)-1 {
			return nil
		}
		m.selectedIndex += 2
	} else {
		if m.selectedIndex+1 > len(m.menuList)-1 && bottomHook != nil {
			loading := NewLoading(m)
			loading.start()
			if res, newPage = bottomHook(m); !res {
				loading.complete()
				return newPage
			}
			m.menuList = m.menu.MenuViews()
			loading.complete()
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
		loading.start()
		if res, newPage = bottomHook(m); !res {
			loading.complete()
			return newPage
		}
		m.menuList = m.menu.MenuViews()
		loading.complete()
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
		loading.start()
		if res, newPage = prePageHook(m); !res {
			loading.complete()
			return newPage
		}
		loading.complete()
	}
	if m.menuCurPage <= 1 {
		return nil
	}
	m.menuCurPage--
	return newPage
}

func (m *Main) NextPage() Page {
	var (
		res     bool
		newPage Page
	)
	if nextPageHook := m.menu.BeforeNextPageHook(); nextPageHook != nil {
		loading := NewLoading(m)
		loading.start()
		if res, newPage = nextPageHook(m); !res {
			loading.complete()
			return newPage
		}
		loading.complete()
	}
	if m.menuCurPage >= int(math.Ceil(float64(len(m.menuList))/float64(m.menuPageSize))) {
		return nil
	}

	m.menuCurPage++
	return newPage
}

func (m *Main) EnterMenu(newMenu Menu, newTitle *MenuItem) Page {
	if (newMenu == nil || newTitle == nil) && m.selectedIndex >= len(m.menuList) {
		return nil
	}

	if newMenu == nil {
		newMenu = m.menu.SubMenu(m.app, m.selectedIndex)
	}
	if newTitle == nil {
		newTitle = &m.menuList[m.selectedIndex]
	}

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
		loading.start()
		if res, newPage = enterMenuHook(m); !res {
			loading.complete()
			m.menuStack.Pop() // 压入的重新弹出
			return newPage
		}
		loading.complete()
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

	var (
		stackItem = m.menuStack.Pop()
		newPage   Page
		res       bool
	)
	if backMenuHook := m.menu.BeforeBackMenuHook(); backMenuHook != nil {
		loading := NewLoading(m)
		loading.start()
		if res, newPage = backMenuHook(m); !res {
			loading.complete()
			m.menuStack.Push(stackItem) // 弹出的重新压入
			return newPage
		}
		loading.complete()
	}
	m.menu.FormatMenuItem(m.menuTitle) // 重新格式化

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

func TickMain(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return tickMainMsg{}
	})
}
