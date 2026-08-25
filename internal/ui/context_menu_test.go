package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/playlist"
)

// withFreshContextMenuContribs 保存并清空包级右键菜单贡献注册表，测试结束
// 后恢复原状——测试内注册的项不会污染其它测试（不引入公开清理 API）。
func withFreshContextMenuContribs(t *testing.T) {
	t.Helper()
	orig := contextMenuContribs
	contextMenuContribs = nil
	t.Cleanup(func() { contextMenuContribs = orig })
}

// --- 注册表 ---

func TestRegisterContextMenuContribValidation(t *testing.T) {
	withFreshContextMenuContribs(t)

	// 空 Title / nil Action 直接 panic 拒绝。
	assertPanics(t, func() {
		RegisterContextMenuContrib(ContextMenuContrib{Title: ""})
	})
	assertPanics(t, func() {
		RegisterContextMenuContrib(ContextMenuContrib{Title: "空 Action"})
	})

	// 合法注册不 panic，且可被快照查询到。
	RegisterContextMenuContrib(ContextMenuContrib{
		Title:  "合法项",
		Action: func(MenuServices, ContextMenuContext) (model.Page, tea.Cmd) { return nil, nil },
	})
	if got := len(ContextMenuContribs()); got != 1 {
		t.Fatalf("registered contrib count = %d, want 1", got)
	}
}

func TestContextMenuContribsSnapshot(t *testing.T) {
	withFreshContextMenuContribs(t)
	RegisterContextMenuContrib(ContextMenuContrib{
		Title:  "快照项",
		Action: func(MenuServices, ContextMenuContext) (model.Page, tea.Cmd) { return nil, nil },
	})
	before := len(ContextMenuContribs())
	snap := ContextMenuContribs()
	_ = append(snap, ContextMenuContrib{Title: "快照污染"})
	if got := len(ContextMenuContribs()); got != before {
		t.Fatalf("ContextMenuContribs() mutated by appending to a snapshot: %d -> %d", before, got)
	}
}

// --- 插件组构建 ---

func TestBuildPluginContextMenuItems(t *testing.T) {
	withFreshContextMenuContribs(t)
	RegisterContextMenuContrib(ContextMenuContrib{
		Title:  "始终显示",
		Action: func(MenuServices, ContextMenuContext) (model.Page, tea.Cmd) { return nil, nil },
	})
	RegisterContextMenuContrib(ContextMenuContrib{
		Title:  "仅选中时",
		Show:   func(ctx ContextMenuContext) bool { return ctx.SelectedIndex >= 0 },
		Action: func(MenuServices, ContextMenuContext) (model.Page, tea.Cmd) { return nil, nil },
	})
	// nil Show 视为恒显示（调用方按需决定是否兜底）。
	RegisterContextMenuContrib(ContextMenuContrib{
		Title:  "Nil Show 恒显示",
		Action: func(MenuServices, ContextMenuContext) (model.Page, tea.Cmd) { return nil, nil },
	})

	// 未选中（-1）：「仅选中时」隐藏，剩 header + 2 项。
	items := buildPluginContextMenuItems(ContextMenuContext{SelectedIndex: -1})
	if len(items) != 3 {
		t.Fatalf("plugin context items = %d, want 3 (header + 2 visible items)", len(items))
	}
	if !items[0].Header || items[0].Label != "插件" {
		t.Fatalf("items[0] = %#v, want 插件 header", items[0])
	}
	if got, want := items[1].ID, "plugin:0"; got != want {
		t.Errorf("items[1].ID = %q, want %q", got, want)
	}
	if got, want := items[1].Label, itemIndent+"始终显示"; got != want {
		t.Errorf("items[1].Label = %q, want %q", got, want)
	}
	// 注册序号按全局注册表（0 和 2），跳过 1。
	if got, want := items[2].ID, "plugin:2"; got != want {
		t.Errorf("items[2].ID = %q, want %q", got, want)
	}
	if got, want := items[2].Label, itemIndent+"Nil Show 恒显示"; got != want {
		t.Errorf("items[2].Label = %q, want %q", got, want)
	}

	// 选中时全部显示，注册序保持。
	selItems := buildPluginContextMenuItems(ContextMenuContext{SelectedIndex: 3})
	if len(selItems) != 4 {
		t.Fatalf("selected plugin context items = %d, want 4 (header + 3 items)", len(selItems))
	}
	if got := selItems[2].ID; got != "plugin:1" {
		t.Errorf("selected items[2].ID = %q, want plugin:1", got)
	}
}

func TestBuildPluginContextMenuItemsEmpty(t *testing.T) {
	withFreshContextMenuContribs(t)
	if items := buildPluginContextMenuItems(ContextMenuContext{}); items != nil {
		t.Fatalf("no-registered-contribs items = %#v, want nil", items)
	}

	// 全部 Show 返回 false 同样无项。
	RegisterContextMenuContrib(ContextMenuContrib{
		Title:  "不显示",
		Show:   func(ContextMenuContext) bool { return false },
		Action: func(MenuServices, ContextMenuContext) (model.Page, tea.Cmd) { return nil, nil },
	})
	if items := buildPluginContextMenuItems(ContextMenuContext{}); items != nil {
		t.Fatalf("all-hidden items = %#v, want nil", items)
	}
}

// --- 分发 ---

