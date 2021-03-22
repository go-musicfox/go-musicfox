package ui

import (
    "math"
)

// 上移
func moveUp(m *neteaseModel) {
    if m.doubleColumn {
        if m.selectedIndex - 2 < 0 && m.topOutHook != nil {
            loading := NewLoading(m)
            loading.start()
            m.topOutHook(m)
            loading.complete()
        }
        if m.selectedIndex - 2 < 0 {
            return
        }
        m.selectedIndex -= 2
    } else {
        if m.selectedIndex - 1 < 0 && m.topOutHook != nil {
            loading := NewLoading(m)
            loading.start()
            m.topOutHook(m)
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
func moveDown(m *neteaseModel) {
    if m.doubleColumn {
        if m.selectedIndex + 2 > len(m.menuList) - 1 && m.bottomOutHook != nil {
            loading := NewLoading(m)
            loading.start()
            m.bottomOutHook(m)
            loading.complete()
        }
        if m.selectedIndex + 2 > len(m.menuList) - 1 {
            return
        }
        m.selectedIndex += 2
    } else {
        if m.selectedIndex + 1 > len(m.menuList) - 1 && m.bottomOutHook != nil {
            loading := NewLoading(m)
            loading.start()
            m.bottomOutHook(m)
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
func moveLeft(m *neteaseModel) {
    if !m.doubleColumn || m.selectedIndex % 2 == 0 || m.selectedIndex - 1 < 0 {
        return
    }
    m.selectedIndex--
}

// 右移
func moveRight(m *neteaseModel) {
    if !m.doubleColumn || m.selectedIndex % 2 != 0 {
        return
    }
    if m.selectedIndex + 1 > len(m.menuList) - 1 && m.bottomOutHook != nil {
        loading := NewLoading(m)
        loading.start()
        m.bottomOutHook(m)
        loading.complete()
    }
    if m.selectedIndex + 1 > len(m.menuList) - 1 {
        return
    }
    m.selectedIndex++
}

// 切换到上一页
func prePage(m *neteaseModel) {
    m.isListeningKey = false
    defer func() {
        m.isListeningKey = true
    }()

    if m.beforePrePageHook != nil {
        loading := NewLoading(m)
        loading.start()
        m.beforePrePageHook(m)
        loading.complete()
    }
    if m.menuCurPage <= 1 {
        return
    }
    m.menuCurPage--
}

// 切换到下一页
func nextPage(m *neteaseModel) {
    m.isListeningKey = false
    defer func() {
        m.isListeningKey = true
    }()

    if m.beforeNextPageHook != nil {
        loading := NewLoading(m)
        loading.start()
        m.beforeNextPageHook(m)
        loading.complete()
    }
    if m.menuCurPage >= int(math.Ceil(float64(len(m.menuList)) / float64(m.menuPageSize))) {
        return
    }

    m.menuCurPage++
}