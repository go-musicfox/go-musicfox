package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/pkg/player"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

// PageType 显示模型的类型
type PageType uint8

const (
	PtMain   PageType = iota // 主页面
	PtLogin                  // 登录页面
	PtSearch                 // 搜索页面
)

func MsgOfPageType(t PageType) tea.Msg {
	switch t {
	case PtMain:
		return tickMainUIMsg{}
	case PtLogin:
		return tickLoginMsg{}
	case PtSearch:
		return tickSearchMsg{}
	}
	return nil
}

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
	return item.Title + " " + SetFgStyle(item.Subtitle, termenv.ANSIBrightBlack)
}

type MainUIModel struct {
	doubleColumn bool // 是否双列显示

	menuTitle            *MenuItem // 菜单标题
	menuTitleStartRow    int       // 菜单标题开始行
	menuTitleStartColumn int       // 菜单标题开始列

	menuStartRow    int // 菜单开始行
	menuStartColumn int // 菜单开始列
	menuBottomRow   int // 菜单最底部所在行

	menuCurPage  int // 菜单当前页
	menuPageSize int // 菜单每页大小

	menuList      []MenuItem   // 菜单列表
	menuStack     *utils.Stack // 菜单栈
	selectedIndex int          // 当前选中的菜单index

	inSearching bool            // 搜索菜单
	searchInput textinput.Model // 搜索输入框

	pageType PageType // 显示的页面类型

	menu   Menu    // 菜单
	player *Player // 播放器
}

func (main *MainUIModel) Close() {
	main.player.Close()
}

func NewMainUIModel(parentModel *NeteaseModel) (m *MainUIModel) {
	m = new(MainUIModel)

	m.menuTitle = &MenuItem{Title: "网易云音乐"}
	m.player = NewPlayer(parentModel)
	m.menu = NewMainMenu(parentModel)
	m.menuList = m.menu.MenuViews()
	m.menuStack = new(utils.Stack)
	m.menuCurPage = 1
	m.menuPageSize = 10
	m.selectedIndex = 0

	m.searchInput = textinput.New()
	m.searchInput.Placeholder = " 搜索"
	m.searchInput.Prompt = GetFocusedPrompt()
	m.searchInput.TextStyle = GetPrimaryFontStyle()
	m.searchInput.CharLimit = 32

	return
}

func (main *MainUIModel) refreshMenuList() {
	main.menuList = main.menu.MenuViews()
}

func (main *MainUIModel) refreshMenuTitle() {
	main.menu.FormatMenuItem(main.menuTitle)
}

// update main ui
func (main *MainUIModel) update(message tea.Msg, m *NeteaseModel) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return main.keyMsgHandle(msg, m)
	case tea.MouseMsg:
		return main.mouseMsgHandle(msg, m)
	case tickMainUIMsg:
		return m, nil
	case tea.WindowSizeMsg:
		m.doubleColumn = msg.Width >= 75 && configs.ConfigRegistry.MainDoubleColumn

		// 菜单开始行、列
		m.menuStartRow = msg.Height / 3
		if !configs.ConfigRegistry.MainShowTitle && m.menuStartRow > 1 {
			m.menuStartRow--
		}
		if m.doubleColumn {
			m.menuStartColumn = (msg.Width - 60) / 2
			m.menuBottomRow = m.menuStartRow + int(math.Ceil(float64(m.menuPageSize)/2)) - 1
		} else {
			m.menuStartColumn = (msg.Width - 20) / 2
			m.menuBottomRow = m.menuStartRow + m.menuPageSize - 1
		}

		// 菜单标题开始行、列
		m.menuTitleStartColumn = m.menuStartColumn
		if configs.ConfigRegistry.MainShowTitle && m.menuStartRow > 2 {
			if m.menuStartRow > 4 {
				m.menuTitleStartRow = m.menuStartRow - 3
			} else {
				m.menuTitleStartRow = 2
			}
		} else if !configs.ConfigRegistry.MainShowTitle && m.menuStartRow > 1 {
			if m.menuStartRow > 3 {
				m.menuTitleStartRow = m.menuStartRow - 3
			} else {
				m.menuTitleStartRow = 2
			}
		}

		// 播放器歌词
		spaceHeight := m.WindowHeight - 4 - m.menuBottomRow
		if spaceHeight < 4 || !configs.ConfigRegistry.MainShowLyric {
			// 不显示歌词
			m.player.showLyric = false
		} else {
			m.player.showLyric = true
			if spaceHeight > 6 {
				// 5行歌词
				m.player.lyricStartRow = (m.WindowHeight-3+m.menuBottomRow)/2 - 3
				m.player.lyricLines = 5
			} else {
				// 3行歌词
				m.player.lyricStartRow = (m.WindowHeight-3+m.menuBottomRow)/2 - 2
				m.player.lyricLines = 3
			}
		}
		return m, m.rerenderTicker(true)
	}

	return m, nil
}

