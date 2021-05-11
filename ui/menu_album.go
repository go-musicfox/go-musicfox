package ui

type AlbumMenu struct {
    menus    []MenuItem
    menuList []IMenu
}

func NewAlbumMenu() *AlbumMenu {
    albumMenu := new(AlbumMenu)
    albumMenu.menus = []MenuItem{
        {Title: "新碟上架"},
        {Title: "最新专辑"},
    }
    albumMenu.menuList = []IMenu{
    }

    return albumMenu
}

func (m *AlbumMenu) MenuData() interface{} {
    return nil
}

func (m *AlbumMenu) IsPlayable() bool {
    return false
}

func (m *AlbumMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *AlbumMenu) GetMenuKey() string {
    return "album_menu"
}

func (m *AlbumMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *AlbumMenu) SubMenu(_ *NeteaseModel, index int) IMenu {

    if index >= len(m.menuList) {
        return nil
    }

    return m.menuList[index]
}

func (m *AlbumMenu) ExtraView() string {
    return ""
}

func (m *AlbumMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumMenu) BeforeEnterMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumMenu) BeforeBackMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *AlbumMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}

