package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config 配置管理器接口
type Config interface {
	// 配置读取
	Get(key string) interface{}
	GetString(key string) (string, bool)
	GetInt(key string) (int, bool)
	GetBool(key string) (bool, bool)
	GetFloat64(key string) (float64, bool)
	GetDuration(key string) time.Duration
	GetStringSlice(key string) ([]string, bool)

	// 配置读取（带默认值）
	GetStringWithDefault(key string, defaultValue string) string
	GetIntWithDefault(key string, defaultValue int) int
	GetBoolWithDefault(key string, defaultValue bool) bool

	// 配置设置
	Set(key string, value interface{}) error
	SetDefault(key string, value interface{})
	Delete(key string)
	Clear()

	// 配置管理
	Load(configPath string) error
	Save(configPath string) error
	Reload() error
	Reset() error

	// 配置查询
	Has(key string) bool
	GetAll() map[string]interface{}
	GetKeys() []string
	GetAllKeys() []string

	// 配置监听
	Watch(key string, callback func(oldValue, newValue interface{}))
	Unwatch(key string)
}

// DefaultConfig 默认配置实现
type DefaultConfig struct {
	data      map[string]interface{}
	defaults  map[string]interface{}
	watchers  map[string][]func(oldValue, newValue interface{})
	configPath string
	mutex     sync.RWMutex
}

// NewConfig 创建新的配置管理器
func NewConfig() Config {
	return &DefaultConfig{
		data:     make(map[string]interface{}),
		defaults: make(map[string]interface{}),
		watchers: make(map[string][]func(oldValue, newValue interface{})),
	}
}

// NewDefaultConfig 创建带有默认配置的配置管理器
func NewDefaultConfig() Config {
	config := &DefaultConfig{
		data:     make(map[string]interface{}),
		defaults: make(map[string]interface{}),
		watchers: make(map[string][]func(oldValue, newValue interface{})),
	}
	
	// 设置默认配置值
	config.SetDefault("log.level", "info")
	config.SetDefault("log.file", "./logs/musicfox.log")
	config.SetDefault("server.port", ":8080")
	config.SetDefault("server.host", "localhost")
	config.SetDefault("plugins.directory", "./plugins")
	config.SetDefault("plugins.auto_load", true)
	
	return config
}

// Get 获取配置值
func (c *DefaultConfig) Get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if value, exists := c.data[key]; exists {
		return value
	}

	if defaultValue, exists := c.defaults[key]; exists {
		return defaultValue
	}

	return nil
}

// GetString 获取字符串配置值
func (c *DefaultConfig) GetString(key string) (string, bool) {
	value := c.Get(key)
	if value == nil {
		return "", false
	}

	if str, ok := value.(string); ok {
		return str, true
	}

	return "", false
}

// GetInt 获取整数配置值
func (c *DefaultConfig) GetInt(key string) (int, bool) {
	value := c.Get(key)
	if value == nil {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetBool 获取布尔配置值
func (c *DefaultConfig) GetBool(key string) (bool, bool) {
	value := c.Get(key)
	if value == nil {
		return false, false
	}

	if b, ok := value.(bool); ok {
		return b, true
	}

	return false, false
}

// GetFloat64 获取浮点数配置值
func (c *DefaultConfig) GetFloat64(key string) (float64, bool) {
	value := c.Get(key)
	if value == nil {
		return 0.0, false
	}

	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0.0, false
	}
}

// GetDuration 获取时间间隔配置值
func (c *DefaultConfig) GetDuration(key string) time.Duration {
	value := c.Get(key)
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case time.Duration:
		return v
	case string:
		if duration, err := time.ParseDuration(v); err == nil {
			return duration
		}
		return 0
	case int64:
		return time.Duration(v)
	default:
		return 0
	}
}

// GetStringSlice 获取字符串切片配置值
func (c *DefaultConfig) GetStringSlice(key string) ([]string, bool) {
	value := c.Get(key)
	if value == nil {
		return nil, false
	}

	if slice, ok := value.([]string); ok {
		return slice, true
	}

	if slice, ok := value.([]interface{}); ok {
		result := make([]string, len(slice))
		for i, v := range slice {
			result[i] = fmt.Sprintf("%v", v)
		}
		return result, true
	}

	return nil, false
}

// Set 设置配置值
func (c *DefaultConfig) Set(key string, value interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	oldValue := c.data[key]
	c.data[key] = value

	// 触发监听器
	if watchers, exists := c.watchers[key]; exists {
		for _, callback := range watchers {
			go callback(oldValue, value)
		}
	}

	return nil
}

// SetDefault 设置默认配置值
func (c *DefaultConfig) SetDefault(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.defaults[key] = value
}

// Load 从文件加载配置
func (c *DefaultConfig) Load(configPath string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.configPath = configPath

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析JSON配置
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// 扁平化嵌套的配置
	flattened := c.flattenConfig(config, "")
	for k, v := range flattened {
		c.data[k] = v
	}

	return nil
}

