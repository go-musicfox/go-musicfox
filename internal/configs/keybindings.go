package configs

// KeybindingsConfig 快捷键绑定配置
type KeybindingsConfig struct {
	// 使用默认键绑定
	UseDefaultKeyBindings bool `koanf:"useDefaultKeyBindings"`
	// 全局快捷键
	Global map[string]string `koanf:"global"`
	// 应用内快捷键
	App map[string][]string `koanf:"app"`
}
