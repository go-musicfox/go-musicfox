package ui

type AlbumNewAreaMenu struct {
	DefaultMenu
	menus []MenuItem
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
