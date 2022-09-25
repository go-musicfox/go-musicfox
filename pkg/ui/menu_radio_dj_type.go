package ui

type RadioDjTypeMenu struct {
	DefaultMenu
	menus    []MenuItem
	menuList []IMenu
}

func NewRadioDjTypeMenu() *RadioDjTypeMenu {
	menu := new(RadioDjTypeMenu)
	menu.menus = []MenuItem{
		{Title: "我的订阅"},
		{Title: "推荐电台"},
		{Title: "今日优选"},
		{Title: "热门电台"},
		{Title: "新晋电台"},
		{Title: "电台分类"},
		{Title: "节目榜单"},
		{Title: "24小时节目榜"},
	}
	menu.menuList = []IMenu{
		NewDjSubListMenu(),
		NewDjRecommendMenu(),
		NewDjTodayRecommendMenu(),
		NewDjHotMenu(DjHot),
		NewDjHotMenu(DjNotHot),
		NewDjCategoryMenu(),
		NewDjProgramRankMenu(),
		NewDjProgramHoursRankMenu(),
	}

	return menu
}

func (m *RadioDjTypeMenu) GetMenuKey() string {
	return "radio_dj_type"
}

func (m *RadioDjTypeMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *RadioDjTypeMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
