package ui

import (
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/framework"
)

// withPluginConfig sets configs.AppConfig to a config with the given disabled
// plugin list and restores the previous value on cleanup (test isolation).
func withPluginConfig(t *testing.T, disabled []string) {
	t.Helper()
	previous := configs.AppConfig
	configs.AppConfig = &configs.Config{Plugins: configs.PluginsConfig{Disabled: disabled}}
	t.Cleanup(func() { configs.AppConfig = previous })
}

// pluginInfoSnapshot returns a snapshot copy of the declared plugin info with
// the given id, or nil.
func pluginInfoSnapshot(t *testing.T, id string) *PluginInfo {
	t.Helper()
	for _, info := range PluginInfos() {
		if info.ID == id {
			return &info
		}
	}
	return nil
}

func containsString(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// --- WithPlugin scope attribution ---

// TestWithPluginAttribution proves registrations made inside a WithPlugin scope
// are recorded under that plugin's PluginInfo (menu keys / page keys /
// main-menu item keys / startup-hook count), while the PluginInfos snapshot is
// a detached copy.
func TestWithPluginAttribution(t *testing.T) {
	const (
		pluginID = "plugin_registry_test_owner"
		menuKey  = "plugin_registry_test_menu"
		pageKey  = "plugin_registry_test_page"
	)
	WithPlugin(pluginID, "归属性测试", func() {
		RegisterMenu(menuKey, func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
			return &testCheckUpdateMenu{baseMenu: base}, nil
		})
		RegisterPage(pageKey, func(opts LoginPageOpts) (model.Page, error) {
			return &LoginPage{}, nil
		})
		// The main-menu entry key must be a registered menu provider (NewMainMenu
		// asserts it at construction), so reuse menuKey.
		RegisterMainMenuItem(menuKey, "归属测试项")
		RegisterStartupHook(func() {})
	})

	info := pluginInfoSnapshot(t, pluginID)
	if info == nil {
		t.Fatalf("plugin %q not declared", pluginID)
	}
	if info.ID != pluginID || info.Name != "归属性测试" {
		t.Fatalf("PluginInfo = %+v, want id %q name 归属性测试", info, pluginID)
	}
	if !containsString(info.MenuKeys, menuKey) {
		t.Fatalf("MenuKeys = %v, want to contain %q", info.MenuKeys, menuKey)
	}
	if !containsString(info.PageKeys, pageKey) {
		t.Fatalf("PageKeys = %v, want to contain %q", info.PageKeys, pageKey)
	}
	if !containsString(info.MainMenuItems, menuKey) {
		t.Fatalf("MainMenuItems = %v, want to contain %q", info.MainMenuItems, menuKey)
	}
	if info.StartupHooks != 1 {
		t.Fatalf("StartupHooks = %d, want 1", info.StartupHooks)
	}

	// The snapshot returned by PluginInfos must be detached: mutating it must
	// not corrupt the live registry (next call stays intact).
	snapshot := PluginInfos()
	for i := range snapshot {
		if snapshot[i].ID == pluginID {
			snapshot[i].MenuKeys = append(snapshot[i].MenuKeys, "polluted")
		}
	}
	info = pluginInfoSnapshot(t, pluginID)
	if containsString(info.MenuKeys, "polluted") {
		t.Fatal("PluginInfos() snapshot aliases the live registry")
	}
}

// TestWithPluginDuplicateIDMerges proves declaring the same plugin id twice
// (e.g. a plugin split across several init() files) merges idempotently onto
// the first PluginInfo instead of panicking: the first declaration's name is
// kept and later-scope registrations are appended to the same info.
func TestWithPluginDuplicateIDMerges(t *testing.T) {
	const pluginID = "plugin_registry_test_merge"
	WithPlugin(pluginID, "第一次声明", func() {
		RegisterMenu("plugin_registry_test_merge_a", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
			return &testCheckUpdateMenu{baseMenu: base}, nil
		})
	})
	// Same id again: must not panic.
	WithPlugin(pluginID, "第二次声明", func() {
		RegisterMenu("plugin_registry_test_merge_b", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
			return &testCheckUpdateMenu{baseMenu: base}, nil
		})
	})

	info := pluginInfoSnapshot(t, pluginID)
	if info == nil {
		t.Fatalf("plugin %q not declared", pluginID)
	}
	if info.Name != "第一次声明" {
		t.Fatalf("name = %q, want first declaration name 第一次声明", info.Name)
	}
	if !containsString(info.MenuKeys, "plugin_registry_test_merge_a") ||
		!containsString(info.MenuKeys, "plugin_registry_test_merge_b") {
		t.Fatalf("MenuKeys = %v, want both scopes' keys recorded", info.MenuKeys)
	}
}

