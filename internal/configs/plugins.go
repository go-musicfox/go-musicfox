// Package configs holds the application configuration model: the koanf-loaded
// Config struct plus the per-section sub-configs and the TOML editing/upgrade
// helpers that keep user config files intact.
package configs

// PluginsConfig 是插件启停与 WASM 插件目录配置。启停只控制插件主菜单入口的
// 可见性与启动钩子的执行——被禁用插件的菜单 key 注册与 BuildMenu 跳转不受
// 影响（其它插件 / 内置代码仍可按 key 跳入其菜单）。
type PluginsConfig struct {
	// Disabled 是禁用的插件 id 列表（插件 id 即插件目录名，如 "search"、
	// "checkupdate"）。空配置表示全部插件启用。
	Disabled []string `koanf:"disabled"`

	// WasmDir is the directory scanned for WASM plugins (each subdirectory is
	// one plugin: manifest.toml + a .wasm file). Empty means the default
	// <config-dir>/wasm-plugins.
	WasmDir string `koanf:"wasmDir"`
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

// IsPluginEnabled reports whether the plugin id is enabled under the current
// [plugins] config. An id never declared and a nil AppConfig (tests) are treated as enabled.
func IsPluginEnabled(id string) bool {
	cfg := AppConfig
	if cfg == nil {
		return true
	}
	return !cfg.Plugins.IsPluginDisabled(id)
}
