package ui

import (
    "fmt"
    tea "github.com/anhoder/bubbletea"
    "github.com/anhoder/go-musicfox/constants"
    "github.com/muesli/termenv"
    "math"
    "strings"
    "time"
    "unicode/utf8"
)

type mainMenuModel struct {
    doubleColumn    bool // 是否双列显示

    menuTitle     	     string // 菜单标题
    menuTitleStartRow    int    // 菜单标题开始行
    menuTitleStartColumn int    // 菜单标题开始列

    menuStartRow    int // 菜单开始行
    menuStartColumn int // 菜单开始列
    menuBottomRow   int // 菜单最底部所在行

    menuCurPage	 int // 菜单当前页
    menuPageSize int // 菜单每页大小

    menuList      []string // 菜单列表
    selectedIndex int	   // 当前选中的菜单index

    beforePrePageHook   func(*neteaseModel) // 切换上一页前的Hook
    beforeNextPageHook  func(*neteaseModel) // 切换下一页前的Hook
    beforeEnterMenuHook func(*neteaseModel) // 进入菜单项前的Hook
    bottomOutHook       func(*neteaseModel) // 触底Hook
    topOutHook          func(*neteaseModel) // 触顶Hook

    player *Player // 播放器
}

// update main ui
func updateMainUI(msg tea.Msg, m *neteaseModel) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        return keyMsgHandle(msg, m)

    case tickMainUIMsg:
        return m, nil

    case tea.WindowSizeMsg:
        m.doubleColumn = msg.Width>=80

        // 菜单开始行、列
        m.menuStartRow = msg.Height/3
        if m.doubleColumn {
            m.menuStartColumn = (msg.Width-60)/2
            m.menuBottomRow = m.menuStartRow+int(math.Ceil(float64(m.menuPageSize)/2))-1
        } else {
            m.menuStartColumn = (msg.Width-20)/2
            m.menuBottomRow = m.menuStartRow+m.menuPageSize-1
        }

        // 菜单标题开始行、列
        m.menuTitleStartColumn = m.menuStartColumn
        if constants.MainShowTitle && m.menuStartRow > 2 {
            if m.menuStartRow > 4 {
                m.menuTitleStartRow = m.menuStartRow - 3
            }
        } else if !constants.MainShowTitle && m.menuStartRow > 1 {
            if m.menuStartRow > 3 {
                m.menuTitleStartRow = m.menuStartRow - 3
            }
        }

        // 播放器歌词
        spaceHeight := m.WindowHeight - 4 - m.menuBottomRow
        if spaceHeight < 3 {
            // 不显示歌词
            m.player.showLyric = false
        } else {
            m.player.showLyric = true

            if spaceHeight > 5 {
                // 5行歌词
                m.player.lyricStartRow = (m.WindowHeight - 3 + m.menuBottomRow) / 2 - 2
                m.player.lyricLines = 5
            } else {
                // 3行歌词
                m.player.lyricStartRow = (m.WindowHeight - 3 + m.menuBottomRow) / 2 - 1
                m.player.lyricLines = 3
            }

        }

    }

    return m, nil
}

// get main ui view
func mainUIView(m *neteaseModel) string {
    if m.WindowWidth <= 0 || m.WindowHeight <= 0 {
        return ""
    }

    var builder strings.Builder

    // 距离顶部的行数
    top := 0

    // title
    if constants.MainShowTitle {
        builder.WriteString(titleView(m, &top))
    }

    // menu title
    builder.WriteString(menuTitleView(m, &top))

    // menu list
    builder.WriteString(menuListView(m, &top))


    return builder.String()
}

// title view
func titleView(m *neteaseModel, top *int) string {
    var titleBuilder strings.Builder
    titleLen := utf8.RuneCountInString(constants.AppName)+2
    prefixLen := (m.WindowWidth-titleLen)/2
    suffixLen := m.WindowWidth-prefixLen-titleLen
    titleBuilder.WriteString(strings.Repeat("─", prefixLen))
    titleBuilder.WriteString(" ")
    titleBuilder.WriteString(strings.ToUpper(constants.AppName))
    titleBuilder.WriteString(" ")
    titleBuilder.WriteString(strings.Repeat("─", suffixLen))

    *top++

    return SetFgStyle(titleBuilder.String(), primaryColor)
}

