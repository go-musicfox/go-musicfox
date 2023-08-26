package ui

import "github.com/anhoder/foxful-cli/model"

type AlbumListMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
}

func NewAlbumListMenu(base baseMenu) *AlbumListMenu {
	albumMenu := &AlbumListMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: "全部新碟"},
			{Title: "新碟上架"},
			{Title: "最新专辑"},
		},
		menuList: []Menu{
			NewAlbumNewAreaMenu(base),
			NewAlbumTopAreaMenu(base),
			NewAlbumNewestMenu(base),
		},
	}

	return albumMenu
}

func (m *AlbumListMenu) GetMenuKey() string {
	return "album_menu"
}

func (m *AlbumListMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumListMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
