package ui

type AlbumTopAreaMenu struct {
    menus    []MenuItem
}

func NewAlbumTopAreaMenu() *AlbumTopAreaMenu {
    areaMenu := new(AlbumTopAreaMenu)
    areaMenu.menus = []MenuItem{
        {Title: "全部"},
        {Title: "华语"},
        {Title: "欧美"},
        {Title: "韩国"},
        {Title: "日本"},
    }

    return areaMenu
}

func (m *AlbumTopAreaMenu) MenuData() interface{} {
    return nil
}

func (m *AlbumTopAreaMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *AlbumTopAreaMenu) IsPlayable() bool {
    return false
}

func (m *AlbumTopAreaMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *AlbumTopAreaMenu) GetMenuKey() string {
    return "album_top_area"
}

func (m *AlbumTopAreaMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *AlbumTopAreaMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    areaValueMapping := []string{
        "ALL",
        "ZH",
        "EA",
        "KR",
        "JP",
    }

    return NewAlbumTopMenu(areaValueMapping[index])
}

func (m *AlbumTopAreaMenu) ExtraView() string {
    return ""
}

func (m *AlbumTopAreaMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumTopAreaMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumTopAreaMenu) BeforeEnterMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumTopAreaMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumTopAreaMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}

