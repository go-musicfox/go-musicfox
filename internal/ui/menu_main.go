package ui

import (
	"fmt"
	"sort"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
)

// mainMenuBuiltinOrder is the order of each built-in main-menu entry in the
// merged menu; they interleave with plugin entries to reproduce the original
// pre-extraction order (see NewMainMenu). The help entry carries order 14 and
// no plugin item may declare that order, so in the full production set it
// always lands at index 14.
const (
	mainMenuSearchOrder = 6
	mainMenuHelpOrder   = 14
)

type MainMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
	// helpIndex is the merged-menu index of the built-in 帮助 entry (its
	// submenu placeholder is nil and Action opens the Markdown popup directly).
	// Computed at construction so it stays correct for any subset of the
	// plugin set (e.g. test binaries); in the full production set it is 14.
	helpIndex int
}

// mainMenuEntry is one row of the merged built-in + plugin main menu before
// the order-based sort. builtin marks built-in entries so ties resolve to
// built-ins first.
type mainMenuEntry struct {
	order   int
	builtin bool
	title   string
	menu    Menu
}

func NewMainMenu(base baseMenu) *MainMenu {
	mainMenu := &MainMenu{baseMenu: base}

	// 内置项与插件项按 Order 归并排序，复现插件化前的主菜单原始顺序：
	// 每日推荐歌曲0 / 每日推荐歌单1 / 我的歌单2 / 我的收藏3 / 私人FM4 /
	// 专辑列表5 / 搜索6 / 排行榜7 / 精选歌单8 / 热门歌手9 / 最近播放歌曲10 /
	// 云盘11 / 主播电台12 / LastFM13 / 帮助14 / 检查更新15。经普通注册形式
	// （RegisterMainMenuItem / RegisterMainMenuItemWith）声明的项携带
	// mainMenuUnsetOrder，排在所有显式顺序项之后（保持既有"追加在末尾"
	// 行为）。无 Build 的插件主菜单项 MUST 是无参菜单：经 mustBuildNoArg
	// 构建，key 未注册或为参数化菜单会在启动时 panic（程序错误，作为启动
	// 完整性信号；先显式断言以给出清晰错误）。带 Build 的项由插件以自身
	// options 构造菜单（参数化 provider 入口）。触发由插件菜单自身的
	// Action / BeforeEnterMenuHook 承担——主菜单不再对插件索引做特判。
	entries := []mainMenuEntry{
		{order: mainMenuSearchOrder, builtin: true, title: "搜索", menu: mustBuildNoArg("search_type", base)},
		{order: mainMenuHelpOrder, builtin: true, title: "帮助", menu: nil}, // 帮助由 Action 直接打开 Markdown 弹窗，不再进入子菜单。
	}
	for _, item := range MainMenuPluginItems() {
		if _, ok := menuRegistry[item.Key]; !ok {
			panic(fmt.Sprintf("main menu plugin item %q: menu provider not registered", item.Key))
		}
		entry := mainMenuEntry{order: item.Order, title: item.Title}
		if item.Build != nil {
			entry.menu = item.Build(base)
		} else {
			entry.menu = mustBuildNoArg(item.Key, base)
		}
		entries = append(entries, entry)
	}
	// 稳定排序：同 order 时内置项在前，插件项保持注册序。
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].order != entries[j].order {
			return entries[i].order < entries[j].order
		}
		return entries[i].builtin && !entries[j].builtin
	})
	for _, e := range entries {
		mainMenu.menus = append(mainMenu.menus, model.MenuItem{Title: e.title})
		mainMenu.menuList = append(mainMenu.menuList, e.menu)
		if e.builtin && e.menu == nil {
			mainMenu.helpIndex = len(mainMenu.menus) - 1
		}
	}
	return mainMenu
}

func (m *MainMenu) FormatMenuItem(item *model.MenuItem) {
	subtitle := "[未登录]"
	if m.svc.User() != nil {
		subtitle = "[" + m.svc.User().Nickname + "]"
	}
	item.Subtitle = subtitle
}

func (m *MainMenu) GetMenuKey() string {
	return "main_menu"
}

// Titles returns the main-menu item titles in display order. Side-effect free:
// subtitles are refreshed by MenuViews / FormatMenuItem when the menu renders,
// so reading the raw titles (e.g. asserting the merged built-in + plugin item
// order in tests) needs no live services.
func (m *MainMenu) Titles() []string {
	titles := make([]string, len(m.menus))
	for i, item := range m.menus {
		titles[i] = item.Title
	}
	return titles
}

func (m *MainMenu) MenuViews() []model.MenuItem {
	for i, menu := range m.menuList {
		if menu != nil {
			menu.FormatMenuItem(&m.menus[i])
		}
	}
	return m.menus
}

func (m *MainMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index < 0 || index >= len(m.menuList) {
		return nil
	}
	return m.menuList[index]
}

func (m *MainMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	switch index {
	case m.helpIndex:
		showHelpPopup(app)
		return app.MustMain(), nil
	default:
		// 插件主菜单项（如检查更新）不再特判：Action 落到默认分支（非播放
		// 菜单返回 nil/nil），由 SubMenu 进入插件菜单，触发由插件菜单自身的
		// Action / BeforeEnterMenuHook 承担。
		return m.baseMenu.Action(app, index)
	}
}
