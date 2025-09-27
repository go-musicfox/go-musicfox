package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config MVP应用配置
type Config struct {
	// 应用配置
	App AppConfig `json:"app"`
	// 音频配置
	Audio AudioConfig `json:"audio"`
	// TUI配置
	TUI TUIConfig `json:"tui"`
	// 网易云音乐配置
	Netease NeteaseConfig `json:"netease"`
	// 日志配置
	Logging LoggingConfig `json:"logging"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Debug   bool   `json:"debug"`
}

// AudioConfig 音频配置
type AudioConfig struct {
	Backend     string `json:"backend"`     // beep, portaudio, etc.
	Volume      int    `json:"volume"`      // 0-100
	BufferSize  int    `json:"buffer_size"`
	SampleRate  int    `json:"sample_rate"`
	Channels    int    `json:"channels"`
	BitDepth    int    `json:"bit_depth"`
}

// TUIConfig TUI配置
type TUIConfig struct {
	Theme           string `json:"theme"`
	ShowLyrics      bool   `json:"show_lyrics"`
	ShowSpectrum    bool   `json:"show_spectrum"`
	RefreshRate     int    `json:"refresh_rate"`
	KeyBindings     map[string]string `json:"key_bindings"`
	ProgressBarChar string `json:"progress_bar_char"`
}

// NeteaseConfig 网易云音乐配置
type NeteaseConfig struct {
	AutoLogin    bool   `json:"auto_login"`
	CacheDir     string `json:"cache_dir"`
	Quality      string `json:"quality"`      // low, medium, high, lossless
	CookieFile   string `json:"cookie_file"`
	UserAgent    string `json:"user_agent"`
	Timeout      int    `json:"timeout"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	File   string `json:"file"`   // 日志文件路径
	Format string `json:"format"` // text, json
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".musicfox")
	cacheDir := filepath.Join(configDir, "cache")
	logFile := filepath.Join(configDir, "musicfox-mvp.log")
	cookieFile := filepath.Join(configDir, "netease_cookies.json")

	return &Config{
		App: AppConfig{
			Name:    "MusicFox MVP",
			Version: "1.0.0",
			Debug:   false,
		},
		Audio: AudioConfig{
			Backend:    "beep",
			Volume:     70,
			BufferSize: 4096,
			SampleRate: 44100,
			Channels:   2,
			BitDepth:   16,
		},
		TUI: TUIConfig{
			Theme:           "default",
			ShowLyrics:      true,
			ShowSpectrum:    false,
			RefreshRate:     30,
			ProgressBarChar: "█",
			KeyBindings: map[string]string{
				"play_pause": "space",
				"next":       "n",
				"prev":       "p",
				"volume_up":  "+",
				"volume_down": "-",
				"quit":       "q",
				"search":     "/",
				"help":       "?",
			},
		},
		Netease: NeteaseConfig{
			AutoLogin:  false,
			CacheDir:   cacheDir,
			Quality:    "high",
			CookieFile: cookieFile,
			UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			Timeout:    30,
		},
		Logging: LoggingConfig{
			Level:  "info",
			File:   logFile,
			Format: "text",
		},
	}
}

// ConfigManager 配置管理器
type ConfigManager struct {
	configPath string
	config     *Config
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		config:     DefaultConfig(),
	}
}

// Load 加载配置文件
func (cm *ConfigManager) Load() error {
	// 如果配置文件不存在，使用默认配置并保存
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		return cm.Save()
	}

	// 读取配置文件
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	if err := json.Unmarshal(data, cm.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// Save 保存配置文件
func (cm *ConfigManager) Save() error {
	// 确保配置目录存在
	configDir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入配置文件
	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// UpdateConfig 更新配置
func (cm *ConfigManager) UpdateConfig(updater func(*Config)) error {
	updater(cm.config)
	return cm.Save()
}

// GetDefaultConfigPath 获取默认配置文件路径
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 如果无法获取用户目录，使用临时目录
		homeDir = os.TempDir()
	}

	configDir := filepath.Join(homeDir, ".musicfox")
	return filepath.Join(configDir, "config.json")
}

// EnsureDirectories 确保必要的目录存在
func (cm *ConfigManager) EnsureDirectories() error {
	config := cm.GetConfig()

	// 创建缓存目录
	if err := os.MkdirAll(config.Netease.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// 创建日志目录
	logDir := filepath.Dir(config.Logging.File)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 创建cookie文件目录
	cookieDir := filepath.Dir(config.Netease.CookieFile)
	if err := os.MkdirAll(cookieDir, 0755); err != nil {
		return fmt.Errorf("failed to create cookie directory: %w", err)
	}

	return nil
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"go_version":   runtime.Version(),
		"num_cpu":      runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
	}
}

// ValidateConfig 验证配置
func (cm *ConfigManager) ValidateConfig() error {
	config := cm.GetConfig()

	// 验证音频配置
	if config.Audio.Volume < 0 || config.Audio.Volume > 100 {
		return fmt.Errorf("invalid audio volume: %d (must be 0-100)", config.Audio.Volume)
	}

	if config.Audio.SampleRate <= 0 {
		return fmt.Errorf("invalid sample rate: %d", config.Audio.SampleRate)
	}

	if config.Audio.Channels <= 0 {
		return fmt.Errorf("invalid channels: %d", config.Audio.Channels)
	}

	// 验证TUI配置
	if config.TUI.RefreshRate <= 0 {
		return fmt.Errorf("invalid refresh rate: %d", config.TUI.RefreshRate)
	}

	// 验证网易云配置
	validQualities := []string{"low", "medium", "high", "lossless"}
	validQuality := false
	for _, q := range validQualities {
		if config.Netease.Quality == q {
			validQuality = true
			break
		}
	}
	if !validQuality {
		return fmt.Errorf("invalid quality: %s (must be one of: %v)", config.Netease.Quality, validQualities)
	}

	// 验证日志配置
	validLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, l := range validLevels {
		if config.Logging.Level == l {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level: %s (must be one of: %v)", config.Logging.Level, validLevels)
	}

	return nil
}