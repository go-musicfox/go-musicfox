package playlist

import (
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

type UserCollectionMenu struct {
	ui.BaseMenu
	menus    []model.MenuItem
	menuList []ui.Menu
}

func NewUserCollectionMenu(base ui.BaseMenu) *UserCollectionMenu {
	menu := &UserCollectionMenu{
		BaseMenu: base,
		menus: []model.MenuItem{
			{Title: "收藏专辑"},
			{Title: "收藏歌手"},
		},
		menuList: []ui.Menu{
			ui.MustBuildNoArg("album_sub_list", base),
			ui.MustBuildNoArg("artists_sub_list", base),
		},
	}

	return menu
}

func (m *UserCollectionMenu) GetMenuKey() string {
	return "user_collect"
}

func (m *UserCollectionMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *UserCollectionMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
