package ui

import "math"

// 上移
func moveUp(m *neteaseModel) {
    if m.doubleColumn {
        if m.selectedIndex - 2 < 0 && m.topOutHook != nil {
            m.topOutHook(m)
        }
        if m.selectedIndex - 2 < 0 {
            return
        }
        m.selectedIndex -= 2
    } else {
        if m.selectedIndex - 1 < 0 && m.topOutHook != nil {
            m.topOutHook(m)
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
            m.bottomOutHook(m)
        }
        if m.selectedIndex + 2 > len(m.menuList) - 1 {
            return
        }
        m.selectedIndex += 2
    } else {
        if m.selectedIndex + 1 > len(m.menuList) - 1 && m.bottomOutHook != nil {
            m.bottomOutHook(m)
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
        m.bottomOutHook(m)
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
        m.beforePrePageHook(m)
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
        originMenuTitle := m.menuTitle
        m.menuTitle = "加载中..."
        m.beforeNextPageHook(m)
        m.menuTitle = originMenuTitle
    }
    if m.menuCurPage >= int(math.Ceil(float64(len(m.menuList)) / float64(m.menuPageSize))) {
        return
    }

    m.menuCurPage++
}