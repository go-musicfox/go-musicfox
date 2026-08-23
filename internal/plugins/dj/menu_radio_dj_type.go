package dj

import (
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// RadioDjTypeMenu 是「主播电台」主菜单入口菜单：列出电台各子入口并逐项构建。
// 嵌入 ui.BaseMenu（导出基座），子菜单经 ui.MustBuild* 经注册表构建——key 与
// 提取前一致。
type RadioDjTypeMenu struct {
	ui.BaseMenu
	menus    []model.MenuItem
	menuList []ui.Menu
}

func NewRadioDjTypeMenu(base ui.BaseMenu) *RadioDjTypeMenu {
	menu := &RadioDjTypeMenu{
		BaseMenu: base,
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
		menuList: []ui.Menu{
			ui.MustBuildNoArg("dj_sub", base),
			ui.MustBuildNoArg("dj_recommend", base),
			ui.MustBuildNoArg("dj_today_recommend", base),
			ui.MustBuild("dj_hot", base, DjHotOpts{HotType: DjHot}),
			ui.MustBuild("dj_hot", base, DjHotOpts{HotType: DjNotHot}),
			ui.MustBuildNoArg("dj_category", base),
			ui.MustBuildNoArg("dj_program_rank", base),
			ui.MustBuildNoArg("dj_program_hour_rank", base),
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
