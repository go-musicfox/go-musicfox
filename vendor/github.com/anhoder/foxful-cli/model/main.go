package model

import (
	"fmt"
	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
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
	m.searchInput.Placeholder = " " + SearchPlaceholder
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

		// menu start col, row
		m.menuStartRow = msg.Height / 3
		// Height of the bottom part of the music player. Used in calculating the number of rows left.
		// 3 lines for search + 5 lines of lyrics + 6 lines of song name and progress bar = 14. But somehow
		// 13 works better.
		bottomHeight := 13
		// If dynamic row count is on, we may want to adjust menuStartRow
		if m.options.DynamicRowCount {
			if m.options.MaxMenuStartRow > 0 {
				// Limit menuStartRow to user-defined value
				if m.menuStartRow > m.options.MaxMenuStartRow {
					m.menuStartRow = m.options.MaxMenuStartRow
				}
			}
		}

		if !m.options.WhetherDisplayTitle && m.menuStartRow > 1 {
			m.menuStartRow--
		}

		if m.options.DynamicRowCount {
			maxEntries := (msg.Height - m.menuStartRow - bottomHeight) * m.getNumColumns()
			if maxEntries > 10 {
				m.menuPageSize = maxEntries
			} else {
				m.menuPageSize = 10
			}
		}

		if m.isDualColumn {
			m.menuStartColumn = (msg.Width - 60) / 2
			m.menuBottomRow = m.menuStartRow + int(math.Ceil(float64(m.menuPageSize)/2)) + 1 // 1 for search bar
		} else {
			m.menuStartColumn = (msg.Width - 20) / 2
			m.menuBottomRow = m.menuStartRow + m.menuPageSize + 1 // 1 for search bar
		}

		// menu title satrt col, row
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

		// components
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
	windowHeight, windowWidth := a.WindowHeight(), a.WindowWidth()
	if windowHeight <= 0 || windowWidth <= 0 {
		return ""
	}

	var (
		builder strings.Builder
		top     int // num of rows from top
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
	builder.WriteString("\n\n")
	top += 2

	// components view
	for _, component := range m.components {
		if component == nil {
			continue
		}
		view, lines := component.View(a, m)
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

func (m *Main) MenuBottomRow() int {
	return m.menuBottomRow
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

// TitleView title view
func (m *Main) TitleView(a *App, top *int) string {
	var (
		titleBuilder strings.Builder
		windowWidth  = a.WindowWidth()
	)
	appName := " " + m.options.AppName + " "
	titleLen := runewidth.StringWidth(appName)
	prefixLen := (windowWidth - titleLen) / 2
	suffixLen := windowWidth - prefixLen - titleLen
	if prefixLen > 0 {
		titleBuilder.WriteString(strings.Repeat("─", prefixLen))
	}
	titleBuilder.WriteString(appName)
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

	if top != nil && m.menuTitleStartRow-*top > 0 {
		menuTitleBuilder.WriteString(strings.Repeat("\n", m.menuTitleStartRow-*top))
	}

	realString := menuTitle.OriginString()
	formatString := menuTitle.String()
	if m.options.CenterEverything {
		stringLen := runewidth.StringWidth(realString)
		if stringLen >= windowWidth {
			title = runewidth.Truncate(formatString, windowWidth, "")
		} else {
			spaceLeft := (windowWidth - stringLen) / 2
			spaceRight := windowWidth - spaceLeft - stringLen
			title = strings.Repeat(" ", spaceLeft) + formatString + strings.Repeat(" ", spaceRight)
		}
	} else {
		if runewidth.StringWidth(realString) > maxLen {
			menuTmp := *menuTitle
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
		if m.menuTitleStartColumn > 0 {
			menuTitleBuilder.WriteString(strings.Repeat(" ", m.menuTitleStartColumn))
		}
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

func (m *Main) getNumColumns() int {
	if m.isDualColumn {
		return 2
	}
	return 1
}

func (m *Main) forceEntryLength(item *MenuItem, targetLength int) string {
	// Case 1:
	// Only enough space for the main title. Not enough width for subtitle.
	titleWidth := runewidth.StringWidth(item.Title)
	minSubtitleWidth := 5
	if titleWidth >= targetLength-minSubtitleWidth {
		title := runewidth.Truncate(item.Title, targetLength, "")
		return runewidth.FillRight(title, targetLength)
	}
	// Case 2:
	// Enough space for everything.
	fullWidth := runewidth.StringWidth(item.OriginString())
	if fullWidth <= targetLength {
		return item.String() + strings.Repeat(" ", targetLength-fullWidth)
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
		s = append(s, r[i])
		currentWidth += runewidth.RuneWidth(r[i])
	}
	// Truncate in case a character of width 2 goes over the limit
	subtitle := runewidth.Truncate(string(s), subtitleSpace, "")
	// Fill with space in case we have 1 space remaining but the next rune has width 2
	subtitle = runewidth.FillRight(subtitle, subtitleSpace)
	return item.Title + " " + util.SetFgStyle(subtitle, termenv.ANSIBrightBlack)
}

func (m *Main) formatEntry(item *MenuItem, index int, targetLength int) string {
	if item == nil {
		return strings.Repeat(" ", targetLength)
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
		return util.SetFgStyle(songEntry, util.GetPrimaryColor())
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
			length := runewidth.StringWidth(menuItem.OriginString())
			titleLengths = append(titleLengths, length)
			allSongs = append(allSongs, &menuItem)
		} else {
			allSongs = append(allSongs, nil)
		}
	}
	allSongs = append(allSongs, nil)

	slices.Sort(titleLengths)
	maxSongTitleLength := 0
	if len(titleLengths) > 0 {
		maxSongTitleLength = titleLengths[len(titleLengths)-1]
	}
	if len(titleLengths) >= 6 && maxSongTitleLength >= 30 {
		// Drop the longest 30% of all titles to prevent the menu from being stretched too long due to outliers
		maxSongTitleLength = titleLengths[int32(0.7*float32(len(titleLengths)))]
		if maxSongTitleLength < 30 {
			maxSongTitleLength = 30
		}
	}

	// Songs have 4 spaces built-in at the front, so we need 4 columns on the right side to balance spaces
	remainingWindowWidth := a.windowWidth - 4

	// Extra padding applied to every segment.
	// If the window is wide, we want more padding.
	extraPadding := (a.windowWidth - 40) / 5
	if extraPadding < 0 {
		extraPadding = 0
	}
	remainingWindowWidth -= extraPadding

	itemMaxLength := remainingWindowWidth / m.getNumColumns()

	entryLength := maxSongTitleLength + 6 + m.getMaxIndexWidth()
	if entryLength > itemMaxLength {
		entryLength = itemMaxLength
	}

	// 4 is the correction
	paddingLength := a.windowWidth - entryLength*m.getNumColumns() - 4
	var (
		leftProportion   = 0.5
		middleProportion = 0.0
		// rightProportion  = 0.5
	)
	if m.IsDualColumn() {
		leftProportion = 0.45
		middleProportion = 0.1
		// rightProportion = 0.45
	}
	paddingLeft := int(math.Round(float64(paddingLength) * leftProportion))
	paddingLength -= paddingLeft
	paddingMiddle := int(math.Round(float64(paddingLength) * middleProportion))
	paddingLength -= paddingMiddle
	paddingRight := paddingLength + 4

	var result strings.Builder
	for i := 0; i < lines; i++ {
		index := i * m.getNumColumns()
		menuIndex := m.getPageStartIndex() + index
		result.WriteString(strings.Repeat(" ", paddingLeft))
		result.WriteString(m.formatEntry(allSongs[index], menuIndex, entryLength))
		if m.isDualColumn {
			result.WriteString(strings.Repeat(" ", paddingMiddle))
			result.WriteString(m.formatEntry(allSongs[index+1], menuIndex+1, entryLength))
		}
		result.WriteString(strings.Repeat(" ", paddingRight))
		result.WriteString("\n")
	}
	return result.String()
}

func (m *Main) menuListView(a *App, top *int) string {
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

	if m.menuStartRow > *top {
		menuListBuilder.WriteString(strings.Repeat("\n", m.menuStartRow-*top))
	}

	if m.options.CenterEverything {
		menuListBuilder.WriteString(m.centeredMenuView(a, lines))
	} else {
		for i := 0; i < lines; i++ {
			str := m.menuLineView(a, i)
			menuListBuilder.WriteString(str)
			menuListBuilder.WriteString("\n")
		}
	}

	// fill blanks
	if maxLines > lines {
		windowWidth := a.WindowWidth()
		if windowWidth-m.menuStartColumn > 0 {
			menuListBuilder.WriteString(strings.Repeat(" ", windowWidth-m.menuStartColumn))
		}
		menuListBuilder.WriteString(strings.Repeat("\n", maxLines-lines))
	}

	*top = m.menuBottomRow

	return menuListBuilder.String()
}

func (m *Main) getPageStartIndex() int {
	return (m.menuCurPage - 1) * m.menuPageSize
}

func (m *Main) menuLineView(a *App, line int) string {
	var (
		menuLineBuilder strings.Builder
		index           int
		windowWidth     = a.WindowWidth()
	)
	if m.isDualColumn {
		index = line*2 + m.getPageStartIndex()
	} else {
		index = line + m.getPageStartIndex()
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

func (m *Main) getMaxIndexWidth() int {
	return int(math.Log10(float64((m.menuPageSize*m.menuCurPage)-1))) + 1
}

func (m *Main) isSelected(index int) bool {
	return !m.inSearching && index == m.selectedIndex
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

	if isSelected {
		menuTitle = fmt.Sprintf(fmt.Sprintf(" => %%%dd. %%s", maxIndexWidth), index, m.menuList[index].Title)
	} else {
		menuTitle = fmt.Sprintf(fmt.Sprintf("    %%%dd. %%s", maxIndexWidth), index, m.menuList[index].Title)
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
		r := []rune(m.menuList[index].Subtitle)
		r = append(r, []rune("   ")...)
		var i int
		if m.options.Ticker != nil {
			i = int(m.options.Ticker.PassedTime().Milliseconds()/500) % len(r)
		}
		s := make([]rune, 0, itemMaxLen-menuTitleLen)
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
			newPage = m.EnterMenu(nil, nil)
		} else {
			m.selectedIndex = target
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

func TickMain(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return tickMainMsg{}
	})
}
