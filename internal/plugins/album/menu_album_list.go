package album

import (
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

type AlbumListMenu struct {
	ui.BaseMenu
	menus    []model.MenuItem
	menuList []ui.Menu
}

func NewAlbumListMenu(base ui.BaseMenu) *AlbumListMenu {
	albumMenu := &AlbumListMenu{
		BaseMenu: base,
		menus: []model.MenuItem{
			{Title: "全部新碟"},
			{Title: "新碟上架"},
			{Title: "最新专辑"},
		},
		menuList: []ui.Menu{
			ui.MustBuildNoArg("album_new_area", base),
			ui.MustBuildNoArg("album_top_area", base),
			ui.MustBuildNoArg("album_new_hot", base),
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
