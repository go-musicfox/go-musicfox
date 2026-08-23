package ui

import (
	"context"
	"io"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/charmbracelet/x/ansi"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

// TestSetupI18nRegistersChineseCatalog 验证中文 catalog 注册后 foxful-cli
// 内置文案返回中文，且空 locale 默认回落到中文。
func TestSetupI18nRegistersChineseCatalog(t *testing.T) {
	prev := model.DefaultCatalog().Locale()
	t.Cleanup(func() { model.SetLocale(prev) })

	SetupI18n("")
	if got := model.DefaultCatalog().Locale(); got != "zh" {
		t.Fatalf("empty locale should default to zh, got %q", got)
	}
	if got := model.T(model.MsgConfirm); got != "确认" {
		t.Errorf("MsgConfirm = %q, want 确认", got)
	}
	if got := model.T(model.MsgCancel); got != "取消" {
		t.Errorf("MsgCancel = %q, want 取消", got)
	}

	// 显式 locale 生效
	SetupI18n("en")
	if got := model.DefaultCatalog().Locale(); got != "en" {
		t.Errorf("locale = %q, want en", got)
	}
	if got := model.T(model.MsgConfirm); got != "Confirm" {
		t.Errorf("en MsgConfirm = %q, want Confirm", got)
	}
}

// TestToastLevelToModel 验证 notify.ToastLevel 到 foxful-cli 级别的映射。
func TestToastLevelToModel(t *testing.T) {
	cases := map[notify.ToastLevel]model.NotificationLevel{
		notify.ToastInfo:    model.NotificationInfo,
		notify.ToastSuccess: model.NotificationSuccess,
		notify.ToastWarning: model.NotificationWarning,
		notify.ToastError:   model.NotificationError,
	}
	for in, want := range cases {
		if got := toastLevelToModel(in); got != want {
			t.Errorf("toastLevelToModel(%d) = %d, want %d", in, got, want)
		}
	}
}

// TestBuildHelpMarkdown 验证帮助弹窗文档为合法 Markdown 且包含关键结构。
func TestBuildHelpMarkdown(t *testing.T) {
	keybindings.InitDefaults(true)
	md := buildHelpMarkdown()
	if !strings.Contains(md, "# go-musicfox 快捷键") {
		t.Errorf("help markdown missing title heading:\n%s", md)
	}
	if !strings.Contains(md, "| 按键 | 操作 |") {
		t.Errorf("help markdown missing table header:\n%s", md)
	}
	// 应能被 foxful-cli 的 markdown 弹窗成功渲染（不返回错误）。
	if _, err := model.NewMarkdownPopup(model.MarkdownPopupSpec{
		Title:           " 帮助 ",
		MarkdownContent: md,
		MaxWidth:        72,
	}); err != nil {
		t.Errorf("NewMarkdownPopup failed: %v", err)
	}
}

func TestMainMenuHelpActionShowsMarkdownPopup(t *testing.T) {
	keybindings.InitDefaults(true)
	app, netease := newFormPageTestApp(t)
	menu := NewMainMenu(newBaseMenu(netease))
	helpIndex := -1
	for i, item := range menu.menus {
		if item.Title == "帮助" {
			helpIndex = i
			break
		}
	}
	if helpIndex < 0 {
		t.Fatal("main menu does not contain help entry")
	}

	page, cmd := menu.Action(app, helpIndex)
	if page != app.MustMain() || cmd != nil {
		t.Fatalf("help action returned page=%T cmd=%v, want current main page and nil command", page, cmd)
	}
	if !app.HasPopup() {
		t.Fatal("help action did not show a Markdown popup")
	}
	if submenu := menu.SubMenu(app, helpIndex); submenu != nil {
		t.Fatalf("help entry still opens full-page submenu %T", submenu)
	}
}

func TestMainMenuPluginItemRoutesToPluginMenu(t *testing.T) {
	// The "check_update" provider and its main-menu entry ship with the
	// external-style plugin (internal/plugins/checkupdate), which this ui test
	// binary cannot link (ui must not import plugins). Register behavior-
	// equivalent test-doubles so the plugin main-menu item is exercised.
	RegisterMenu("check_update", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return &testCheckUpdateMenu{baseMenu: base}, nil
	})
	RegisterMainMenuItem("check_update", "检查更新")

	app, netease := newFormPageTestApp(t)
	menu := NewMainMenu(newBaseMenu(netease))
	updateIndex := -1
	for i, item := range menu.menus {
		if item.Title == "检查更新" {
			updateIndex = i
			break
		}
	}
	if updateIndex < 0 {
		t.Fatal("main menu does not contain check-update entry")
	}

	// The main menu no longer special-cases the check-update index: Action
	// falls through (nil/nil) and the trigger is handled by the plugin menu's
	// own BeforeEnterMenuHook/Action.
	if page, cmd := menu.Action(app, updateIndex); page != nil || cmd != nil {
		t.Fatalf("main menu check-update action = page=%T cmd=%v, want nil/nil (fall through to plugin menu)", page, cmd)
	}
	submenu := menu.SubMenu(app, updateIndex)
	if submenu == nil {
		t.Fatal("check-update entry does not route to a plugin submenu")
	}
	if key := submenu.GetMenuKey(); key != "check_update" {
		t.Fatalf("check-update submenu key = %q, want check_update", key)
	}
}

