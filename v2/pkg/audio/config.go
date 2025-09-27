package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// AudioConfig 音频系统配置
type AudioConfig struct {
	DefaultBackend string                    `json:"default_backend" yaml:"default_backend"`
	Backends       map[string]*BackendConfig `json:"backends" yaml:"backends"`
	GlobalSettings *GlobalAudioSettings      `json:"global_settings" yaml:"global_settings"`
	HotReload      bool                      `json:"hot_reload" yaml:"hot_reload"`
	ConfigPath     string                    `json:"-" yaml:"-"`
}

// GlobalAudioSettings 全局音频设置
type GlobalAudioSettings struct {
	DefaultVolume    float64 `json:"default_volume" yaml:"default_volume"`
	BufferSize       int     `json:"buffer_size" yaml:"buffer_size"`
	SampleRate       int     `json:"sample_rate" yaml:"sample_rate"`
	Channels         int     `json:"channels" yaml:"channels"`
	AutoSwitchBackend bool    `json:"auto_switch_backend" yaml:"auto_switch_backend"`
	RetryAttempts    int     `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay       string  `json:"retry_delay" yaml:"retry_delay"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	config        *AudioConfig
	configPath    string
	watcher       *fsnotify.Watcher
	callbacks     []ConfigChangeCallback
	mutex         sync.RWMutex
	running       bool
	shutdownCh    chan struct{}
}

// ConfigChangeCallback 配置变化回调函数
type ConfigChangeCallback func(*AudioConfig, *AudioConfig) error

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath string) (*ConfigManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}
	
	cm := &ConfigManager{
		configPath: configPath,
		watcher:    watcher,
		callbacks:  make([]ConfigChangeCallback, 0),
		shutdownCh: make(chan struct{}),
	}
	
	// 加载初始配置
	if err := cm.LoadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load initial config: %w", err)
	}
	
	return cm, nil
}

// LoadConfig 加载配置文件
func (cm *ConfigManager) LoadConfig() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		defaultConfig := cm.createDefaultConfig()
		if err := cm.saveConfigToFile(defaultConfig); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
		cm.config = defaultConfig
		return nil
	}
	
	// 读取配置文件
	data, err := ioutil.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	// 解析配置
	var config AudioConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	config.ConfigPath = cm.configPath
	cm.config = &config
	
	return nil
}

// SaveConfig 保存配置到文件
func (cm *ConfigManager) SaveConfig() error {
	cm.mutex.RLock()
	config := cm.config
	cm.mutex.RUnlock()
	
	if config == nil {
		return fmt.Errorf("no config to save")
	}
	
	return cm.saveConfigToFile(config)
}

// saveConfigToFile 保存配置到文件
func (cm *ConfigManager) saveConfigToFile(config *AudioConfig) error {
	// 确保目录存在
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// 写入文件
	if err := ioutil.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// createDefaultConfig 创建默认配置
func (cm *ConfigManager) createDefaultConfig() *AudioConfig {
	return &AudioConfig{
		DefaultBackend: "beep",
		Backends: map[string]*BackendConfig{
			"beep": {
				Name:          "beep",
				Enabled:       true,
				Priority:      5,
				BufferSize:    4096,
				SampleRate:    44100,
				Channels:      2,
				DefaultVolume: 0.8,
				Settings:      make(map[string]interface{}),
			},
			"mpv": {
				Name:          "mpv",
				Enabled:       true,
				Priority:      7,
				BufferSize:    8192,
				SampleRate:    44100,
				Channels:      2,
				DefaultVolume: 0.8,
				Settings: map[string]interface{}{
					"audio_driver": "auto",
					"cache":        true,
				},
			},
		},
		GlobalSettings: &GlobalAudioSettings{
			DefaultVolume:     0.8,
			BufferSize:        4096,
			SampleRate:        44100,
			Channels:          2,
			AutoSwitchBackend: true,
			RetryAttempts:     3,
			RetryDelay:        "1s",
		},
		HotReload:  true,
		ConfigPath: cm.configPath,
	}
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() *AudioConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	if cm.config == nil {
		return nil
	}
	
	// 返回配置副本
	configCopy := *cm.config
	return &configCopy
}

// UpdateConfig 更新配置
func (cm *ConfigManager) UpdateConfig(newConfig *AudioConfig) error {
	cm.mutex.Lock()
	oldConfig := cm.config
	cm.config = newConfig
	cm.config.ConfigPath = cm.configPath
	cm.mutex.Unlock()
	
	// 保存到文件
	if err := cm.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save updated config: %w", err)
	}
	
	// 通知配置变化
	cm.notifyConfigChange(oldConfig, newConfig)
	
	return nil
}

// UpdateBackendConfig 更新特定后端配置
func (cm *ConfigManager) UpdateBackendConfig(backendName string, config *BackendConfig) error {
	cm.mutex.Lock()
	if cm.config.Backends == nil {
		cm.config.Backends = make(map[string]*BackendConfig)
	}
	
	oldConfig := cm.config
	newConfig := *cm.config
	newConfig.Backends = make(map[string]*BackendConfig)
	for k, v := range cm.config.Backends {
		newConfig.Backends[k] = v
	}
	newConfig.Backends[backendName] = config
	
	cm.config = &newConfig
	cm.mutex.Unlock()
	
	// 保存到文件
	if err := cm.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save backend config: %w", err)
	}
	
	// 通知配置变化
	cm.notifyConfigChange(oldConfig, &newConfig)
	
	return nil
}