// get main ui view
func (main *MainUIModel) view(m *NeteaseModel) string {
	if m.WindowWidth <= 0 || m.WindowHeight <= 0 {
		return ""
	}

	var builder strings.Builder

	// 距离顶部的行数
	top := 0

	// title
	if configs.ConfigRegistry.MainShowTitle {
		builder.WriteString(main.titleView(m, &top))
	} else {
		top++
	}

	// menu title
	builder.WriteString(main.menuTitleView(m, &top, nil))

	// menu list
	builder.WriteString(main.menuListView(m, &top))

	// search input
	builder.WriteString(main.searchInputView(m, &top))

	// player view
	builder.WriteString(m.player.playerView(&top))

	if top < m.WindowHeight {
		builder.WriteString(strings.Repeat("\n", m.WindowHeight-top-1))
	}

	return builder.String()
}

// title view
func (main *MainUIModel) titleView(m *NeteaseModel, top *int) string {
	var titleBuilder strings.Builder
	titleLen := utf8.RuneCountInString(constants.AppName) + 2
	prefixLen := (m.WindowWidth - titleLen) / 2
	suffixLen := m.WindowWidth - prefixLen - titleLen
	if prefixLen > 0 {
		titleBuilder.WriteString(strings.Repeat("─", prefixLen))
	}
	titleBuilder.WriteString(" ")
	titleBuilder.WriteString(constants.AppName)
	titleBuilder.WriteString(" ")
	if suffixLen > 0 {
		titleBuilder.WriteString(strings.Repeat("─", suffixLen))
	}

	*top++

	return SetFgStyle(titleBuilder.String(), GetPrimaryColor())
}