func TestHandlePluginContextAction(t *testing.T) {
	withFreshContextMenuContribs(t)
	var gotCtx ContextMenuContext
	cmd := func() tea.Msg { return model.ShowNotificationMsg{} }
	RegisterContextMenuContrib(ContextMenuContrib{
		Title: "分发项",
		Action: func(_ MenuServices, ctx ContextMenuContext) (model.Page, tea.Cmd) {
			gotCtx = ctx
			return nil, cmd
		},
	})

	// 非法 / 越界 ID 均返回 nil, nil，不 panic。
	if page, cmd := handlePluginContextAction(nil, ContextMenuContext{}, "plugin:bad"); page != nil || cmd != nil {
		t.Fatalf("plugin:bad → page=%T cmd=%v, want nil/nil", page, cmd)
	}
	if page, cmd := handlePluginContextAction(nil, ContextMenuContext{}, "plugin:5"); page != nil || cmd != nil {
		t.Fatalf("plugin:5 → page=%T cmd=%v, want nil/nil", page, cmd)
	}
	if page, cmd := handlePluginContextAction(nil, ContextMenuContext{}, "plugin:"); page != nil || cmd != nil {
		t.Fatalf("plugin: → page=%T cmd=%v, want nil/nil", page, cmd)
	}
	if page, cmd := handlePluginContextAction(nil, ContextMenuContext{}, "noplugin:0"); page != nil || cmd != nil {
		t.Fatalf("noplugin:0 → page=%T cmd=%v, want nil/nil", page, cmd)
	}

	// 有效分发：ctx 原样透传，返回 Action 的 (page, cmd)。
	page, cmd := handlePluginContextAction(nil, ContextMenuContext{SelectedIndex: 2}, "plugin:0")
	if page != nil || cmd == nil {
		t.Fatalf("plugin:0 → page=%T cmd=%T, want nil page + non-nil cmd", page, cmd)
	}
	if gotCtx.SelectedIndex != 2 {
		t.Errorf("Action ctx.SelectedIndex = %d, want 2", gotCtx.SelectedIndex)
	}
}

// --- BaseMenu 集成 ---

// contextMenuTestApp 构造一个当前菜单为普通菜单（非 action_menu）的 foxful
// app，并把 Netease 壳接到含全部启动服务的全新 framework context（Player
// 携带已初始化的播放列表面板）。返回壳供 baseMenu 构造。
func contextMenuTestApp(t *testing.T) (*model.App, *Netease) {
	t.Helper()
	app, _ := newFormPageTestApp(t)
	n := testNetease()
	n.App = app
	n.ctx = &framework.Context{}
	n.player = &Player{playlistManager: playlist.NewPlaylistManager()}
	if err := registerServices(n.ctx, n); err != nil {
		t.Fatalf("registerServices() error = %v", err)
	}
	return app, n
}

// TestBaseMenuContextMenuItemsAppendsPluginGroup 验证 ContextMenuItems 把
// 插件组追加在 generic 全局组之后：无选中、无播放时结果为
// [generic:refresh, generic:switchTheme, separator, 插件 header, plugin:0]。
func TestBaseMenuContextMenuItemsAppendsPluginGroup(t *testing.T) {
	withFreshContextMenuContribs(t)
	app, n := contextMenuTestApp(t)
	RegisterContextMenuContrib(ContextMenuContrib{
		Title:  "菜单集成项",
		Action: func(MenuServices, ContextMenuContext) (model.Page, tea.Cmd) { return nil, nil },
	})
	RegisterContextMenuContrib(ContextMenuContrib{
		Title:  "隐藏项",
		Show:   func(ContextMenuContext) bool { return false },
		Action: func(MenuServices, ContextMenuContext) (model.Page, tea.Cmd) { return nil, nil },
	})

	base := newBaseMenu(n)
	items := base.ContextMenuItems(app, -1)
	if len(items) != 5 {
		t.Fatalf("context items = %d, want 5 (refresh, switchTheme, separator, 插件 header, 插件项)", len(items))
	}
	if got := items[0].ID; got != "generic:refresh" {
		t.Fatalf("items[0].ID = %q, want generic:refresh (generic 全局组在插件组之前)", got)
	}
	if !items[2].Separator {
		t.Fatalf("items[2] = %#v, want separator before plugin group", items[2])
	}
	if !items[3].Header || items[3].Label != "插件" {
		t.Fatalf("items[3] = %#v, want 插件 header", items[3])
	}
	if got, want := items[4].ID, "plugin:0"; got != want {
		t.Fatalf("items[4].ID = %q, want %q", got, want)
	}
	if got, want := items[4].Label, itemIndent+"菜单集成项"; got != want {
		t.Fatalf("items[4].Label = %q, want %q", got, want)
	}
}

func TestBaseMenuContextMenuActionDispatchesPluginItem(t *testing.T) {
	withFreshContextMenuContribs(t)
	app, n := contextMenuTestApp(t)
	var gotSelectedIndex int
	cmd := func() tea.Msg { return model.ShowNotificationMsg{} }
	RegisterContextMenuContrib(ContextMenuContrib{
		Title: "动作分发项",
		Action: func(_ MenuServices, ctx ContextMenuContext) (model.Page, tea.Cmd) {
			gotSelectedIndex = ctx.SelectedIndex
			return nil, cmd
		},
	})

	base := newBaseMenu(n)
	page, cmd := base.ContextMenuAction(app, 3, model.ContextMenuItem{ID: "plugin:0"})
	if page != nil || cmd == nil {
		t.Fatalf("plugin action → page=%T cmd=%T, want nil page + non-nil cmd", page, cmd)
	}
	if gotSelectedIndex != 3 {
		t.Errorf("Action ctx.SelectedIndex = %d, want 3 (from ContextMenuAction index param)", gotSelectedIndex)
	}

	// 越界序号不 panic，返回 nil, nil。
	page, cmd = base.ContextMenuAction(app, 3, model.ContextMenuItem{ID: "plugin:99"})
	if page != nil || cmd != nil {
		t.Fatalf("plugin:99 → page=%T cmd=%T, want nil/nil", page, cmd)
	}
}
