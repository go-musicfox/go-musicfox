package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"

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
	menu := NewMainMenu(netease)
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

func TestMainMenuCheckUpdateActionReturnsNotificationCommand(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	menu := NewMainMenu(netease)
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

	page, cmd := menu.Action(app, updateIndex)
	if page != app.MustMain() || cmd == nil {
		t.Fatalf("check-update action returned page=%T cmd=%v, want current main page and notification command", page, cmd)
	}
	if submenu := menu.SubMenu(app, updateIndex); submenu != nil {
		t.Fatalf("check-update entry still opens submenu %T", submenu)
	}
}

func TestCheckUpdateNotificationMsgCoversAllResults(t *testing.T) {
	tests := []struct {
		name          string
		hasUpdate     bool
		latestVersion string
		wantLevel     model.NotificationLevel
		wantText      string
		wantActions   int
	}{
		{name: "update available", hasUpdate: true, latestVersion: "v9.9.9", wantLevel: model.NotificationInfo, wantText: "v9.9.9", wantActions: 1},
		{name: "already latest", latestVersion: types.AppVersion, wantLevel: model.NotificationSuccess, wantText: "已是最新版本"},
		{name: "check failed", wantLevel: model.NotificationError, wantText: "无法获取最新版本信息"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := checkUpdateNotificationMsg(tt.hasUpdate, tt.latestVersion)
			if msg.Spec.Level != tt.wantLevel {
				t.Fatalf("notification level = %v, want %v", msg.Spec.Level, tt.wantLevel)
			}
			if text := msg.Spec.Title + " " + msg.Spec.Message; !strings.Contains(text, tt.wantText) {
				t.Fatalf("notification text = %q, want it to contain %q", text, tt.wantText)
			}
			if len(msg.Spec.Actions) != tt.wantActions {
				t.Fatalf("notification actions = %d, want %d", len(msg.Spec.Actions), tt.wantActions)
			}
		})
	}
}

func TestCheckUpdateResultRendersDirectlyInTUI(t *testing.T) {
	app, _ := newFormPageTestApp(t)
	msg := checkUpdateNotificationMsg(false, types.AppVersion)
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

func TestUpdateNotificationOpensVersionRelease(t *testing.T) {
	content := newVersionNotifyContent("v9.9.9")
	wantURL := types.AppGithubUrl + "/releases/tag/v9.9.9"
	if content.Url != wantURL || content.ActionLabel != "前往 GitHub" {
		t.Fatalf("update notification = %#v, want URL %q with GitHub action", content, wantURL)
	}

	openedURL := ""
	spec := buildToastNotificationSpec(content, notify.ToastInfo, func(url string) error {
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
