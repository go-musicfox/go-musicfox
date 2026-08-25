package checkupdate

import (
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// init is the compile-time registration entry: it runs when the plugin package
// is linked into the binary (via the internal/plugins aggregator blank import,
// which cmd/musicfox.go pulls in). The factory's base parameter is written as
// ui.BaseMenu — the registry signature uses its alias baseMenu, and the two are
// interchangeable (BaseMenu is exported exactly for external factories).
func init() {
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
}
