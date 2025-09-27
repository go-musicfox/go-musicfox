package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// MPVConfig MPV应用配置
type MPVConfig struct {
	App struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Debug   bool   `yaml:"debug"`
	} `yaml:"app"`

	Audio struct {
		Backend    string   `yaml:"backend"`     // 强制使用"mpv"
		MPVPath    string   `yaml:"mpv_path"`    // MPV可执行文件路径
		MPVArgs    []string `yaml:"mpv_args"`    // 额外的MPV参数
		BufferSize int      `yaml:"buffer_size"` // 缓冲区大小
		Volume     int      `yaml:"volume"`      // 默认音量
	} `yaml:"audio"`

	Playlist struct {
		AutoSave      bool   `yaml:"auto_save"`      // 自动保存播放列表
		DefaultFormat string `yaml:"default_format"` // 默认播放列表格式
		SaveDir       string `yaml:"save_dir"`       // 播放列表保存目录
	} `yaml:"playlist"`

	Netease struct {
		Enabled  bool   `yaml:"enabled"`   // 是否启用网易云插件
		CacheDir string `yaml:"cache_dir"` // 缓存目录
	} `yaml:"netease"`

	TUI struct {
		Enabled    bool   `yaml:"enabled"`     // 是否启用TUI插件
		Theme      string `yaml:"theme"`       // TUI主题
		AutoStart  bool   `yaml:"auto_start"`  // 是否自动启动TUI
		FullScreen bool   `yaml:"full_screen"` // 是否全屏模式
	} `yaml:"tui"`

	Logging struct {
		Level  string `yaml:"level"`  // 日志级别
		File   string `yaml:"file"`   // 日志文件路径
		MaxSize int   `yaml:"max_size"` // 日志文件最大大小(MB)
	} `yaml:"logging"`
}

// LoadMPVConfig 加载MPV配置
func LoadMPVConfig(configPath string) (*MPVConfig, error) {
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultMPVConfig()
		if err := SaveMPVConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML配置
	var config MPVConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证和修正配置
	if err := validateAndFixConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// SaveMPVConfig 保存MPV配置
func SaveMPVConfig(config *MPVConfig, configPath string) error {
	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化配置
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultMPVConfig 创建默认MPV配置
func DefaultMPVConfig() *MPVConfig {
	config := &MPVConfig{}

	// 应用配置
	config.App.Name = "musicfox-mpv"
	config.App.Version = "2.0.0"
	config.App.Debug = false

	// 音频配置 - 强制使用MPV
	config.Audio.Backend = "mpv"
	config.Audio.MPVPath = "mpv" // 使用PATH中的mpv
	config.Audio.MPVArgs = []string{
		"--no-video",        // 禁用视频输出
		"--quiet",           // 静默模式
		"--really-quiet",    // 真正的静默模式
		"--no-terminal",     // 禁用终端控制
		"--idle",            // 空闲模式
		"--force-window=no", // 不强制显示窗口
	}
	config.Audio.BufferSize = 4096
	config.Audio.Volume = 80

	// 播放列表配置
	config.Playlist.AutoSave = true
	config.Playlist.DefaultFormat = "m3u8"
	config.Playlist.SaveDir = "./playlists"

	// 网易云配置
	config.Netease.Enabled = true
	config.Netease.CacheDir = "./cache/netease"

	// TUI配置
	config.TUI.Enabled = true
	config.TUI.Theme = "default"
	config.TUI.AutoStart = false
	config.TUI.FullScreen = false

	// 日志配置
	config.Logging.Level = "info"
	config.Logging.File = "./logs/musicfox-mpv.log"
	config.Logging.MaxSize = 100

	return config
}

// validateAndFixConfig 验证和修正配置
func validateAndFixConfig(config *MPVConfig) error {
	// 强制音频后端为MPV
	if config.Audio.Backend != "mpv" {
		config.Audio.Backend = "mpv"
	}

	// 检查MPV路径
	if config.Audio.MPVPath == "" {
		config.Audio.MPVPath = "mpv"
	}

	// 验证音量范围
	if config.Audio.Volume < 0 || config.Audio.Volume > 100 {
		config.Audio.Volume = 80
	}

	// 验证缓冲区大小
	if config.Audio.BufferSize <= 0 {
		config.Audio.BufferSize = 4096
	}

	// 确保MPV参数包含基本选项
	if len(config.Audio.MPVArgs) == 0 {
		config.Audio.MPVArgs = []string{
			"--no-video",
			"--quiet",
			"--really-quiet",
			"--no-terminal",
			"--idle",
			"--force-window=no",
		}
	} else {
		// 确保包含关键参数
		ensureArg := func(args []string, arg string) []string {
			for _, a := range args {
				if a == arg {
					return args
				}
			}
			return append(args, arg)
		}

		config.Audio.MPVArgs = ensureArg(config.Audio.MPVArgs, "--no-video")
		config.Audio.MPVArgs = ensureArg(config.Audio.MPVArgs, "--no-terminal")
	}

	// 验证播放列表格式
	validFormats := []string{"m3u", "m3u8", "pls", "xspf"}
	validFormat := false
	for _, format := range validFormats {
		if config.Playlist.DefaultFormat == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		config.Playlist.DefaultFormat = "m3u8"
	}

	// 验证日志级别
	validLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLevels {
		if config.Logging.Level == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		config.Logging.Level = "info"
	}

	// 验证日志文件大小
	if config.Logging.MaxSize <= 0 {
		config.Logging.MaxSize = 100
	}

	return nil
}

// getDefaultConfigPath 获取默认配置文件路径
func getDefaultConfigPath() string {
	// 优先使用用户配置目录
	if configDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(configDir, "musicfox", "mpv-config.yaml")
	}

	// 回退到用户主目录
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".config", "musicfox", "mpv-config.yaml")
	}

	// 最后回退到当前目录
	return "./config/mpv-config.yaml"
}

// GetMPVCommand 获取完整的MPV命令参数
func (c *MPVConfig) GetMPVCommand() []string {
	cmd := []string{c.Audio.MPVPath}
	cmd = append(cmd, c.Audio.MPVArgs...)
	return cmd
}

// GetMPVPath 获取MPV可执行文件路径
func (c *MPVConfig) GetMPVPath() string {
	return c.Audio.MPVPath
}

// GetMPVArgs 获取MPV参数
func (c *MPVConfig) GetMPVArgs() []string {
	return c.Audio.MPVArgs
}

// IsDebugEnabled 检查是否启用调试模式
func (c *MPVConfig) IsDebugEnabled() bool {
	return c.App.Debug
}

// GetLogLevel 获取日志级别
func (c *MPVConfig) GetLogLevel() string {
	return c.Logging.Level
}