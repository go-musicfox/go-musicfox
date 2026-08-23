package album

import (
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

type AlbumTopAreaMenu struct {
	ui.BaseMenu
	menus []model.MenuItem
}

func NewAlbumTopAreaMenu(base ui.BaseMenu) *AlbumTopAreaMenu {
	areaMenu := &AlbumTopAreaMenu{
		BaseMenu: base,
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

	return ui.BuildMenuOrToast("album_top", m.BaseMenu, AlbumTopOpts{Area: areaValueMapping[index]})
}