// menu title
func (main *MainUIModel) menuTitleView(m *NeteaseModel, top *int, menuTitle *MenuItem) string {
	var (
		menuTitleBuilder strings.Builder
		title            string
		maxLen           = m.WindowWidth - m.menuTitleStartColumn
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
	menuTitleBuilder.WriteString(SetFgStyle(title, termenv.ANSIBrightGreen))

	if top != nil {
		*top = m.menuTitleStartRow
	}

	return menuTitleBuilder.String()
}

// 菜单列表
func (main *MainUIModel) menuListView(m *NeteaseModel, top *int) string {
	var menuListBuilder strings.Builder
	menus := main.getCurPageMenus(m)
	var lines, maxLines int
	if m.doubleColumn {
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
		str = main.menuLineView(m, i)
		menuListBuilder.WriteString(str)
		menuListBuilder.WriteString("\n")
	}

	// 补全空白
	if maxLines > lines {
		if m.WindowWidth-m.menuStartColumn > 0 {
			menuListBuilder.WriteString(strings.Repeat(" ", m.WindowWidth-m.menuStartColumn))
		}
		menuListBuilder.WriteString(strings.Repeat("\n", maxLines-lines))
	}

	*top = m.menuBottomRow

	return menuListBuilder.String()
}

// 菜单Line
func (main *MainUIModel) menuLineView(m *NeteaseModel, line int) string {
	var menuLineBuilder strings.Builder
	var index int
	if m.doubleColumn {
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
	menuItemStr, menuItemLen := main.menuItemView(m, index)
	menuLineBuilder.WriteString(menuItemStr)
	if m.doubleColumn {
		var secondMenuItemLen int
		if index < len(m.menuList)-1 {
			var secondMenuItemStr string
			secondMenuItemStr, secondMenuItemLen = main.menuItemView(m, index+1)
			menuLineBuilder.WriteString(secondMenuItemStr)
		} else {
			menuLineBuilder.WriteString("    ")
			secondMenuItemLen = 4
		}
		if m.WindowWidth-menuItemLen-secondMenuItemLen-m.menuStartColumn > 0 {
			menuLineBuilder.WriteString(strings.Repeat(" ", m.WindowWidth-menuItemLen-secondMenuItemLen-m.menuStartColumn))
		}
	}

	return menuLineBuilder.String()
}

// 菜单Item
func (main *MainUIModel) menuItemView(m *NeteaseModel, index int) (string, int) {
	var (
		menuItemBuilder strings.Builder
		menuTitle       string
		itemMaxLen      int
		menuName        string
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

	if m.doubleColumn {
		if m.WindowWidth <= 88 {
			itemMaxLen = (m.WindowWidth - m.menuStartColumn - 4) / 2
		} else {
			if index%2 == 0 {
				itemMaxLen = 44
			} else {
				itemMaxLen = m.WindowWidth - m.menuStartColumn - 44
			}
		}
	} else {
		itemMaxLen = m.WindowWidth - m.menuStartColumn
	}

	menuTitleLen := runewidth.StringWidth(menuTitle)
	menuSubtitleLen := runewidth.StringWidth(m.menuList[index].Subtitle)

	var tmp string
	if menuTitleLen > itemMaxLen {
		tmp = runewidth.Truncate(menuTitle, itemMaxLen, "")
		tmp = runewidth.FillRight(tmp, itemMaxLen) // fix: 切割中文后缺少字符导致未对齐
		if isSelected {
			menuName = SetFgStyle(tmp, GetPrimaryColor())
		} else {
			menuName = SetNormalStyle(tmp)
		}
	} else if menuTitleLen+menuSubtitleLen > itemMaxLen {
		var r = []rune(m.menuList[index].Subtitle)
		r = append(r, []rune("   ")...)
		i := int(m.player.PassedTime().Milliseconds()/500) % len(r) // 使用播放时间控制，暂停保持滚动位置
		var s = make([]rune, 0, itemMaxLen-menuTitleLen)
		for j := i; j < i+itemMaxLen-menuTitleLen; j++ {
			s = append(s, r[j%len(r)])
		}
		tmp = runewidth.Truncate(string(s), itemMaxLen-menuTitleLen, "")
		tmp = runewidth.FillRight(tmp, itemMaxLen-menuTitleLen)
		if isSelected {
			menuName = SetFgStyle(menuTitle, GetPrimaryColor()) + SetFgStyle(tmp, termenv.ANSIBrightBlack)
		} else {
			menuName = SetNormalStyle(menuTitle) + SetFgStyle(tmp, termenv.ANSIBrightBlack)
		}
	} else {
		tmp = runewidth.FillRight(m.menuList[index].Subtitle, itemMaxLen-menuTitleLen)
		if isSelected {
			menuName = SetFgStyle(menuTitle, GetPrimaryColor()) + SetFgStyle(tmp, termenv.ANSIBrightBlack)
		} else {
			menuName = SetNormalStyle(menuTitle) + SetFgStyle(tmp, termenv.ANSIBrightBlack)
		}
	}

	menuItemBuilder.WriteString(menuName)

	return menuItemBuilder.String(), itemMaxLen
}

// 菜单搜索
func (main *MainUIModel) searchInputView(m *NeteaseModel, top *int) string {
	if !main.inSearching {
		*top++
		return "\n"
	}

	var builder strings.Builder
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
		if spaceLen := m.WindowWidth - startColumn - valueLen - 3; spaceLen > 0 {
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

// 获取当前页的菜单
func (main *MainUIModel) getCurPageMenus(m *NeteaseModel) []MenuItem {
	start := (m.menuCurPage - 1) * m.menuPageSize
	end := int(math.Min(float64(len(m.menuList)), float64(m.menuCurPage*m.menuPageSize)))

	return m.menuList[start:end]
}

// key handle
func (main *MainUIModel) keyMsgHandle(msg tea.KeyMsg, m *NeteaseModel) (tea.Model, tea.Cmd) {
	if !m.isListeningKey {
		return m, nil
	}

	if m.inSearching {
		switch msg.String() {
		case "esc":
			m.inSearching = false
			m.searchInput.Blur()
			m.searchInput.Reset()
			return m, m.rerenderTicker(true)
		case "enter":
			searchMenuHandle(m)
			return m, m.rerenderTicker(true)
		}

		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)

		return m, tea.Batch(cmd)
	}

	key := msg.String()
	switch key {
	case "j", "J", "down":
		moveDown(m)
	case "k", "K", "up":
		moveUp(m)
	case "h", "H", "left":
		moveLeft(m)
	case "l", "L", "right":
		moveRight(m)
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		num, _ := strconv.Atoi(key)
		start := (m.menuCurPage - 1) * m.menuPageSize
		if start+num < len(m.menuList) {
			m.selectedIndex = start + num
		}
	case "g":
		moveTop(m)
	case "G":
		moveBottom(m)
	case "n", "N":
		enterMenu(m, nil, nil)
	case "enter":
		enterKeyHandle(m)
	case "b", "B", "esc":
		backMenu(m)
	case "c", "C":
		if _, ok := m.menu.(*CurPlaylist); !ok {
			var subTitle string
			if !m.player.playlistUpdateAt.IsZero() {
				subTitle = m.player.playlistUpdateAt.Format("[更新于2006-01-02 15:04:05]")
			}
			enterMenu(m, NewCurPlaylist(m.player.playlist), &MenuItem{Title: "当前播放列表", Subtitle: subTitle})
			m.player.LocatePlayingSong()
		}
	case " ", "　":
		spaceKeyHandle(m)
	case "v":
		m.player.Seek(m.player.PassedTime() + time.Second*5)
	case "V":
		m.player.Seek(m.player.PassedTime() + time.Second*10)
	case "x":
		m.player.Seek(m.player.PassedTime() - time.Second*1)
	case "X":
		m.player.Seek(m.player.PassedTime() - time.Second*5)
	case "[", "【":
		m.player.PreviousSong(true)
	case "]", "】":
		m.player.NextSong(true)
	case "p":
		m.player.SetPlayMode(0)
	case "P":
		m.player.Intelligence(false)
	case ",", "，":
		// like playing song
		likePlayingSong(m, true)
	case ".", "。":
		// unlike playing song
		likePlayingSong(m, false)
	case "w", "W":
		// logout and quit
		logout()
		m.startup.quitting = true
		return m, tea.Quit
	case "-", "−", "ー": // half-width, full-width and katakana
		m.player.DownVolume()
	case "=", "＝":
		m.player.UpVolume()
	case "d":
		downloadPlayingSong(m)
	case "D":
		downloadSelectedSong(m)
	case "t":
		// trash playing song
		trashPlayingSong(m)
	case "T":
		// trash selected song
		trashSelectedSong(m)
	case "<", "〈", "＜", "《", "«": // half-width, full-width, Japanese, Chinese and French
		// like selected song
		likeSelectedSong(m, true)
	case ">", "〉", "＞", "》", "»":
		// unlike selected song
		likeSelectedSong(m, false)
	case "/", "／":
		// 搜索菜单
		if m.menu.IsSearchable() {
			m.inSearching = true
			m.searchInput.Focus()
		}
	case "?", "？":
		// 帮助
		enterMenu(m, NewHelpMenu(), &MenuItem{Title: "帮助"})
	case "tab":
		openAddSongToUserPlaylistMenu(m, true, true)
	case "shift+tab":
		openAddSongToUserPlaylistMenu(m, true, false)
	case "`":
		openAddSongToUserPlaylistMenu(m, false, true)
	case "~", "～":
		openAddSongToUserPlaylistMenu(m, false, false)
	case "a":
		// 当前歌曲所属专辑
		albumOfPlayingSong(m)
	case "A":
		// 选中歌曲所属专辑
		albumOfSelectedSong(m)
	case "s":
		// 当前歌曲所属歌手
		artistOfPlayingSong(m)
	case "S":
		// 选中歌曲所属歌手
		artistOfSelectedSong(m)
	case "o":
		// 网页打开当前歌曲
		openPlayingSongInWeb(m)
	case "O":
		// 网页打开选中项
		openSelectedItemInWeb(m)
	case ";", ":", "：", "；":
		// 收藏选中歌单
		collectSelectedPlaylist(m, true)
	case "'", "\"":
		// 取消收藏选中歌单
		collectSelectedPlaylist(m, false)
	case "\\", "、":
		// 从播放列表删除歌曲,仅在当前播放列表界面有效
		delSongFromPlaylist(m)
	case "e":
		// 追加到下一曲播放
		addSongToPlaylist(m, true)
	case "E":
		// 追加到播放列表末尾
		addSongToPlaylist(m, false)
	case "r", "R":
		// rerender
		return m, m.rerenderTicker(true)
	}

	return m, tickMainUI(time.Nanosecond)
}

// mouse handle
func (main *MainUIModel) mouseMsgHandle(msg tea.MouseMsg, m *NeteaseModel) (tea.Model, tea.Cmd) {
	if !m.isListeningKey {
		return m, nil
	}
	switch msg.Type {
	case tea.MouseLeft:
		x, y := msg.X, msg.Y
		w := len(m.player.progressRamp)
		if y+1 == m.WindowHeight && x+1 <= len(m.player.progressRamp) {
			allDuration := int(m.player.CurMusic().Duration.Seconds())
			if allDuration == 0 {
				return m, nil
			}
			duration := float64(x) * m.player.CurMusic().Duration.Seconds() / float64(w)
			m.player.Seek(time.Second * time.Duration(duration))
			if m.player.State() != player.Playing {
				m.player.Resume()
			}
		}
	case tea.MouseWheelDown:
		m.player.DownVolume()
	case tea.MouseWheelUp:
		m.player.UpVolume()
	}

	return m, tickMainUI(time.Nanosecond)
}
