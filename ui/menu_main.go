package ui

type MainMenu struct {}

func (m *MainMenu) IsPlayable() bool {
    return false
}

func (m *MainMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *MainMenu) GetMenuKey() string {
    return "main_menu"
}

func (m *MainMenu) MenuViews() []MenuItem {
    return []MenuItem{
        {Title: "每日推荐歌曲"},
        {Title: "每日推荐歌单"},
        {Title: "我的歌单"},
        {Title: "私人FM"},
        {Title: "专辑列表"},
        {Title: "搜索"},
        {Title: "排行榜"},
        {Title: "精选歌单"},
        {Title: "热门歌手"},
        {Title: "云盘"},
        {Title: "主播电台"},
        {Title: "帮助"},
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

