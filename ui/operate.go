package ui

import (
    "math"
)

type menuStackItem struct {
    menuList      []string
    selectedIndex int
    menuTitle     string
    menuData      interface{}
}

// 上移
func moveUp(m *NeteaseModel) {
    if m.doubleColumn {
        if m.selectedIndex - 2 < 0 {
            loading := NewLoading(m)
            loading.start()
            m.menu.TopOutHook(m)
            loading.complete()
        }
        if m.selectedIndex - 2 < 0 {
            return
        }
        m.selectedIndex -= 2
    } else {
        if m.selectedIndex - 1 < 0 {
            loading := NewLoading(m)
            loading.start()
            m.menu.TopOutHook(m)
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
    if m.doubleColumn {
        if m.selectedIndex + 2 > len(m.menuList) - 1 {
            loading := NewLoading(m)
            loading.start()
            m.menu.TopOutHook(m)
            loading.complete()
        }
        if m.selectedIndex + 2 > len(m.menuList) - 1 {
            return
        }
        m.selectedIndex += 2
    } else {
        if m.selectedIndex + 1 > len(m.menuList) - 1 {
            loading := NewLoading(m)
            loading.start()
            m.menu.TopOutHook(m)
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
    if m.selectedIndex + 1 > len(m.menuList) - 1 {
        loading := NewLoading(m)
        loading.start()
        m.menu.BottomOutHook(m)
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

    loading := NewLoading(m)
    loading.start()
    m.menu.BeforePrePageHook(m)
    loading.complete()

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

    loading := NewLoading(m)
    loading.start()
    m.menu.BeforeNextPageHook(m)
    loading.complete()
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

    loading := NewLoading(m)
    loading.start()
    newMenu := m.menu.BeforeEnterMenuHook(m)
    loading.complete()

    if len(newMenu) <= 0 {
        m.menuStack.Pop()
        return
    }
    newTitle := m.menuList[m.selectedIndex]
    stackItem := &menuStackItem{
        menuList: m.menuList,
        selectedIndex: m.selectedIndex,
        menuTitle: m.menuTitle,
        menuData: m.menuData,
    }
    m.menuStack.Push(stackItem)

    m.menuTitle = newTitle
    m.selectedIndex = 0
    m.menuCurPage = 1
}