// Save 保存配置到文件
func (c *DefaultConfig) Save(configPath string) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if configPath == "" {
		configPath = c.configPath
	}

	if configPath == "" {
		return fmt.Errorf("no config path specified")
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Reload 重新加载配置
func (c *DefaultConfig) Reload() error {
	if c.configPath == "" {
		return fmt.Errorf("no file loaded")
	}
	return c.Load(c.configPath)
}

// Reset 重置配置
func (c *DefaultConfig) Reset() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]interface{})
	return nil
}

// Has 检查配置键是否存在
func (c *DefaultConfig) Has(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, exists := c.data[key]
	if !exists {
		_, exists = c.defaults[key]
	}

	return exists
}

// GetAll 获取所有配置
func (c *DefaultConfig) GetAll() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result := make(map[string]interface{})

	// 先复制默认值
	for k, v := range c.defaults {
		result[k] = v
	}

	// 再复制实际配置值（覆盖默认值）
	for k, v := range c.data {
		result[k] = v
	}

	return result
}

// GetKeys 获取所有配置键
func (c *DefaultConfig) GetKeys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make(map[string]bool)

	// 收集默认配置键
	for k := range c.defaults {
		keys[k] = true
	}

	// 收集实际配置键
	for k := range c.data {
		keys[k] = true
	}

	result := make([]string, 0, len(keys))
	for k := range keys {
		result = append(result, k)
	}

	return result
}

// Watch 监听配置变化
func (c *DefaultConfig) Watch(key string, callback func(oldValue, newValue interface{})) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.watchers[key] == nil {
		c.watchers[key] = make([]func(oldValue, newValue interface{}), 0)
	}

	c.watchers[key] = append(c.watchers[key], callback)
}

// Unwatch 取消监听配置变化
func (c *DefaultConfig) Unwatch(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.watchers, key)
}

// GetStringWithDefault 获取字符串配置值（带默认值）
func (c *DefaultConfig) GetStringWithDefault(key string, defaultValue string) string {
	if value, ok := c.GetString(key); ok {
		return value
	}
	return defaultValue
}

// GetIntWithDefault 获取整数配置值（带默认值）
func (c *DefaultConfig) GetIntWithDefault(key string, defaultValue int) int {
	if value, ok := c.GetInt(key); ok {
		return value
	}
	return defaultValue
}

// GetBoolWithDefault 获取布尔配置值（带默认值）
func (c *DefaultConfig) GetBoolWithDefault(key string, defaultValue bool) bool {
	if value, ok := c.GetBool(key); ok {
		return value
	}
	return defaultValue
}

// Delete 删除配置键
func (c *DefaultConfig) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

// GetAllKeys 获取所有配置键（别名方法）
func (c *DefaultConfig) GetAllKeys() []string {
	return c.GetKeys()
}

// Clear 清空所有配置
func (c *DefaultConfig) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]interface{})
	c.defaults = make(map[string]interface{})
}

// flattenConfig 扁平化嵌套的配置
func (c *DefaultConfig) flattenConfig(config map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range config {
		var fullKey string
		if prefix == "" {
			fullKey = key
		} else {
			fullKey = prefix + "." + key
		}

		if nested, ok := value.(map[string]interface{}); ok {
			// 递归处理嵌套对象
			for k, v := range c.flattenConfig(nested, fullKey) {
				result[k] = v
			}
		} else {
			result[fullKey] = value
		}
	}

	return result
}

// createDefaultConfig 创建默认配置文件
func (c *DefaultConfig) createDefaultConfig(configPath string) error {
	// 设置默认配置
	defaultConfig := map[string]interface{}{
		"kernel": map[string]interface{}{
			"name":         "go-musicfox",
			"version":      "2.0.0",
			"description":  "Go MusicFox Microkernel",
			"log_level":    "info",
			"plugin_dir":   "./plugins",
			"config_dir":   "./config",
			"data_dir":     "./data",
			"cache_dir":    "./cache",
			"temp_dir":     "./temp",
		},
		"security": map[string]interface{}{
			"enable_sandbox":     true,
			"max_memory_mb":      512,
			"max_cpu_percent":    50,
			"max_disk_mb":        1024,
			"max_network_mbps":   10,
			"allowed_operations": []string{"read", "write", "execute"},
		},
		"plugins": map[string]interface{}{
			"auto_load":        true,
			"load_timeout":     "30s",
			"start_timeout":    "10s",
			"stop_timeout":     "10s",
			"health_check":     true,
			"health_interval":  "30s",
		},
		"events": map[string]interface{}{
			"buffer_size":      1000,
			"worker_count":     4,
			"batch_size":       10,
			"flush_interval":   "1s",
		},
		"services": map[string]interface{}{
			"health_check_interval": "30s",
			"startup_timeout":       "60s",
			"shutdown_timeout":      "30s",
		},
	}

	c.data = defaultConfig

	// 保存默认配置到文件
	return c.Save(configPath)
}