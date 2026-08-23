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
	// 声明主菜单入口：NewMainMenu 在全部内置项之后追加「检查更新」并构建
	// 插件菜单（无参菜单，key 即上面的注册）。
	ui.RegisterMainMenuItem("check_update", "检查更新")
	// 注册启动自动检查为启动钩子（原 shell 级硬编码启动检查移入插件）。
	ui.RegisterStartupHook(startupCheck)
}
