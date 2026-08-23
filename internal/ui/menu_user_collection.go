package ui

import "github.com/anhoder/foxful-cli/model"

type UserCollectionMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
}

func NewUserCollectionMenu(base baseMenu) *UserCollectionMenu {
	menu := &UserCollectionMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: "收藏专辑"},
			{Title: "收藏歌手"},
		},
		menuList: []Menu{
			mustBuildNoArg("album_sub_list", base),
			mustBuildNoArg("artists_sub_list", base),
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
