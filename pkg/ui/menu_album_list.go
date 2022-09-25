package ui

type AlbumListMenu struct {
	DefaultMenu
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
