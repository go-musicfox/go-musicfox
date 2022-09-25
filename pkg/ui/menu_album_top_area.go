package ui

type AlbumTopAreaMenu struct {
	DefaultMenu
	menus []MenuItem
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
