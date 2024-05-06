package ui

import "github.com/anhoder/foxful-cli/model"

type UserCollectMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
}

func NewUserCollectMenu(base baseMenu) *UserCollectMenu {
	menu := &UserCollectMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: "收藏专辑"},
			{Title: "收藏歌手"},
		},
		menuList: []Menu{
			NewAlbumSubListMenu(base),
			NewArtistsSubListMenu(base),
		},
	}

	return menu
}

func (m *UserCollectMenu) GetMenuKey() string {
	return "user_collect"
}

func (m *UserCollectMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *UserCollectMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
