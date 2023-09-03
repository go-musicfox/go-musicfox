package ui

import "github.com/anhoder/foxful-cli/model"

type AlbumTopAreaMenu struct {
	baseMenu
	menus []model.MenuItem
}

func NewAlbumTopAreaMenu(base baseMenu) *AlbumTopAreaMenu {
	areaMenu := &AlbumTopAreaMenu{
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

func (m *AlbumTopAreaMenu) GetMenuKey() string {
	return "album_top_area"
}

func (m *AlbumTopAreaMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumTopAreaMenu) SubMenu(_ *model.App, index int) model.Menu {
	areaValueMapping := []string{
		"ALL",
		"ZH",
		"EA",
		"KR",
		"JP",
	}

	return NewAlbumTopMenu(m.baseMenu, areaValueMapping[index])
}