// SetDefaultBackend 设置默认后端
func (cm *ConfigManager) SetDefaultBackend(backendName string) error {
	cm.mutex.Lock()
	oldConfig := cm.config
	newConfig := *cm.config
	newConfig.DefaultBackend = backendName
	cm.config = &newConfig
	cm.mutex.Unlock()
	
	// 保存到文件
	if err := cm.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save default backend: %w", err)
	}
	
	// 通知配置变化
	cm.notifyConfigChange(oldConfig, &newConfig)
	
	return nil
}

// AddConfigChangeCallback 添加配置变化回调
func (cm *ConfigManager) AddConfigChangeCallback(callback ConfigChangeCallback) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.callbacks = append(cm.callbacks, callback)
}

// StartWatching 开始监听配置文件变化
func (cm *ConfigManager) StartWatching(ctx context.Context) error {
	cm.mutex.Lock()
	if cm.running {
		cm.mutex.Unlock()
		return fmt.Errorf("config watcher already running")
	}
	cm.running = true
	cm.mutex.Unlock()
	
	// 添加配置文件到监听列表
	if err := cm.watcher.Add(cm.configPath); err != nil {
		return fmt.Errorf("failed to watch config file: %w", err)
	}
	
	// 启动监听协程
	go cm.watchLoop(ctx)
	
	return nil
}

// StopWatching 停止监听配置文件变化
func (cm *ConfigManager) StopWatching() error {
	cm.mutex.Lock()
	if !cm.running {
		cm.mutex.Unlock()
		return nil
	}
	cm.running = false
	cm.mutex.Unlock()
	
	close(cm.shutdownCh)
	return cm.watcher.Close()
}

// watchLoop 监听循环
func (cm *ConfigManager) watchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.shutdownCh:
			return
		case event, ok := <-cm.watcher.Events:
			if !ok {
				return
			}
			
			// 只处理写入和创建事件
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				cm.handleConfigFileChange()
			}
			
		case err, ok := <-cm.watcher.Errors:
			if !ok {
				return
			}
			// 记录错误，但继续运行
			fmt.Printf("Config watcher error: %v\n", err)
		}
	}
}

// handleConfigFileChange 处理配置文件变化
func (cm *ConfigManager) handleConfigFileChange() {
	// 延迟一点时间，避免文件正在写入时读取
	time.Sleep(100 * time.Millisecond)
	
	cm.mutex.Lock()
	oldConfig := cm.config
	cm.mutex.Unlock()
	
	// 重新加载配置
	if err := cm.LoadConfig(); err != nil {
		fmt.Printf("Failed to reload config: %v\n", err)
		return
	}
	
	cm.mutex.RLock()
	newConfig := cm.config
	cm.mutex.RUnlock()
	
	// 通知配置变化
	cm.notifyConfigChange(oldConfig, newConfig)
}

// notifyConfigChange 通知配置变化
func (cm *ConfigManager) notifyConfigChange(oldConfig, newConfig *AudioConfig) {
	cm.mutex.RLock()
	callbacks := make([]ConfigChangeCallback, len(cm.callbacks))
	copy(callbacks, cm.callbacks)
	cm.mutex.RUnlock()
	
	// 异步通知所有回调
	for _, callback := range callbacks {
		go func(cb ConfigChangeCallback) {
			if err := cb(oldConfig, newConfig); err != nil {
				fmt.Printf("Config change callback error: %v\n", err)
			}
		}(callback)
	}
}

// GetBackendConfig 获取特定后端配置
func (cm *ConfigManager) GetBackendConfig(backendName string) (*BackendConfig, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	if cm.config == nil || cm.config.Backends == nil {
		return nil, fmt.Errorf("no config available")
	}
	
	config, exists := cm.config.Backends[backendName]
	if !exists {
		return nil, fmt.Errorf("backend '%s' not found in config", backendName)
	}
	
	// 返回配置副本
	configCopy := *config
	return &configCopy, nil
}

// GetDefaultBackend 获取默认后端名称
func (cm *ConfigManager) GetDefaultBackend() string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	if cm.config == nil {
		return "beep" // 默认值
	}
	
	return cm.config.DefaultBackend
}

// IsHotReloadEnabled 检查是否启用热重载
func (cm *ConfigManager) IsHotReloadEnabled() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	if cm.config == nil {
		return true // 默认启用
	}
	
	return cm.config.HotReload
}

// ValidateConfig 验证配置
func (cm *ConfigManager) ValidateConfig(config *AudioConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	if config.DefaultBackend == "" {
		return fmt.Errorf("default backend cannot be empty")
	}
	
	// 检查默认后端是否在后端列表中
	if config.Backends != nil {
		if _, exists := config.Backends[config.DefaultBackend]; !exists {
			return fmt.Errorf("default backend '%s' not found in backends list", config.DefaultBackend)
		}
	}
	
	// 验证全局设置
	if config.GlobalSettings != nil {
		if config.GlobalSettings.DefaultVolume < 0 || config.GlobalSettings.DefaultVolume > 1 {
			return fmt.Errorf("default volume must be between 0 and 1")
		}
		
		if config.GlobalSettings.BufferSize <= 0 {
			return fmt.Errorf("buffer size must be positive")
		}
		
		if config.GlobalSettings.SampleRate <= 0 {
			return fmt.Errorf("sample rate must be positive")
		}
		
		if config.GlobalSettings.Channels <= 0 {
			return fmt.Errorf("channels must be positive")
		}
	}
	
	// 验证后端配置
	for name, backendConfig := range config.Backends {
		if backendConfig.Name != name {
			return fmt.Errorf("backend name mismatch: key='%s', config.name='%s'", name, backendConfig.Name)
		}
		
		if backendConfig.DefaultVolume < 0 || backendConfig.DefaultVolume > 1 {
			return fmt.Errorf("backend '%s' default volume must be between 0 and 1", name)
		}
	}
	
	return nil
}