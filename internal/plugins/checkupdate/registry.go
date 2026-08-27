package checkupdate

import (
	"github.com/go-musicfox/go-musicfox/internal/framework"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// Plugin is the checkupdate business plugin (P5 cordis shape): its Start
// registers the menu provider, the main-menu entry and the startup hook into
// the frontend scope — the registration window moves from package init() to
// the frontend scope Start. init() only declares the constructor in the
// framework registry; the frontend scope mounts the enabled subset with
// AddWithEnabled (disabled = not started = nothing registered).
type Plugin struct {
	framework.NoopPlugin
}

// Start registers the plugin's contributions. It re-enters ui.WithPlugin so
// the attribution stamp (currentPluginID) records the registrations under
// "checkupdate" — the stamp is a set/unset guard that works identically at
// Start time (see ui/plugin_registry.go). The factory's base parameter is
// written as ui.BaseMenu — the registry signature uses its alias baseMenu,
// and the two are interchangeable (BaseMenu is exported exactly for external
// factories).
func (p *Plugin) Start(_ *framework.Context) error {
	ui.WithPlugin("checkupdate", "检查更新", func() {
		ui.RegisterMenu("check_update", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return &CheckUpdateMenu{BaseMenu: base}, nil
		})
		// 声明主菜单入口：NewMainMenu 经 After 锚点链归并复现插件化前的主菜单
		// 原始顺序（检查更新跟在帮助（内置）后，位于链尾）。
		ui.RegisterMainMenuItemAfter("check_update", "检查更新", "help", nil)
		// 注册启动自动检查为启动钩子（原 shell 级硬编码启动检查移入插件）。
		ui.RegisterStartupHook(startupCheck)
	})
	return nil
}

// init is the compile-time registration entry: it runs when the plugin package
// is linked into the binary (via the internal/plugins aggregator blank import,
// which cmd/musicfox.go pulls in) and only declares the plugin constructor —
// actual registrations happen in Start (frontend scope).
func init() {
	framework.RegisterPlugin("checkupdate", func() framework.Plugin { return &Plugin{} })
}
