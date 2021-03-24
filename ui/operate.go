package ui

import (
    "math"
)

// 上移
func moveUp(m *NeteaseModel) {
    topHook := m.menu.TopOutHook()
    if m.doubleColumn {
        if m.selectedIndex - 2 < 0 && topHook != nil {
            loading := NewLoading(m)
            loading.start()
            topHook(m)
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
            topHook(m)
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
            bottomHook(m)
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
            bottomHook(m)
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
    bottomHook := m.menu.BottomOutHook()
    if m.selectedIndex + 1 > len(m.menuList) - 1 && bottomHook != nil {
        loading := NewLoading(m)
        loading.start()
        bottomHook(m)
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


    if beforePrePageHook := m.menu.BeforePrePageHook(); beforePrePageHook != nil {
        loading := NewLoading(m)
        loading.start()
        beforePrePageHook(m)
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

    if beforeNextPageHook := m.menu.BeforeNextPageHook(); beforeNextPageHook != nil {
        loading := NewLoading(m)
        loading.start()
        beforeNextPageHook(m)
        loading.complete()
    }
    if m.menuCurPage >= int(math.Ceil(float64(len(m.menuList)) / float64(m.menuPageSize))) {
        return
    }

    m.menuCurPage++
}