// menu title
func menuTitleView(m *neteaseModel, top *int) string {
    var menuTitleBuilder strings.Builder
    var title = m.menuTitle
    if len(m.menuTitle) > 50 {
        menuTitleRunes := []rune(m.menuTitle)
        title = string(menuTitleRunes[:50])
    }

    menuTitleBuilder.WriteString(strings.Repeat("\n", m.menuTitleStartRow - *top))
    menuTitleBuilder.WriteString(strings.Repeat(" ", m.menuTitleStartColumn))
    menuTitleBuilder.WriteString(SetFgStyle(title, termenv.ANSIGreen))

    *top = m.menuTitleStartRow

    return menuTitleBuilder.String()
}

// 菜单列表
func menuListView(m *neteaseModel, top *int) string {
    var menuListBuilder strings.Builder
    menus := getCurPageMenus(m)
    var lines, maxLines int
    if m.doubleColumn {
        lines = int(math.Ceil(float64(len(menus))/2))
        maxLines = int(math.Ceil(float64(m.menuPageSize)/2))
    } else {
        lines = len(menus)
        maxLines = m.menuPageSize
    }

    menuListBuilder.WriteString(strings.Repeat("\n", m.menuStartRow-*top))

    var str string
    for i := 0; i < lines; i++ {
        str = menuLineView(m, i)
        menuListBuilder.WriteString(str)
        menuListBuilder.WriteString("\n")
    }
    if maxLines > lines {
        menuListBuilder.WriteString(strings.Repeat(" ", m.WindowWidth - m.menuStartColumn))
        menuListBuilder.WriteString("\n")
    }

    return menuListBuilder.String()
}

// 菜单Line
func menuLineView(m *neteaseModel, line int) string {
    var menuLineBuilder strings.Builder
    var index int
    if m.doubleColumn {
        index = line * 2 + (m.menuCurPage - 1) * m.menuPageSize
    } else {
        index = line + (m.menuCurPage - 1) * m.menuPageSize
    }
    if index > len(m.menuList) - 1 {
        index = len(m.menuList) - 1
    }
    menuLineBuilder.WriteString(strings.Repeat(" ", m.menuStartColumn - 4))
    menuLineBuilder.WriteString(menuItemView(m, index))
    if m.doubleColumn && index < len(m.menuList) - 1 {
        menuLineBuilder.WriteString(strings.Repeat(" ", 20))
        menuLineBuilder.WriteString(menuItemView(m, index + 1))
    }

    return menuLineBuilder.String()
}

// 菜单Item
func menuItemView(m *neteaseModel, index int) string {
    var menuItemBuilder strings.Builder

    menuName := []rune(m.menuList[index])
    if len(menuName) > 20 {
        menuName = menuName[:20]
    }

    if index == m.selectedIndex {
        menuItemBuilder.WriteString(SetFgStyle(fmt.Sprintf(" => %d. %s", index, string(menuName)), primaryColor))
    } else {
        menuItemBuilder.WriteString(SetFgStyle(fmt.Sprintf("    %d. %s", index, string(menuName)), termProfile.Color(constants.MainMenuUnselectedColor)))
    }

    return menuItemBuilder.String()
}

// 获取当前页的菜单
func getCurPageMenus(m *neteaseModel) []string {
    start := (m.menuCurPage - 1) * m.menuPageSize
    end := int(math.Min(float64(len(m.menuList)), float64(m.menuCurPage*m.menuPageSize)))

    return m.menuList[start:end]
}

// key handle
func keyMsgHandle(msg tea.KeyMsg, m *neteaseModel) (tea.Model, tea.Cmd) {
    if !m.isListeningKey {
        return m, nil
    }
    switch msg.String() {
    case "j", "down":
       moveDown(m)
    case "k", "up":
       moveUp(m)
    case "h", "left":
        moveLeft(m)
    case "l", "right":
        moveRight(m)
    case "enter":
    }

    return m, tickMainUI(time.Nanosecond)
}