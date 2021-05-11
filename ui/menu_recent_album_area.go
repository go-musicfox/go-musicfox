package ui

type RecentAlbumAreaMenu struct {
    menus    []MenuItem
}

func NewRecentAlbumAreaMenu() *RecentAlbumAreaMenu {
    areaMenu := new(RecentAlbumAreaMenu)
    areaMenu.menus = []MenuItem{
        {Title: "全部"},
        {Title: "华语"},
        {Title: "欧美"},
        {Title: "韩国"},
        {Title: "日本"},
    }

    return areaMenu
}

func (m *RecentAlbumAreaMenu) MenuData() interface{} {
    return nil
}

func (m *RecentAlbumAreaMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *RecentAlbumAreaMenu) IsPlayable() bool {
    return false
}

func (m *RecentAlbumAreaMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *RecentAlbumAreaMenu) GetMenuKey() string {
    return "recent_album_area"
}

func (m *RecentAlbumAreaMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *RecentAlbumAreaMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    areaValueMapping := []string{
        "ALL",
        "ZH",
        "EA",
        "KR",
        "JP",
    }

    return NewRecentAlbumMenu(areaValueMapping[index])
}

func (m *RecentAlbumAreaMenu) ExtraView() string {
    return ""
}

func (m *RecentAlbumAreaMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *RecentAlbumAreaMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *RecentAlbumAreaMenu) BeforeEnterMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *RecentAlbumAreaMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *RecentAlbumAreaMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}

