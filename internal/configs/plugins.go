// Package configs holds the application configuration model: the koanf-loaded
// Config struct plus the per-section sub-configs and the TOML editing/upgrade
// helpers that keep user config files intact.
package configs

// PluginsConfig 是插件启停配置。启停只控制插件主菜单入口的可见性与启动钩子
// 的执行——被禁用插件的菜单 key 注册与 BuildMenu 跳转不受影响（其它插件 /
// 内置代码仍可按 key 跳入其菜单）。
type PluginsConfig struct {
	// Disabled 是禁用的插件 id 列表（插件 id 即插件目录名，如 "search"、
	// "checkupdate"）。空配置表示全部插件启用。
	Disabled []string `koanf:"disabled"`
}

// IsPluginDisabled 报告 id 是否在禁用列表中。
func (p PluginsConfig) IsPluginDisabled(id string) bool {
	for _, disabled := range p.Disabled {
		if disabled == id {
			return true
		}
	}
	return false
}
