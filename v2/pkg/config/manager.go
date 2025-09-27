package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

// ConfigManager 配置管理器接口
type ConfigManager interface {
	// 加载配置
	Load() error
	LoadFromFile(path string) error
	LoadFromEnv(prefix string) error
	LoadFromFlags(flags *pflag.FlagSet) error

	// 获取配置
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetPluginConfig(name string) (*PluginConfig, error)
	GetAllPluginConfigs() (map[string]*PluginConfig, error)

	// 设置配置
	Set(key string, value interface{}) error
	SetPluginConfig(name string, config *PluginConfig) error

	// 保存配置
	Save() error
	SaveToFile(path string) error

	// 合并配置
	Merge(other *koanf.Koanf) error

	// 验证配置
	Validate() error

	// 获取原始koanf实例
	GetKoanf() *koanf.Koanf
}

// Manager 配置管理器实现
type Manager struct {
	k          *koanf.Koanf
	configDir  string
	configFile string
}

// NewManager 创建新的配置管理器
func NewManager(configDir, configFile string) *Manager {
	m := &Manager{
		k:          koanf.New("."),
		configDir:  configDir,
		configFile: configFile,
	}

	// 设置默认配置
	_ = m.createDefaultConfig()

	return m
}

// Load 加载默认配置
func (m *Manager) Load() error {
	// 1. 加载默认配置文件
	if m.configFile != "" {
		if err := m.LoadFromFile(m.configFile); err != nil {
			// 如果配置文件不存在，创建默认配置
			if os.IsNotExist(err) {
				if err := m.createDefaultConfig(); err != nil {
					return fmt.Errorf("failed to create default config: %w", err)
				}
			} else {
				return fmt.Errorf("failed to load config file: %w", err)
			}
		}
	}

	// 2. 加载环境变量
	if err := m.LoadFromEnv("MUSICFOX_"); err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	return nil
}

// LoadFromFile 从文件加载配置
func (m *Manager) LoadFromFile(path string) error {
	if path == "" {
		return ErrConfigNotFound
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrConfigNotFound
	}

	// 根据文件扩展名选择解析器
	var parser koanf.Parser
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		parser = json.Parser()
	case ".yaml", ".yml":
		parser = yaml.Parser()
	case ".toml":
		parser = toml.Parser()
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedConfigSource, ext)
	}

	// 加载配置文件
	if err := m.k.Load(file.Provider(path), parser); err != nil {
		return fmt.Errorf("%w: %v", ErrConfigLoadFailed, err)
	}

	return nil
}

// LoadFromEnv 从环境变量加载配置
func (m *Manager) LoadFromEnv(prefix string) error {
	return m.k.Load(env.Provider(prefix, ".", func(s string) string {
		// 转换环境变量名为配置键
		// 例如: MUSICFOX_PLUGIN_NETEASE_ENABLED -> plugin.netease.enabled
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, prefix)), "_", ".")
	}), nil)
}

// LoadFromFlags 从命令行参数加载配置
func (m *Manager) LoadFromFlags(flags *pflag.FlagSet) error {
	return m.k.Load(posflag.Provider(flags, ".", m.k), nil)
}

// Get 获取配置值
func (m *Manager) Get(key string) interface{} {
	return m.k.Get(key)
}

// GetString 获取字符串配置
func (m *Manager) GetString(key string) string {
	return m.k.String(key)
}

// GetInt 获取整数配置
func (m *Manager) GetInt(key string) int {
	return m.k.Int(key)
}

// GetBool 获取布尔配置
func (m *Manager) GetBool(key string) bool {
	return m.k.Bool(key)
}

// GetPluginConfig 获取插件配置
func (m *Manager) GetPluginConfig(name string) (*PluginConfig, error) {
	key := fmt.Sprintf("plugins.%s", name)
	var config PluginConfig
	if err := m.k.Unmarshal(key, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin config for %s: %w", name, err)
	}

	// 设置默认值
	config.SetDefaults()

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid plugin config for %s: %w", name, err)
	}

	return &config, nil
}

