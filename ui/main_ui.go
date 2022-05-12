package ui

import (
    "fmt"
    tea "github.com/anhoder/bubbletea"
    "github.com/mattn/go-runewidth"
    "github.com/muesli/termenv"
    "go-musicfox/config"
    "go-musicfox/constants"
    "go-musicfox/utils"
    "math"
    "strings"
    "time"
    "unicode/utf8"
)

// ModelType 显示模型的类型
type ModelType uint8

const (
    MtMain   ModelType = iota // 主页面
    MtLogin                   // 登录页面
    MtSearch                  // 搜索页面
)

type MenuItem struct {
    Title    string
    Subtitle string
}

type mainUIModel struct {
    doubleColumn bool // 是否双列显示

    menuTitle            string // 菜单标题
    menuTitleStartRow    int    // 菜单标题开始行
    menuTitleStartColumn int    // 菜单标题开始列

    menuStartRow    int // 菜单开始行
    menuStartColumn int // 菜单开始列
    menuBottomRow   int // 菜单最底部所在行

    menuCurPage  int // 菜单当前页
    menuPageSize int // 菜单每页大小

    menuList      []MenuItem   // 菜单列表
    menuStack     *utils.Stack // 菜单栈
    selectedIndex int          // 当前选中的菜单index

    showLogin bool      // 显示登陆
    modelType ModelType // 显示的页面类型

    menu   IMenu   // 菜单
    player *Player // 播放器
}

func NewMainUIModel(parentModel *NeteaseModel) (m *mainUIModel) {
    m = new(mainUIModel)
    m.menuTitle = "网易云音乐"
    m.player = NewPlayer(parentModel)
    m.menu = NewMainMenu()
    m.menuList = m.menu.MenuViews()
    m.menuStack = new(utils.Stack)
    m.menuCurPage = 1
    m.menuPageSize = 10
    m.selectedIndex = 0

    return
}

// update main ui
func updateMainUI(msg tea.Msg, m *NeteaseModel) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        return keyMsgHandle(msg, m)

    case tea.ClearScreenMsg:
        return m, tickMainUI(time.Nanosecond)

    case tickMainUIMsg:
        return m, nil

    case tea.WindowSizeMsg:
        m.doubleColumn = msg.Width >= 75

        // 菜单开始行、列
        m.menuStartRow = msg.Height / 3
        if !config.ConfigRegistry.MainShowTitle && m.menuStartRow > 1 {
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
        if config.ConfigRegistry.MainShowTitle && m.menuStartRow > 2 {
            if m.menuStartRow > 4 {
                m.menuTitleStartRow = m.menuStartRow - 3
            } else {
                m.menuTitleStartRow = 2
            }
        } else if !config.ConfigRegistry.MainShowTitle && m.menuStartRow > 1 {
            if m.menuStartRow > 3 {
                m.menuTitleStartRow = m.menuStartRow - 3
            } else {
                m.menuTitleStartRow = 2
            }
        }

        // 播放器歌词
        spaceHeight := m.WindowHeight - 4 - m.menuBottomRow
        if spaceHeight < 4 {
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

    }

    return m, nil
}

// get main ui view
func mainUIView(m *NeteaseModel) string {
    if m.WindowWidth <= 0 || m.WindowHeight <= 0 {
        return ""
    }

    var builder strings.Builder

    // 距离顶部的行数
    top := 0

    // title
    if config.ConfigRegistry.MainShowTitle {
        builder.WriteString(titleView(m, &top))
    } else {
        top++
    }

    // menu title
    builder.WriteString(menuTitleView(m, &top, ""))

    // menu list
    builder.WriteString(menuListView(m, &top))

    // player view
    builder.WriteString(m.player.playerView(&top))

    if top < m.WindowHeight {
        builder.WriteString(strings.Repeat("\n", m.WindowHeight-top-1))
    }

    return builder.String()
}

