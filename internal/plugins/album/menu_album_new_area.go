package album

import (
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

type AlbumNewAreaMenu struct {
	ui.BaseMenu
	menus []model.MenuItem
}

func NewAlbumNewAreaMenu(base ui.BaseMenu) *AlbumNewAreaMenu {
	areaMenu := &AlbumNewAreaMenu{
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

	return ui.BuildMenuOrToast("album_new", m.BaseMenu, AlbumNewOpts{Area: areaValueMapping[index]})
}
