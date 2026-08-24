package ui

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
)

type MainMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
	// helpIndex is the merged-menu index of the built-in 帮助 entry (its
	// submenu placeholder is nil and Action opens the Markdown popup directly).
	// Computed at construction so it stays correct for any subset of the
	// plugin set (e.g. test binaries); in the full production set it is 15.
	helpIndex int
}

// mainMenuEntry is one row of the merged built-in + plugin main menu before
// the after-anchor chain walk. builtin marks built-in entries so the help
// entry (menu == nil) is recognized for the helpIndex computation.
type mainMenuEntry struct {
	key     string
	after   string
	title   string
	menu    Menu
	builtin bool
}

func NewMainMenu(base baseMenu) *MainMenu {
	mainMenu := &MainMenu{baseMenu: base}

	// 内置项与插件项经 After 锚点链归并，复现插件化前的主菜单原始顺序：
	// 每日推荐歌曲 / 每日推荐歌单 / 我的歌单 / 我的收藏 / 私人FM / 专辑列表 /
	// 搜索 / 排行榜 / 精选歌单 / 热门歌手 / 最近播放歌曲 / 云盘 / 主播电台 /
	// LastFM / 帮助 / 检查更新。「搜索」主菜单项不再内置——现由 search 插件经
	// RegisterMainMenuItemAfter("search_type", "搜索", "album_menu", nil) 声明
	// （After 锚点 album_menu，位置与内置时一致）。每个入口声明其前驱项 key
	// （After），插入一个菜单只需声明一个锚点，其余项不漂移。链从 MainMenuStart
	// 走到无处可去，空 After 的项（RegisterMainMenuItem /
	// RegisterMainMenuItemWith 便捷形式）追加在链尾（注册序保持，既有"追加在
	// 末尾"行为）。无 Build 的插件主菜单项 MUST 是无参菜单：经 mustBuildNoArg
	// 构建，key 未注册或为参数化菜单会在启动时 panic（程序错误，作为启动完整
	// 性信号；先显式断言以给出清晰错误）。带 Build 的项由插件以自身 options
	// 构造菜单（参数化 provider 入口）。触发由插件菜单自身的 Action /
	// BeforeEnterMenuHook 承担——主菜单不再对插件索引做特判。
	entries := []mainMenuEntry{
		// 内置项也参与锚点链：帮助跟在 LastFM 后（搜索项已由 search 插件提供）。
		{key: "help", after: "last_fm", builtin: true, title: "帮助", menu: nil}, // 帮助由 Action 直接打开 Markdown 弹窗，不再进入子菜单。
	}
	for _, item := range MainMenuPluginItems() {
		if _, ok := menuRegistry[item.Key]; !ok {
			panic(fmt.Sprintf("main menu plugin item %q: menu provider not registered", item.Key))
		}
		entry := mainMenuEntry{key: item.Key, after: item.After, title: item.Title}
		if item.Build != nil {
			entry.menu = item.Build(base)
		} else {
			entry.menu = mustBuildNoArg(item.Key, base)
		}
		entries = append(entries, entry)
	}
	for _, e := range orderMainMenuEntries(entries) {
		mainMenu.menus = append(mainMenu.menus, model.MenuItem{Title: e.title})
		mainMenu.menuList = append(mainMenu.menuList, e.menu)
		if e.builtin && e.menu == nil {
			mainMenu.helpIndex = len(mainMenu.menus) - 1
		}
	}
	return mainMenu
}

// orderMainMenuEntries walks the after-anchor chain from MainMenuStart and
// asserts the chain's integrity with explicit panics (programmer errors):
//
//   - every After target exists — each anchor must be MainMenuStart or another
//     entry's key;
//   - every entry is reachable exactly once — the chain length must equal the
//     total entry count, which catches orphaned entries (their After anchor
//     was duplicated by an earlier entry or they close an unreachable cycle)
//     and cycles (an entry re-visited while walking from MainMenuStart);
//   - entries with an empty After (the end-append convenience forms) follow
//     the chain in registration order.
func orderMainMenuEntries(entries []mainMenuEntry) []mainMenuEntry {
	// Every After target must exist: MainMenuStart or another entry's key.
	known := make(map[string]bool, len(entries)+1)
	known[MainMenuStart] = true
	for _, e := range entries {
		known[e.key] = true
	}
	var missing []string
	for _, e := range entries {
		if e.after != "" && !known[e.after] {
			missing = append(missing, fmt.Sprintf("%q (after %q)", e.key, e.after))
		}
	}
	if len(missing) > 0 {
		panic("main menu chain: After anchor not registered: " + strings.Join(missing, ", "))
	}

	// Index entries by their After anchor; duplicate anchors collapse to one
	// entry (the other becomes orphaned — caught by the length assertion
	// below). End-append entries (empty After) are excluded from the walk.
	byAfter := make(map[string]mainMenuEntry, len(entries))
	for _, e := range entries {
		if e.after != "" {
			byAfter[e.after] = e
		}
	}

	ordered := make([]mainMenuEntry, 0, len(entries))
	visited := make(map[string]bool, len(entries))
	cur := MainMenuStart
	for {
		next, ok := byAfter[cur]
		if !ok {
			break
		}
		if visited[next.key] {
			panic(fmt.Sprintf("main menu chain: cycle or duplicate After anchor at %q", next.key))
		}
		visited[next.key] = true
		ordered = append(ordered, next)
		cur = next.key
	}
	// End-append entries (empty After) follow the chain in registration order.
	for _, e := range entries {
		if e.after == "" && !visited[e.key] {
			visited[e.key] = true
			ordered = append(ordered, e)
		}
	}
	if len(ordered) != len(entries) {
		var orphans []string
		for _, e := range entries {
			if !visited[e.key] {
				orphans = append(orphans, e.key)
			}
		}
		sort.Strings(orphans)
		panic(fmt.Sprintf("main menu chain: %d of %d entries reachable from %q (orphaned entries — missing or duplicate After anchor): %v",
			len(ordered), len(entries), MainMenuStart, orphans))
	}
	return ordered
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