// title view
func titleView(m *NeteaseModel, top *int) string {
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
func menuTitleView(m *NeteaseModel, top *int, menuTitle string) string {
    var (
        menuTitleBuilder strings.Builder
        title            string
        maxLen           = 50
    )
    if maxLen > m.WindowWidth-m.menuTitleStartColumn {
        maxLen = m.WindowWidth - m.menuTitleStartColumn
    }

    if len(menuTitle) <= 0 {
        menuTitle = m.menuTitle
    }

    if runewidth.StringWidth(menuTitle) > maxLen {
        title = runewidth.Truncate(menuTitle, maxLen, "")
    } else {
        title = runewidth.FillRight(menuTitle, maxLen)
    }

    if m.menuTitleStartRow-*top > 0 {
        menuTitleBuilder.WriteString(strings.Repeat("\n", m.menuTitleStartRow-*top))
    }
    if m.menuTitleStartColumn > 0 {
        menuTitleBuilder.WriteString(strings.Repeat(" ", m.menuTitleStartColumn))
    }
    menuTitleBuilder.WriteString(SetFgStyle(title, termenv.ANSIBrightGreen))

    *top = m.menuTitleStartRow

    return menuTitleBuilder.String()
}

// 菜单列表
func menuListView(m *NeteaseModel, top *int) string {
    var menuListBuilder strings.Builder
    menus := getCurPageMenus(m)
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
        str = menuLineView(m, i)
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
func menuLineView(m *NeteaseModel, line int) string {
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
    menuLineBuilder.WriteString(menuItemView(m, index))
    if m.doubleColumn {
        if index < len(m.menuList)-1 {
            menuLineBuilder.WriteString(menuItemView(m, index+1))
        } else {
            menuLineBuilder.WriteString("    ")
        }
    }

    return menuLineBuilder.String()
}

// 菜单Item
func menuItemView(m *NeteaseModel, index int) string {
    var (
        menuItemBuilder strings.Builder
        menuTitle       string
        itemMaxLen      int
        menuName        string
    )

    if index == m.selectedIndex {
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
        if index == m.selectedIndex {
            menuName = SetFgStyle(tmp, GetPrimaryColor())
        } else {
            menuName = SetNormalStyle(tmp)
        }
    } else if menuTitleLen+menuSubtitleLen > itemMaxLen {
        tmp = runewidth.Truncate(m.menuList[index].Subtitle, itemMaxLen-menuTitleLen, "")
        tmp = runewidth.FillRight(tmp, itemMaxLen-menuTitleLen)
        if index == m.selectedIndex {
            menuName = fmt.Sprintf("%s%s", SetFgStyle(menuTitle, GetPrimaryColor()), SetFgStyle(tmp, termenv.ANSIBrightBlack))
        } else {
            menuName = fmt.Sprintf("%s%s", SetNormalStyle(menuTitle), SetFgStyle(tmp, termenv.ANSIBrightBlack))
        }
    } else {
        tmp = runewidth.FillRight(m.menuList[index].Subtitle, itemMaxLen-menuTitleLen)
        if index == m.selectedIndex {
            menuName = fmt.Sprintf("%s%s", SetFgStyle(menuTitle, GetPrimaryColor()), SetFgStyle(tmp, termenv.ANSIBrightBlack))
        } else {
            menuName = fmt.Sprintf("%s%s", SetNormalStyle(menuTitle), SetFgStyle(tmp, termenv.ANSIBrightBlack))
        }
    }

    menuItemBuilder.WriteString(menuName)

    return menuItemBuilder.String()
}

// 获取当前页的菜单
func getCurPageMenus(m *NeteaseModel) []MenuItem {
    start := (m.menuCurPage - 1) * m.menuPageSize
    end := int(math.Min(float64(len(m.menuList)), float64(m.menuCurPage*m.menuPageSize)))

    return m.menuList[start:end]
}

// key handle
func keyMsgHandle(msg tea.KeyMsg, m *NeteaseModel) (tea.Model, tea.Cmd) {
    if !m.isListeningKey {
        return m, nil
    }
    switch msg.String() {
    case "j", "J", "down":
        moveDown(m)
    case "k", "K", "up":
        moveUp(m)
    case "h", "H", "left":
        moveLeft(m)
    case "l", "L", "right":
        moveRight(m)
    case "n", "N", "enter":
        enterMenu(m)
    case "b", "B", "esc":
        backMenu(m)
    case " ":
        spaceKeyHandle(m)
    case "[":
        m.player.PreSong()
    case "]":
        m.player.NextSong()
    case "p":
        m.player.SetPlayMode("")
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
        m.quitting = true
        return m, tea.Quit
    case "-":
        m.player.DownVolume()
    case "=":
        m.player.UpVolume()
    case "/":
        // trash playing song
        trashPlayingSong(m)
    case "<":
        // like selected song
        likeSelectedSong(m, true)
    case ">":
        // unlike selected song
        likeSelectedSong(m, false)
    case "?":
        // trash selected song
        trashSelectedSong(m)
    case "r", "R":
        // rerender
        return m, func() tea.Msg {
            return tea.ClearScreenMsg{}
        }
    }

    return m, tickMainUI(time.Nanosecond)
}
