package ui

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
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
// entry (menu == nil) is recognized for the helpIndex computation. pluginID is
// the plugin scope the entry was declared in (empty for built-ins). Since P5
// the 9 business plugins register their items only when enabled (disabled =
// not started = not registered), so the pluginID-based filtering below is a
// defense-in-depth for residual registrations that register unconditionally
// (WASM command-menu items adapted by registerCommandMenus); NewMainMenu hides
// entries whose plugin is disabled after the chain walk.
type mainMenuEntry struct {
	key      string
	after    string
	title    string
	menu     Menu
	builtin  bool
	pluginID string
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
	// Since P5 the 9 business plugins register their main-menu items only when
	// enabled (a disabled plugin never starts, so its item is absent from this
	// list); the IsPluginEnabled gates below are defense-in-depth for items
	// registered unconditionally (WASM command-menu items), and
	// orderMainMenuEntries tolerates a disabled plugin's missing anchor by
	// re-anchoring the dependent entries to the chain tail (relaxed mode).
	for _, item := range MainMenuPluginItems() {
		if _, ok := menuRegistry[item.Key]; !ok {
			panic(fmt.Sprintf("main menu plugin item %q: menu provider not registered", item.Key))
		}
		entry := mainMenuEntry{key: item.Key, after: item.After, title: item.Title, pluginID: item.PluginID}
		if IsPluginEnabled(item.PluginID) {
			if item.Build != nil {
				entry.menu = item.Build(base)
			} else {
				entry.menu = mustBuildNoArg(item.Key, base)
			}
		}
		entries = append(entries, entry)
	}
	for _, e := range orderMainMenuEntries(entries) {
		if !IsPluginEnabled(e.pluginID) {
			continue
		}
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
//     entry's key. With all plugins enabled a missing anchor is a programmer
//     error and panics; under the P5 "disabled = nonexistent" semantics (any
//     plugin disabled in [plugins]) a missing anchor is a config outcome — the
//     anchored-after plugin's item was never registered — so the entry is
//     re-anchored to the chain tail (with its followers) and the build
//     continues;
//   - every entry is reachable exactly once — the chain length must equal the
//     total entry count, which catches orphaned entries (their After anchor
//     was duplicated by an earlier entry or they close an unreachable cycle)
//     and cycles (an entry re-visited while walking from MainMenuStart). In
//     relaxed mode (a plugin disabled) orphan detection is skipped: a broken
//     chain makes exact positioning impossible, so the remaining entries are
//     appended in declaration order to keep the menu buildable;
//   - entries with an empty After (the end-append convenience forms) follow
//     the chain in registration order.
func orderMainMenuEntries(entries []mainMenuEntry) []mainMenuEntry {
	// Every After target must exist: MainMenuStart or another entry's key.
	known := make(map[string]bool, len(entries)+1)
	known[MainMenuStart] = true
	for _, e := range entries {
		known[e.key] = true
	}

	// relaxed is on when any plugin is disabled in [plugins]: the disabled
	// plugin's items are not registered at all (P5 disabled = nonexistent), so
	// dependent After anchors can legitimately go missing. The TUI-connect
	// shell (S6) is always relaxed: the frontend scope that registers the 9
	// business plugins is skipped (B10), so every plugin anchor is absent.
	relaxed := connectMode || (configs.AppConfig != nil && len(configs.AppConfig.Plugins.Disabled) > 0)

	// Index entries by their After anchor; entries with a missing anchor are
	// deferred (relaxed mode re-anchors them at the tail). Duplicate anchors
	// collapse to one entry (the other becomes orphaned — caught by the length
	// assertion below). End-append entries (empty After) are excluded from the
	// walk.
	byAfter := make(map[string]mainMenuEntry, len(entries))
	var deferred []mainMenuEntry
	var missing []string
	for _, e := range entries {
		if e.after == "" {
			continue
		}
		if !known[e.after] {
			missing = append(missing, fmt.Sprintf("%q (after %q)", e.key, e.after))
			deferred = append(deferred, e)
			continue
		}
		byAfter[e.after] = e
	}
	if len(missing) > 0 {
		if !relaxed {
			panic("main menu chain: After anchor not registered: " + strings.Join(missing, ", "))
		}
		if connectMode {
			slog.Warn("main menu chain: plugin anchor not mounted in connect mode (lastfm excluded; disabled plugins), re-anchoring entry to the chain tail", "missing", strings.Join(missing, ", "))
		} else {
			slog.Warn("main menu chain: After anchor not registered (plugin disabled?), re-anchoring entries to the chain tail", "missing", strings.Join(missing, ", "))
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
	// Relaxed mode: re-anchor the deferred entries (missing anchor) and their
	// followers at the chain tail. Followers are the entries whose After names
	// the just-appended key (the byAfter links), so the tail keeps the original
	// chain order.
	for _, e := range deferred {
		appendChainTail(e, &ordered, visited, byAfter)
	}
	if len(ordered) != len(entries) {
		if relaxed {
			// Last resort: a broken chain makes exact positioning impossible;
			// append the still-unvisited entries in declaration order so the
			// menu stays buildable (no startup panic from config).
			for _, e := range entries {
				appendChainTail(e, &ordered, visited, byAfter)
			}
		} else {
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
	}
	return ordered
}

// appendChainTail appends head and its followers (entries whose After names the
// head's key, transitively via byAfter) to ordered, marking them visited. It is
// the tail-append used by the relaxed (a plugin disabled) chain re-anchoring
// and the declaration-order last resort.
func appendChainTail(head mainMenuEntry, ordered *[]mainMenuEntry, visited map[string]bool, byAfter map[string]mainMenuEntry) {
	if visited[head.key] {
		return
	}
	visited[head.key] = true
	*ordered = append(*ordered, head)
	cur := head.key
	for {
		next, ok := byAfter[cur]
		if !ok || visited[next.key] {
			break
		}
		visited[next.key] = true
		*ordered = append(*ordered, next)
		cur = next.key
	}
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
