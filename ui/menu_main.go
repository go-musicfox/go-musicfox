package ui

type MainMenu struct {}

func (m *MainMenu) IsPlayable() bool {
    return false
}

func (m *MainMenu) ResetPlaylistWhenEnter() bool {
    return false
}

func (m *MainMenu) GetMenuKey() string {
    return "main_menu"
}

func (m *MainMenu) MenuViews() []string {
    return []string{
        "每日推荐歌曲",
        "每日推荐歌单",
        "我的歌单",
        "私人FM",
        "专辑列表",
        "搜索",
        "排行榜",
        "精选歌单",
        "热门歌手",
        "云盘",
        "主播电台",
        "帮助",
    }
}

func (m *MainMenu) SubMenu(index int) IMenu {
    menuList := []IMenu{
        &DailyRecommendSongsMenu{},
    }

    if index >= len(menuList) {
        return nil
    }

    return menuList[index]
}

func (m *MainMenu) ExtraView() string {
    return ""
}

func (m *MainMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *MainMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *MainMenu) BeforeEnterMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *MainMenu) BeforeBackMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *MainMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *MainMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}