// testCheckUpdateMenu is the ui test-double for the plugin-supplied
// "check_update" provider: its Action returns the current main page plus a
// notification command, mirroring the CheckUpdateMenu contract in
// internal/plugins/checkupdate.
type testCheckUpdateMenu struct {
	baseMenu
}

func (m *testCheckUpdateMenu) GetMenuKey() string          { return "check_update" }
func (m *testCheckUpdateMenu) MenuViews() []model.MenuItem { return nil }
func (m *testCheckUpdateMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}
func (m *testCheckUpdateMenu) Action(a *model.App, _ int) (model.Page, tea.Cmd) {
	return a.MustMain(), func() tea.Msg { return model.ShowNotificationMsg{} }
}

// testAlbumSubListMenu is the ui test-double for the plugin-supplied
// "album_sub_list" provider (the album cluster moved into
// internal/plugins/album, Phase 3.9.x). The built-in 我的收藏 menu
// (NewUserCollectionMenu) builds its album_sub_list sub-menu through
// mustBuildNoArg, which needs the key registered in this ui test binary — the
// plugin cannot be linked here (ui must not import plugins).
type testAlbumSubListMenu struct {
	baseMenu
}

func (m *testAlbumSubListMenu) GetMenuKey() string          { return "album_sub_list" }
func (m *testAlbumSubListMenu) MenuViews() []model.MenuItem { return nil }
func (m *testAlbumSubListMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

// testArtistsSubListMenu is the ui test-double for the plugin-supplied
// "artists_sub_list" provider (the artist cluster moved into
// internal/plugins/artist, Phase 3.9.x). Same rationale as the album double:
// NewUserCollectionMenu builds its artists_sub_list sub-menu through
// mustBuildNoArg, which needs the key registered in this ui test binary.
type testArtistsSubListMenu struct {
	baseMenu
}

func (m *testArtistsSubListMenu) GetMenuKey() string          { return "artists_sub_list" }
func (m *testArtistsSubListMenu) MenuViews() []model.MenuItem { return nil }
func (m *testArtistsSubListMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

// init registers the album_sub_list and artists_sub_list test-doubles so
// NewMainMenu / the built-in user_collect menu construction works in this ui
// test binary (the album/artist plugins' init() registrations are not linked
// here).
func init() {
	RegisterMenu("album_sub_list", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return &testAlbumSubListMenu{baseMenu: base}, nil
	})
	RegisterMenu("artists_sub_list", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return &testArtistsSubListMenu{baseMenu: base}, nil
	})
}

// TestCheckUpdateResultRendersDirectlyInTUI 验证 model.ShowNotificationMsg
// 经 app.Update 直接渲染为 TUI 通知。检查更新消息的构造已随插件提取移入
// internal/plugins/checkupdate，这里内联构造等效消息验证框架渲染路径。
func TestCheckUpdateResultRendersDirectlyInTUI(t *testing.T) {
	app, _ := newFormPageTestApp(t)
	msg := model.ShowNotificationMsg{Spec: model.NotificationSpec{
		Level:   model.NotificationSuccess,
		Title:   "检查更新",
		Message: types.AppVersion + " 已是最新版本",
	}}
	_, _ = app.Update(msg)

	if view := app.View().Content; !strings.Contains(view, "已是最新版本") {
		t.Fatalf("TUI view does not contain check-update result: %q", view)
	}
}

func TestHelpPopupFitsTerminalHeight(t *testing.T) {
	keybindings.InitDefaults(true)
	app, _ := newFormPageTestApp(t)
	showHelpPopup(app)

	const terminalHeight = 24
	if got := lipgloss.Height(app.View().Content); got != terminalHeight {
		t.Fatalf("view height with help popup = %d, want terminal height %d", got, terminalHeight)
	}
}

