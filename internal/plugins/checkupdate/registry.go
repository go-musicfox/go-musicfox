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
	ui.RegisterMenu("check_update", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return &CheckUpdateMenu{BaseMenu: base}, nil
	})
	// 声明主菜单入口：NewMainMenu 按 Order 归并排序复现插件化前的主菜单
	// 原始顺序（检查更新15，排在帮助14 之后）。
	ui.RegisterMainMenuItemWithOrder("check_update", "检查更新", 15, nil)
	// 注册启动自动检查为启动钩子（原 shell 级硬编码启动检查移入插件）。
	ui.RegisterStartupHook(startupCheck)
}