func TestWithPluginValidation(t *testing.T) {
	assertPanics(t, func() { WithPlugin("", "空 id", func() {}) })
	assertPanics(t, func() { WithPlugin("plugin_registry_test_nil", "空函数", nil) })
}

// TestRegistrationsOutsideScopeNotAttributed proves registrations made outside
// any WithPlugin scope (empty attribution — built-in bootstrap, test binaries)
// are not recorded under any plugin.
func TestRegistrationsOutsideScopeNotAttributed(t *testing.T) {
	const key = "plugin_registry_test_unattributed"
	RegisterMenu(key, func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return &testCheckUpdateMenu{baseMenu: base}, nil
	})
	for _, info := range PluginInfos() {
		for _, k := range info.MenuKeys {
			if k == key {
				t.Fatalf("unattributed menu %q recorded under plugin %q", key, info.ID)
			}
		}
	}
}

// --- IsPluginEnabled ---

func TestIsPluginEnabledNilConfig(t *testing.T) {
	previous := configs.AppConfig
	configs.AppConfig = nil
	t.Cleanup(func() { configs.AppConfig = previous })

	for _, id := range []string{"search", "checkupdate", "never_declared"} {
		if !IsPluginEnabled(id) {
			t.Fatalf("IsPluginEnabled(%q) with nil AppConfig = false, want true", id)
		}
	}
}

func TestIsPluginEnabledReadsConfig(t *testing.T) {
	withPluginConfig(t, []string{"search", "checkupdate"})

	for _, id := range []string{"search", "checkupdate"} {
		if IsPluginEnabled(id) {
			t.Fatalf("IsPluginEnabled(%q) with disabled list = true, want false", id)
		}
	}
	for _, id := range []string{"lastfm", "dj", "never_declared"} {
		if !IsPluginEnabled(id) {
			t.Fatalf("IsPluginEnabled(%q) with disabled list = false, want true", id)
		}
	}
}

// --- consumption-point filtering ---

// TestNewMainMenuHidesDisabledPluginItem proves the main-menu consumption point
// filtering: with a plugin disabled in config, its main-menu item disappears
// from NewMainMenu Titles() while the rest of the chain still builds (no
// panic) and keeps its order.
func TestNewMainMenuHidesDisabledPluginItem(t *testing.T) {
	const pluginID = "ui_filter_owner"
	RegisterMenu("ui_filter_test_menu", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return &testCheckUpdateMenu{baseMenu: base}, nil
	})
	WithPlugin(pluginID, "过滤归属", func() {
		RegisterMainMenuItem("ui_filter_test_menu", "被禁用项")
	})

	// Enabled: the item is visible (appended at the end of the chain).
	enabled := NewMainMenu(testBase).Titles()
	if !containsString(enabled, "被禁用项") {
		t.Fatalf("enabled plugin item missing from main menu: %v", enabled)
	}

	// Disabled: the item disappears; the chain still builds without panicking
	// and the built-in 帮助 entry survives.
	withPluginConfig(t, []string{pluginID})
	disabled := NewMainMenu(testBase).Titles()
	if containsString(disabled, "被禁用项") {
		t.Fatalf("disabled plugin item still visible in main menu: %v", disabled)
	}
	if !containsString(disabled, "帮助") {
		t.Fatalf("disabled plugin filtering dropped unrelated entries: %v", disabled)
	}
}

// TestRunStartupHooksSkipsDisabledPlugin proves the startup-hook consumption
// point filtering: hooks registered inside a WithPlugin scope of a disabled
// plugin are skipped by the framework hook runner; hooks with empty attribution
// always run. Registration order is preserved for the rest.
func TestRunStartupHooksSkipsDisabledPlugin(t *testing.T) {
	const pluginID = "plugin_registry_test_hooks"
	withPluginConfig(t, []string{pluginID})

	var ran []string
	RegisterStartupHook(func() { ran = append(ran, "unattributed") })
	WithPlugin(pluginID, "钩子过滤", func() {
		RegisterStartupHook(func() { ran = append(ran, "disabled") })
	})
	// A non-disabled plugin-scoped hook must still run.
	WithPlugin("plugin_registry_test_hooks_enabled", "钩子启用", func() {
		RegisterStartupHook(func() { ran = append(ran, "enabled-scope") })
	})

	framework.RunStartupHooks(configs.IsPluginEnabled)

	want := []string{"unattributed", "enabled-scope"}
	if len(ran) != len(want) || ran[0] != want[0] || ran[1] != want[1] {
		t.Fatalf("startup hooks ran = %v, want %v (disabled plugin hook skipped)", ran, want)
	}
}