func TestDualColumnSelectedMenuItemDoesNotWrap(t *testing.T) {
	const (
		width  = 75
		height = 24
		title  = "我的音乐你听吗 第9期"
	)

	app := newDualColumnOverflowTestApp(t, width, height)
	app.MustMain().EnterMenu(nil, nil)

	view := ansi.Strip(app.View().Content)
	backButtonFound := false
	menuRowFound := false
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > width {
			t.Fatalf("rendered line width = %d, terminal width = %d: %q", lipgloss.Width(line), width, line)
		}
		if strings.Contains(line, "← 收藏专辑") {
			backButtonFound = true
		}
		if strings.Contains(line, title) && strings.Contains(line, "另一张专辑") {
			menuRowFound = true
		}
	}
	if !backButtonFound {
		t.Fatalf("submenu title is missing its back button:\n%s", view)
	}
	if !menuRowFound {
		t.Fatalf("dual-column menu row wrapped:\n%s", view)
	}
}

func TestUpdateNotificationOpensVersionRelease(t *testing.T) {
	content := notify.NotifyContent{
		Title:       "发现新版本: v9.9.9",
		Text:        "去看看呗",
		Url:         types.AppGithubUrl + "/releases/tag/v9.9.9",
		ActionLabel: "前往 GitHub",
	}
	wantURL := types.AppGithubUrl + "/releases/tag/v9.9.9"
	if content.Url != wantURL || content.ActionLabel != "前往 GitHub" {
		t.Fatalf("update notification = %#v, want URL %q with GitHub action", content, wantURL)
	}

	openedURL := ""
	spec := BuildToastNotificationSpec(content, notify.ToastInfo, func(url string) error {
		openedURL = url
		return nil
	})
	if len(spec.Actions) != 1 {
		t.Fatalf("notification actions = %d, want 1", len(spec.Actions))
	}
	action := spec.Actions[0]
	if action.ID != toastOpenURLActionID || action.Label != content.ActionLabel {
		t.Fatalf("notification action = %#v", action)
	}
	if spec.OnAction == nil {
		t.Fatal("notification URL action callback is nil")
	}
	spec.OnAction(model.NotificationActionResult{ActionID: "other"})
	if openedURL != "" {
		t.Fatalf("unrelated action opened %q", openedURL)
	}
	spec.OnAction(model.NotificationActionResult{ActionID: toastOpenURLActionID})
	if openedURL != wantURL {
		t.Fatalf("opened URL = %q, want %q", openedURL, wantURL)
	}
}

type dualColumnOverflowParentMenu struct {
	model.DefaultMenu
}

func (dualColumnOverflowParentMenu) GetMenuKey() string {
	return "dual-column-overflow-parent"
}

func (dualColumnOverflowParentMenu) MenuViews() []model.MenuItem {
	return []model.MenuItem{{Title: "收藏专辑"}}
}

func (dualColumnOverflowParentMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return &dualColumnOverflowChildMenu{}
}

func (dualColumnOverflowParentMenu) HelpHints() []model.HelpHint {
	return nil
}

type dualColumnOverflowChildMenu struct {
	model.DefaultMenu
}

func (dualColumnOverflowChildMenu) GetMenuKey() string {
	return "dual-column-overflow-child"
}

func (dualColumnOverflowChildMenu) MenuViews() []model.MenuItem {
	return []model.MenuItem{
		{Title: "我的音乐你听吗 第9期", Subtitle: "[我的音乐你听吗]"},
		{Title: "另一张专辑"},
	}
}

func (dualColumnOverflowChildMenu) HelpHints() []model.HelpHint {
	return nil
}

func newDualColumnOverflowTestApp(t *testing.T, width, height int) *model.App {
	t.Helper()

	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Theme.ShowTitle = true
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	opts := model.DefaultOptions()
	opts.EnableStartup = false
	opts.WhetherDisplayTitle = true
	opts.DualColumn = true
	opts.MainMenu = &dualColumnOverflowParentMenu{}
	opts.MainMenuTitle = &model.MenuItem{Title: "我的音乐"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opts.TeaOptions = []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithInput(nil),
		tea.WithOutput(io.Discard),
	}

	app := model.NewApp(opts)
	_ = app.Run()
	_, _ = app.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return app
}

