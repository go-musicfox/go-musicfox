package checkupdate

import (
	"strings"
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// TestCheckUpdateMenuBuildsViaRegistryFactory proves the plugin menu is
// constructible through the real registry factory: the factory closure in
// registry.go builds a *CheckUpdateMenu from a ui.BaseMenu base (the exported
// base degrades to zero values, mirroring the plugin-boundary compile proof in
// internal/ui/plugin_boundary_external_test.go).
func TestCheckUpdateMenuBuildsViaRegistryFactory(t *testing.T) {
	menu, err := ui.BuildMenu("check_update", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(check_update) error = %v", err)
	}
	checkMenu, ok := menu.(*CheckUpdateMenu)
	if !ok {
		t.Fatalf("BuildMenu(check_update) = %T, want *CheckUpdateMenu", menu)
	}
	if key := checkMenu.GetMenuKey(); key != "check_update" {
		t.Fatalf("GetMenuKey() = %q, want check_update", key)
	}
	if checkMenu.IsPlayable() {
		t.Fatal("check update menu must not be playable")
	}
	if checkMenu.IsLocatable() {
		t.Fatal("check update menu must not be locatable")
	}
	if views := checkMenu.MenuViews(); views == nil {
		t.Fatal("MenuViews() must be non-nil")
	}
	if m := checkMenu.SubMenu(nil, 0); m != nil {
		t.Fatalf("SubMenu() = %v, want nil", m)
	}
}

// TestCheckUpdateMenuBeforeEnterHookInstalled 验证插件菜单安装了进入钩子：
// 主菜单选中「检查更新」即触发检查并弹回主页面（单次 Enter 等价，不再需要
// main_menu 的 index-15 特判）。钩子内部检查在后台 goroutine 执行（走真实
// 网络，此处不调用只验证安装）。
func TestCheckUpdateMenuBeforeEnterHookInstalled(t *testing.T) {
	menu, err := ui.BuildMenu("check_update", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(check_update) error = %v", err)
	}
	checkMenu := menu.(*CheckUpdateMenu)
	if hook := checkMenu.BeforeEnterMenuHook(); hook == nil {
		t.Fatal("BeforeEnterMenuHook() returned nil, want the trigger-check hook")
	}
}

// TestStartupCheckConfigGated 验证启动钩子受 [startup] checkUpdate 配置门控：
// 关闭时直接返回，不触发网络检查。
func TestStartupCheckConfigGated(t *testing.T) {
	previous := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previous })

	configs.AppConfig.Startup.CheckUpdate = false
	startupCheck() // 门控短路，无网络调用，不 panic
}

// TestCheckUpdateNotificationMsgCoversAllResults 验证三种检查结果的
// NotificationSpec 构造（与提取前 internal/ui/foxful_integration_test.go 的
// 同名测试一致，逻辑随插件整体迁移至此）。
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

// TestNewVersionNotifyContent 验证「发现新版本」通知内容的跳转链接与动作按钮
// （与提取前 internal/ui/foxful_integration_test.go 的对应断言一致）。
func TestNewVersionNotifyContent(t *testing.T) {
	content := newVersionNotifyContent("v9.9.9")
	wantURL := types.AppGithubUrl + "/releases/tag/v9.9.9"
	if content.Url != wantURL || content.ActionLabel != "前往 GitHub" {
		t.Fatalf("update notification = %#v, want URL %q with GitHub action", content, wantURL)
	}
}
