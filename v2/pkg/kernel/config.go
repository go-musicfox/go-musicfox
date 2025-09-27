package kernel

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// ConfigManager 配置管理器接口
type ConfigManager interface {
	// 生命周期
	Initialize(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Shutdown(ctx context.Context) error

	// 配置操作
	Load(configPath string) error
	Reload() error
	Validate() error
	Get(key string) interface{}
	Set(key string, value interface{}) error
	Exists(key string) bool

	// 热更新
	EnableHotReload() error
	DisableHotReload() error
	IsHotReloadEnabled() bool

	// 事件监听
	OnConfigChanged(callback ConfigChangeCallback)
	RemoveConfigChangeCallback(callback ConfigChangeCallback)

	// 获取配置实例
	GetKoanf() *koanf.Koanf
}

// ConfigChangeCallback 配置变更回调函数
type ConfigChangeCallback func(key string, oldValue, newValue interface{}) error

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	Key      string      `json:"key"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Timestamp time.Time  `json:"timestamp"`
}

// KernelConfigManager 内核配置管理器实现
type KernelConfigManager struct {
	koanf    *koanf.Koanf
	logger   *slog.Logger
	configPath string

	// 热更新相关
	watcher   *fsnotify.Watcher
	hotReload bool

	// 回调函数
	callbacks []ConfigChangeCallback
	mutex     sync.RWMutex

	// 生命周期
	ctx    context.Context
	cancel context.CancelFunc
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager(logger *slog.Logger) ConfigManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &KernelConfigManager{
		koanf:     koanf.New("."),
		logger:    logger,
		callbacks: make([]ConfigChangeCallback, 0),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Initialize 初始化配置管理器
func (cm *KernelConfigManager) Initialize(ctx context.Context) error {
	cm.logger.Info("Initializing config manager...")

	// 先加载环境变量
	if err := cm.loadEnvironmentVariables(); err != nil {
		cm.logger.Warn("Failed to load environment variables", "error", err)
	}

	// 然后设置默认配置（只设置不存在的键）
	if err := cm.setDefaultsIfNotExists(); err != nil {
		return fmt.Errorf("failed to set default config: %w", err)
	}

	cm.logger.Info("Config manager initialized successfully")
	return nil
}

// Start 启动配置管理器
func (cm *KernelConfigManager) Start(ctx context.Context) error {
	cm.logger.Info("Starting config manager...")

	// 如果启用了热更新，启动文件监控
	if cm.hotReload && cm.configPath != "" {
		if err := cm.startFileWatcher(); err != nil {
			cm.logger.Warn("Failed to start file watcher", "error", err)
		}
	}

	cm.logger.Info("Config manager started successfully")
	return nil
}

// Stop 停止配置管理器
func (cm *KernelConfigManager) Stop(ctx context.Context) error {
	cm.logger.Info("Stopping config manager...")

	// 停止文件监控
	if cm.watcher != nil {
		if err := cm.watcher.Close(); err != nil {
			cm.logger.Warn("Failed to close file watcher", "error", err)
		}
		cm.watcher = nil
	}

	cm.logger.Info("Config manager stopped successfully")
	return nil
}

// Shutdown 关闭配置管理器
func (cm *KernelConfigManager) Shutdown(ctx context.Context) error {
	cm.logger.Info("Shutting down config manager...")

	// 停止所有操作
	if err := cm.Stop(ctx); err != nil {
		cm.logger.Warn("Failed to stop config manager during shutdown", "error", err)
	}

	// 取消上下文
	cm.cancel()

	cm.logger.Info("Config manager shutdown completed")
	return nil
}

// Load 加载配置文件
func (cm *KernelConfigManager) Load(configPath string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if configPath == "" {
		return fmt.Errorf("config path cannot be empty")
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", configPath)
	}

	// 根据文件扩展名选择解析器
	parser, err := cm.getParserByExtension(configPath)
	if err != nil {
		return fmt.Errorf("unsupported config file format: %w", err)
	}

	// 保存旧配置用于比较
	oldConfig := cm.koanf.All()

	// 加载配置文件
	if err := cm.koanf.Load(file.Provider(configPath), parser); err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	cm.configPath = configPath

	// 触发配置变更事件
	cm.notifyConfigChanges(oldConfig, cm.koanf.All())

	cm.logger.Info("Config file loaded successfully", "path", configPath)
	return nil
}

// Reload 重新加载配置
func (cm *KernelConfigManager) Reload() error {
	if cm.configPath == "" {
		return fmt.Errorf("no config file to reload")
	}

	cm.logger.Info("Reloading config file", "path", cm.configPath)
	return cm.Load(cm.configPath)
}

// Validate 验证配置
func (cm *KernelConfigManager) Validate() error {
	// 验证必需的配置项
	requiredKeys := []string{
		"kernel.name",
		"kernel.version",
	}

	for _, key := range requiredKeys {
		if !cm.koanf.Exists(key) {
			return fmt.Errorf("required config key missing: %s", key)
		}
	}

	// 验证日志级别
	if logLevel := cm.koanf.String("kernel.log_level"); logLevel != "" {
		validLevels := []string{"debug", "info", "warn", "error"}
		valid := false
		for _, level := range validLevels {
			if logLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log level: %s", logLevel)
		}
	}

	// 验证目录路径
	dirKeys := []string{"kernel.data_dir", "kernel.plugin_dir", "kernel.config_dir"}
	for _, key := range dirKeys {
		if path := cm.koanf.String(key); path != "" {
			if !filepath.IsAbs(path) {
				// 相对路径转换为绝对路径
				absPath, err := filepath.Abs(path)
				if err != nil {
					return fmt.Errorf("invalid path for %s: %w", key, err)
				}
				cm.koanf.Set(key, absPath)
			}
		}
	}

	cm.logger.Debug("Config validation completed successfully")
	return nil
}

// Get 获取配置值
func (cm *KernelConfigManager) Get(key string) interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.koanf.Get(key)
}

// Set 设置配置值
func (cm *KernelConfigManager) Set(key string, value interface{}) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	oldValue := cm.koanf.Get(key)
	cm.koanf.Set(key, value)

	// 触发配置变更事件
	go cm.notifyConfigChange(key, oldValue, value)

	return nil
}

// Exists 检查配置键是否存在
func (cm *KernelConfigManager) Exists(key string) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.koanf.Exists(key)
}

// EnableHotReload 启用热更新
func (cm *KernelConfigManager) EnableHotReload() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.hotReload = true

	// 如果配置文件已加载且文件监控未启动，启动监控
	if cm.configPath != "" && cm.watcher == nil {
		return cm.startFileWatcher()
	}

	return nil
}

// DisableHotReload 禁用热更新
func (cm *KernelConfigManager) DisableHotReload() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.hotReload = false

	// 停止文件监控
	if cm.watcher != nil {
		if err := cm.watcher.Close(); err != nil {
			return err
		}
		cm.watcher = nil
	}

	return nil
}

// IsHotReloadEnabled 检查是否启用热更新
func (cm *KernelConfigManager) IsHotReloadEnabled() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.hotReload
}

// OnConfigChanged 注册配置变更回调
func (cm *KernelConfigManager) OnConfigChanged(callback ConfigChangeCallback) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.callbacks = append(cm.callbacks, callback)
}

// RemoveConfigChangeCallback 移除配置变更回调
func (cm *KernelConfigManager) RemoveConfigChangeCallback(callback ConfigChangeCallback) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for i, cb := range cm.callbacks {
		if reflect.ValueOf(cb).Pointer() == reflect.ValueOf(callback).Pointer() {
			cm.callbacks = append(cm.callbacks[:i], cm.callbacks[i+1:]...)
			break
		}
	}
}

// GetKoanf 获取koanf实例
func (cm *KernelConfigManager) GetKoanf() *koanf.Koanf {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.koanf
}

// setDefaults 设置默认配置
func (cm *KernelConfigManager) setDefaults() error {
	defaults := map[string]interface{}{
		"kernel.name":         "go-musicfox",
		"kernel.version":      "2.0.0",
		"kernel.log_level":    "info",
		"kernel.log_format":   "text",
		"kernel.data_dir":     filepath.Join(os.TempDir(), "go-musicfox"),
		"kernel.plugin_dir":   "plugins",
		"kernel.config_dir":   "config",
		"plugins.enabled":     true,
		"plugins.auto_load":   true,
		"plugins.scan_dirs":   []string{"plugins"},
		"security.enabled":    true,
		"security.sandbox":    true,
		"registry.enabled":    true,
		"events.enabled":      true,
		"events.buffer_size":  1000,
	}

	for key, value := range defaults {
		cm.koanf.Set(key, value)
	}

	return nil
}

// setDefaultsIfNotExists 只为不存在的键设置默认配置
func (cm *KernelConfigManager) setDefaultsIfNotExists() error {
	defaults := map[string]interface{}{
		"kernel.name":         "go-musicfox",
		"kernel.version":      "2.0.0",
		"kernel.log_level":    "info",
		"kernel.log_format":   "text",
		"kernel.data_dir":     filepath.Join(os.TempDir(), "go-musicfox"),
		"kernel.plugin_dir":   "plugins",
		"kernel.config_dir":   "config",
		"plugins.enabled":     true,
		"plugins.auto_load":   true,
		"plugins.scan_dirs":   []string{"plugins"},
		"security.enabled":    true,
		"security.sandbox":    true,
		"registry.enabled":    true,
		"events.enabled":      true,
		"events.buffer_size":  1000,
	}

	for key, value := range defaults {
		if !cm.koanf.Exists(key) {
			cm.koanf.Set(key, value)
		}
	}

	return nil
}

// loadEnvironmentVariables 加载环境变量
func (cm *KernelConfigManager) loadEnvironmentVariables() error {
	err := cm.koanf.Load(env.Provider("MUSICFOX_", ".", func(s string) string {
		// 转换环境变量名：MUSICFOX_KERNEL_NAME -> kernel.name
		key := strings.ToLower(strings.TrimPrefix(s, "MUSICFOX_"))
		key = strings.ReplaceAll(key, "_", ".")
		cm.logger.Debug("Loading environment variable", "original", s, "transformed", key)
		return key
	}), nil)
	if err != nil {
		cm.logger.Error("Failed to load environment variables", "error", err)
	}
	return err
}

// getParserByExtension 根据文件扩展名获取解析器
func (cm *KernelConfigManager) getParserByExtension(configPath string) (koanf.Parser, error) {
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".yaml", ".yml":
		return yaml.Parser(), nil
	case ".json":
		return json.Parser(), nil
	case ".toml":
		return toml.Parser(), nil
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// startFileWatcher 启动文件监控
func (cm *KernelConfigManager) startFileWatcher() error {
	var err error
	cm.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// 监控配置文件
	if err := cm.watcher.Add(cm.configPath); err != nil {
		cm.watcher.Close()
		return fmt.Errorf("failed to watch config file: %w", err)
	}

	// 启动监控协程
	go cm.watchConfigFile()

	cm.logger.Info("File watcher started", "path", cm.configPath)
	return nil
}

// watchConfigFile 监控配置文件变化
func (cm *KernelConfigManager) watchConfigFile() {
	for {
		select {
		case event, ok := <-cm.watcher.Events:
			if !ok {
				return
			}

			// 只处理写入和创建事件
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				cm.logger.Info("Config file changed, reloading...", "file", event.Name)

				// 延迟一点时间，确保文件写入完成
				time.Sleep(100 * time.Millisecond)

				if err := cm.Reload(); err != nil {
					cm.logger.Error("Failed to reload config file", "error", err)
				}
			}

		case err, ok := <-cm.watcher.Errors:
			if !ok {
				return
			}
			cm.logger.Error("File watcher error", "error", err)

		case <-cm.ctx.Done():
			return
		}
	}
}

// notifyConfigChanges 通知配置变更
func (cm *KernelConfigManager) notifyConfigChanges(oldConfig, newConfig map[string]interface{}) {
	// 比较配置变化
	for key, newValue := range newConfig {
		oldValue, exists := oldConfig[key]
		if !exists || !reflect.DeepEqual(oldValue, newValue) {
			go cm.notifyConfigChange(key, oldValue, newValue)
		}
	}

	// 检查删除的配置项
	for key, oldValue := range oldConfig {
		if _, exists := newConfig[key]; !exists {
			go cm.notifyConfigChange(key, oldValue, nil)
		}
	}
}

// notifyConfigChange 通知单个配置变更
func (cm *KernelConfigManager) notifyConfigChange(key string, oldValue, newValue interface{}) {
	cm.mutex.RLock()
	callbacks := make([]ConfigChangeCallback, len(cm.callbacks))
	copy(callbacks, cm.callbacks)
	cm.mutex.RUnlock()

	// 调用所有回调函数
	for _, callback := range callbacks {
		if err := callback(key, oldValue, newValue); err != nil {
			cm.logger.Error("Config change callback error", "key", key, "error", err)
		}
	}
}