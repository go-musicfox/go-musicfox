package ui

import (
    "math"
)

type menuStackItem struct {
    menuList      []string
    selectedIndex int
    menuTitle     string
    menuData      interface{}
    menu          IMenu
}

// 上移
func moveUp(m *NeteaseModel) {
    topHook := m.menu.TopOutHook()
    if m.doubleColumn {
        if m.selectedIndex - 2 < 0 && topHook != nil {
            loading := NewLoading(m)
            loading.start()
            if res := topHook(m); !res {
                loading.complete()
                return
            }
            loading.complete()
        }
        if m.selectedIndex - 2 < 0 {
            return
        }
        m.selectedIndex -= 2
    } else {
        if m.selectedIndex - 1 < 0 && topHook != nil {
            loading := NewLoading(m)
            loading.start()
            if res := topHook(m); !res {
                loading.complete()
                return
            }
            loading.complete()
        }
        if m.selectedIndex - 1 < 0 {
            return
        }
        m.selectedIndex--
    }
    if m.selectedIndex < (m.menuCurPage - 1) * m.menuPageSize {
        prePage(m)
    }
}

// 下移
func moveDown(m *NeteaseModel) {
    bottomHook := m.menu.BottomOutHook()
    if m.doubleColumn {
        if m.selectedIndex + 2 > len(m.menuList) - 1 && bottomHook != nil {
            loading := NewLoading(m)
            loading.start()
            if res := bottomHook(m); !res {
                loading.complete()
                return
            }
            loading.complete()
        }
        if m.selectedIndex + 2 > len(m.menuList) - 1 {
            return
        }
        m.selectedIndex += 2
    } else {
        if m.selectedIndex + 1 > len(m.menuList) - 1 && bottomHook != nil {
            loading := NewLoading(m)
            loading.start()
            if res := bottomHook(m); !res {
                loading.complete()
                return
            }
            loading.complete()
        }
        if m.selectedIndex + 1 > len(m.menuList) - 1 {
            return
        }
        m.selectedIndex++
    }
    if m.selectedIndex >= m.menuCurPage * m.menuPageSize {
        nextPage(m)
    }
}

// 左移
func moveLeft(m *NeteaseModel) {
    if !m.doubleColumn || m.selectedIndex % 2 == 0 || m.selectedIndex - 1 < 0 {
        return
    }
    m.selectedIndex--
}

// 右移
func moveRight(m *NeteaseModel) {
    if !m.doubleColumn || m.selectedIndex % 2 != 0 {
        return
    }
    if bottomHook := m.menu.BottomOutHook(); m.selectedIndex + 1 > len(m.menuList) - 1 && bottomHook != nil {
        loading := NewLoading(m)
        loading.start()
        if res := bottomHook(m); !res {
            loading.complete()
            return
        }
        loading.complete()
    }
    if m.selectedIndex + 1 > len(m.menuList) - 1 {
        return
    }
    m.selectedIndex++
}

// 切换到上一页
func prePage(m *NeteaseModel) {
    m.isListeningKey = false
    defer func() {
        m.isListeningKey = true
    }()

    if prePageHook := m.menu.BeforePrePageHook(); prePageHook != nil {
        loading := NewLoading(m)
        loading.start()
        if res := prePageHook(m); !res {
            loading.complete()
            return
        }
        loading.complete()
    }

    if m.menuCurPage <= 1 {
        return
    }
    m.menuCurPage--
}

// 切换到下一页
func nextPage(m *NeteaseModel) {
    m.isListeningKey = false
    defer func() {
        m.isListeningKey = true
    }()

    if nextPageHook := m.menu.BeforeNextPageHook(); nextPageHook != nil {
        loading := NewLoading(m)
        loading.start()
        if res := nextPageHook(m); !res {
            loading.complete()
            return
        }
        loading.complete()
    }
    if m.menuCurPage >= int(math.Ceil(float64(len(m.menuList)) / float64(m.menuPageSize))) {
        return
    }

    m.menuCurPage++
}

// 进入菜单
func enterMain(m *NeteaseModel) {
    m.isListeningKey = false
    defer func() {
        m.isListeningKey = true
    }()

    if m.selectedIndex >= len(m.menuList) {
        return
    }

    newTitle := m.menuList[m.selectedIndex]
    stackItem := &menuStackItem{
        menuList: m.menuList,
        selectedIndex: m.selectedIndex,
        menuTitle: m.menuTitle,
        menuData: m.menuData,
        menu: m.menu,
    }
    m.menuStack.Push(stackItem)

    menu := m.menu.SubMenu(m.selectedIndex)
    if menu == nil {
        m.menuStack.Pop()
        return
    }

    if enterMenuHook := menu.BeforeEnterMenuHook(); enterMenuHook != nil {
        loading := NewLoading(m)
        loading.start()
        if res := enterMenuHook(m); !res {
            loading.complete()
            return
        }
        loading.complete()
    }

    menuList := menu.MenuViews()
    if len(menuList) <= 0 {
        m.menuStack.Pop()
        return
    }

    m.menu = menu
    m.menuList = menuList
    m.menuTitle = newTitle
    m.selectedIndex = 0
    m.menuCurPage = 1
}

// 菜单返回
func backMenu(m *NeteaseModel) {
    m.isListeningKey = false
    defer func() {
        m.isListeningKey = true
    }()

    if m.menuStack.Len() <= 0 {
        return
    }

    stackItem := m.menuStack.Pop()
    if backMenuHook := m.menu.BeforeBackMenuHook(); backMenuHook != nil {
        loading := NewLoading(m)
        loading.start()
        if res := backMenuHook(m); !res {
            loading.complete()
            return
        }
        loading.complete()
    }

    stackMenu, ok := stackItem.(*menuStackItem)
    if !ok {
        return
    }

    m.menuList = stackMenu.menuList
    m.menu = stackMenu.menu
    m.menuTitle = stackMenu.menuTitle
    m.menuData = stackMenu.menuData
    m.selectedIndex = stackMenu.selectedIndex
}