// TestConfirmPopupDefaultsToCancelAndRunsOnConfirm verifies the destructive-
// action safety property of showConfirmPopup: 取消 is the default-focused
// button (a bare Enter cancels and does NOT run onConfirm), while focusing
// 确定 via Tab then Enter runs onConfirm. Also asserts the popup renders both
// buttons and is pushed onto the modal stack.
func TestConfirmPopupDefaultsToCancelAndRunsOnConfirm(t *testing.T) {
	app, _ := newFormPageTestApp(t)

	ran := false
	showConfirmPopup(app, "测试确认", "确定执行吗？", func() { ran = true })

	if !app.HasPopup() {
		t.Fatal("showConfirmPopup did not push a popup onto the modal stack")
	}
	if view := app.View().Content; !strings.Contains(view, "取消") || !strings.Contains(view, "确定") {
		t.Fatalf("confirm popup view missing action buttons: %q", view)
	}

	// Default focus is 取消: a bare Enter must cancel without running onConfirm.
	app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if ran {
		t.Fatal("onConfirm ran on bare Enter; 取消 must be the default-focused button")
	}
	if app.HasPopup() {
		t.Fatal("bare Enter did not dismiss the popup")
	}

	// Re-show, move focus to 确定 (Tab), then confirm (Enter).
	showConfirmPopup(app, "测试确认", "确定执行吗？", func() { ran = true })
	app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !ran {
		t.Fatal("onConfirm did not run after focusing 确定 and pressing Enter")
	}
}

// dualColumnTruncateMenu reproduces the daily-recommend layout where a
// left-column title overflows the fixed 44-cell budget. The overflowing title
// is crafted so truncation stops one cell short (right before a wide rune),
// which previously left the item under-filled and shifted that row's right
// column left by one cell.
type dualColumnTruncateMenu struct {
	model.DefaultMenu
}

func (dualColumnTruncateMenu) GetMenuKey() string { return "dual-column-truncate" }

func (dualColumnTruncateMenu) MenuViews() []model.MenuItem {
	titles := []string{
		"私人雷达", // 0 (left)
		"华语必听100首，金曲新歌全都有", // 1 (right)
		"滚石唱片", // 2 (left)
		"神级翻唱：戴上耳机领略仙音！", // 3 (right)
		// 4 (left): 36 ASCII + wide runes. "    4. " prefix is 7 cells, so the
		// title truncates to 43 cells stopping before a wide rune.
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa中文哈哈",
		"许嵩『安泊猜想巡回演唱会』2026", // 5 (right)
	}
	items := make([]model.MenuItem, 0, len(titles))
	for _, title := range titles {
		items = append(items, model.MenuItem{Title: title})
	}
	return items
}

func (dualColumnTruncateMenu) HelpHints() []model.HelpHint { return nil }

// TestDualColumnTruncatedItemKeepsRightColumnAligned verifies that when a
// left-column menu item is truncated (stopping before a wide CJK/full-width
// rune), the item is still padded to its fixed column width so every
// right-column item starts at the same visual column.
func TestDualColumnTruncatedItemKeepsRightColumnAligned(t *testing.T) {
	const width, height = 130, 30

	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Theme.ShowTitle = true
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	opts := model.DefaultOptions()
	opts.EnableStartup = false
	opts.WhetherDisplayTitle = true
	opts.DualColumn = true
	opts.MainMenu = &dualColumnTruncateMenu{}
	opts.MainMenuTitle = &model.MenuItem{Title: "每日推荐歌单"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opts.TeaOptions = []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithInput(nil),
		tea.WithOutput(io.Discard),
	}
	app := model.NewApp(opts)
	_ = app.Run()
	_, _ = app.Update(tea.WindowSizeMsg{Width: width, Height: height})

	view := ansi.Strip(app.View().Content)
	rightMarkers := []string{"1. 华语必听", "3. 神级翻唱", "5. 许嵩"}
	starts := make([]int, 0, len(rightMarkers))
	for _, marker := range rightMarkers {
		found := false
		for _, line := range strings.Split(view, "\n") {
			if idx := strings.Index(line, marker); idx >= 0 {
				starts = append(starts, lipgloss.Width(line[:idx]))
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("right-column marker %q not found:\n%s", marker, view)
		}
	}
	for i := 1; i < len(starts); i++ {
		if starts[i] != starts[0] {
			t.Fatalf("right-column items misaligned: %q starts at %d, want %d (all: %v)\n%s",
				rightMarkers[i], starts[i], starts[0], starts, view)
		}
	}
}
