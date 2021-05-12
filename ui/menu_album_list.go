package ui

type AlbumListMenu struct {
    menus    []MenuItem
    menuList []IMenu
}

func NewAlbumListMenu() *AlbumListMenu {
    albumMenu := new(AlbumListMenu)
    albumMenu.menus = []MenuItem{
        {Title: "全部新碟"},
        {Title: "新碟上架"},
        {Title: "最新专辑"},
    }
    albumMenu.menuList = []IMenu{
        NewAlbumNewAreaMenu(),
        NewAlbumTopAreaMenu(),
        NewAlbumNewestMenu(),
    }

    return albumMenu
}

func (m *AlbumListMenu) MenuData() interface{} {
    return nil
}

func (m *AlbumListMenu) IsPlayable() bool {
    return false
}

func (m *AlbumListMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *AlbumListMenu) GetMenuKey() string {
    return "album_menu"
}

func (m *AlbumListMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *AlbumListMenu) SubMenu(_ *NeteaseModel, index int) IMenu {

    if index >= len(m.menuList) {
        return nil
    }

    return m.menuList[index]
}

func (m *AlbumListMenu) ExtraView() string {
    return ""
}

func (m *AlbumListMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumListMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumListMenu) BeforeEnterMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumListMenu) BeforeBackMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumListMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumListMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}

