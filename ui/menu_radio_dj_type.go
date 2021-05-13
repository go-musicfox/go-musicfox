package ui

type RadioDjTypeMenu struct {
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

func (m *RadioDjTypeMenu) MenuData() interface{} {
    return nil
}

func (m *RadioDjTypeMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *RadioDjTypeMenu) IsPlayable() bool {
    return false
}

func (m *RadioDjTypeMenu) ResetPlaylistWhenPlay() bool {
    return false
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

func (m *RadioDjTypeMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *RadioDjTypeMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *RadioDjTypeMenu) BeforeEnterMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *RadioDjTypeMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *RadioDjTypeMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}

