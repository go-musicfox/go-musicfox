package ui

import "github.com/anhoder/foxful-cli/model"

type RadioDjTypeMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
}

func NewRadioDjTypeMenu(base baseMenu) *RadioDjTypeMenu {
	menu := &RadioDjTypeMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: "我的订阅"},
			{Title: "推荐电台"},
			{Title: "今日优选"},
			{Title: "热门电台"},
			{Title: "新晋电台"},
			{Title: "电台分类"},
			{Title: "节目榜单"},
			{Title: "24小时节目榜"},
		},
		menuList: []Menu{
			NewDjSubListMenu(base),
			NewDjRecommendMenu(base),
			NewDjTodayRecommendMenu(base),
			NewDjHotMenu(base, DjHot),
			NewDjHotMenu(base, DjNotHot),
			NewDjCategoryMenu(base),
			NewDjProgramRankMenu(base),
			NewDjProgramHoursRankMenu(base),
		},
	}

	return menu
}

func (m *RadioDjTypeMenu) GetMenuKey() string {
	return "radio_dj_type"
}

func (m *RadioDjTypeMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *RadioDjTypeMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