// GetAllPluginConfigs 获取所有插件配置
func (m *Manager) GetAllPluginConfigs() (map[string]*PluginConfig, error) {
	configs := make(map[string]*PluginConfig)
	pluginKeys := m.k.MapKeys("plugins")

	for _, key := range pluginKeys {
		config, err := m.GetPluginConfig(key)
		if err != nil {
			return nil, err
		}
		configs[key] = config
	}

	return configs, nil
}

// Set 设置配置值
func (m *Manager) Set(key string, value interface{}) error {
	return m.k.Set(key, value)
}

// SetPluginConfig 设置插件配置
func (m *Manager) SetPluginConfig(name string, config *PluginConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid plugin config: %w", err)
	}

	key := fmt.Sprintf("plugins.%s", name)
	return m.k.Set(key, config)
}

// Save 保存配置到默认文件
func (m *Manager) Save() error {
	if m.configFile == "" {
		return ErrConfigNotFound
	}
	return m.SaveToFile(m.configFile)
}

// SaveToFile 保存配置到指定文件
func (m *Manager) SaveToFile(path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 根据文件扩展名选择格式
	ext := strings.ToLower(filepath.Ext(path))
	var data []byte
	var err error

	switch ext {
	case ".json":
		data, err = m.k.Marshal(json.Parser())
	case ".yaml", ".yml":
		data, err = m.k.Marshal(yaml.Parser())
	case ".toml":
		data, err = m.k.Marshal(toml.Parser())
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedConfigSource, ext)
	}

	if err != nil {
		return fmt.Errorf("%w: %v", ErrConfigParseFailed, err)
	}

	// 写入文件
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("%w: %v", ErrConfigSaveFailed, err)
	}

	return nil
}

// Merge 合并其他配置
func (m *Manager) Merge(other *koanf.Koanf) error {
	return m.k.Merge(other)
}

// Validate 验证所有配置
func (m *Manager) Validate() error {
	// 验证所有插件配置
	configs, err := m.GetAllPluginConfigs()
	if err != nil {
		return err
	}

	for name, config := range configs {
		if err := config.Validate(); err != nil {
			return fmt.Errorf("invalid config for plugin %s: %w", name, err)
		}
	}

	return nil
}

// GetKoanf 获取原始koanf实例
func (m *Manager) GetKoanf() *koanf.Koanf {
	return m.k
}

// GetConfigDir 获取配置目录
func (m *Manager) GetConfigDir() string {
	return m.configDir
}

// GetConfigFile 获取配置文件路径
func (m *Manager) GetConfigFile() string {
	return m.configFile
}

// createDefaultConfig 创建默认配置
func (m *Manager) createDefaultConfig() error {
	// 设置默认配置
	defaultConfig := map[string]interface{}{
		"app": map[string]interface{}{
			"name":    "go-musicfox",
			"version": "2.0.0",
			"debug":   false,
		},
		"server": map[string]interface{}{
			"port": 8080,
			"host": "localhost",
		},
		"plugins": map[string]interface{}{
			"netease": map[string]interface{}{
				"name":       "netease",
				"type":       "rpc",
				"path":       "plugins/netease",
				"enabled":    true,
				"auto_start": true,
				"priority":   80,
				"resources": map[string]interface{}{
					"max_memory":     104857600, // 100MB
					"max_cpu":        0.5,
					"max_goroutines": 100,
					"timeout":        "30s",
				},
				"security": map[string]interface{}{
					"sandbox": true,
				},
			},
		},
	}

	// 加载默认配置
	for key, value := range defaultConfig {
		if err := m.k.Set(key, value); err != nil {
			return err
		}
	}

	// 保存到文件
	if m.configFile != "" {
		return m.SaveToFile(m.configFile)
	}

	return nil
}
