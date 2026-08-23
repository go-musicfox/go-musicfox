package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
)

// mainMenuHelpIndex is the hardcoded index of the built-in 帮助 entry (the
// last built-in item; plugin main-menu items are appended after it and thus
// never shift this index). Its submenu placeholder is nil and Action handles
// the Markdown popup directly.
const mainMenuHelpIndex = 10

type MainMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
}

func NewMainMenu(base baseMenu) *MainMenu {
	mainMenu := &MainMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: "每日推荐歌曲"},
			{Title: "每日推荐歌单"},
			{Title: "我的歌单"},
			{Title: "我的收藏"},
			{Title: "私人FM"},
			{Title: "搜索"},
			{Title: "排行榜"},
			{Title: "精选歌单"},
			{Title: "最近播放歌曲"},
			{Title: "云盘"},
			{Title: "帮助"},
		},
		menuList: []Menu{
			mustBuildNoArg("daily_songs", base),
			mustBuildNoArg("daily_playlists", base),
			mustBuild("user_playlist", base, UserPlaylistOpts{UserID: CurUser}),
			mustBuildNoArg("user_collect", base),
			mustBuildNoArg("personal_fm", base),
			mustBuildNoArg("search_type", base),
			mustBuildNoArg("ranks", base),
			mustBuildNoArg("high_quality_playlists", base),
			mustBuildNoArg("recent_songs", base),
			mustBuildNoArg("could", base),
			nil, // 帮助由 Action 直接打开 Markdown 弹窗，不再进入子菜单。
		},
	}

	// 追加插件声明的主菜单项（Phase 3.9）。无 Build 的插件主菜单项 MUST 是
	// 无参菜单：经 mustBuildNoArg 构建，key 未注册或为参数化菜单会在启动时
	// panic（程序错误，作为启动完整性信号；先显式断言以给出清晰错误）。
	// 带 Build 的项由插件以自身 options 构造菜单（参数化 provider 入口）。
	// 触发由插件菜单自身的 Action / BeforeEnterMenuHook 承担——主菜单不再
	// 对插件索引做特判（检查更新的 index-15 特判已随插件化移除）。
	for _, item := range MainMenuPluginItems() {
		if _, ok := menuRegistry[item.Key]; !ok {
			panic(fmt.Sprintf("main menu plugin item %q: menu provider not registered", item.Key))
		}
		mainMenu.menus = append(mainMenu.menus, model.MenuItem{Title: item.Title})
		if item.Build != nil {
			mainMenu.menuList = append(mainMenu.menuList, item.Build(base))
		} else {
			mainMenu.menuList = append(mainMenu.menuList, mustBuildNoArg(item.Key, base))
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
	case mainMenuHelpIndex:
		showHelpPopup(app)
		return app.MustMain(), nil
	default:
		// 插件主菜单项（如检查更新）不再特判：Action 落到默认分支（非播放
		// 菜单返回 nil/nil），由 SubMenu 进入插件菜单，触发由插件菜单自身的
		// Action / BeforeEnterMenuHook 承担。
		return m.baseMenu.Action(app, index)
	}
}
