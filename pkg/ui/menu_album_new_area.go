package ui

type AlbumNewAreaMenu struct {
    menus    []MenuItem
}

func NewAlbumNewAreaMenu() *AlbumNewAreaMenu {
    areaMenu := new(AlbumNewAreaMenu)
    areaMenu.menus = []MenuItem{
        {Title: "全部"},
        {Title: "华语"},
        {Title: "欧美"},
        {Title: "韩国"},
        {Title: "日本"},
    }

    return areaMenu
}

func (m *AlbumNewAreaMenu) MenuData() interface{} {
    return nil
}

func (m *AlbumNewAreaMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *AlbumNewAreaMenu) IsPlayable() bool {
    return false
}

func (m *AlbumNewAreaMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *AlbumNewAreaMenu) GetMenuKey() string {
    return "album_new_area"
}

func (m *AlbumNewAreaMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *AlbumNewAreaMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    areaValueMapping := []string{
        "ALL",
        "ZH",
        "EA",
        "KR",
        "JP",
    }

    return NewAlbumNewMenu(areaValueMapping[index])
}

func (m *AlbumNewAreaMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumNewAreaMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumNewAreaMenu) BeforeEnterMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumNewAreaMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumNewAreaMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}

