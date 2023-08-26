package ui

import "github.com/anhoder/foxful-cli/model"

type AlbumNewAreaMenu struct {
	baseMenu
	menus []model.MenuItem
}

func NewAlbumNewAreaMenu(base baseMenu) *AlbumNewAreaMenu {
	areaMenu := &AlbumNewAreaMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: "全部"},
			{Title: "华语"},
			{Title: "欧美"},
			{Title: "韩国"},
			{Title: "日本"},
		},
	}

	return areaMenu
}

func (m *AlbumNewAreaMenu) GetMenuKey() string {
	return "album_new_area"
}

func (m *AlbumNewAreaMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumNewAreaMenu) SubMenu(_ *model.App, index int) model.Menu {
	areaValueMapping := []string{
		"ALL",
		"ZH",
		"EA",
		"KR",
		"JP",
	}

	return NewAlbumNewMenu(m.baseMenu, areaValueMapping[index])
}
