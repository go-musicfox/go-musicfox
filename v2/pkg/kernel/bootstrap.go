package kernel

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// BootstrapConfig 启动配置
type BootstrapConfig struct {
	ConfigPath   string        `json:"config_path"`
	LogLevel     string        `json:"log_level"`
	LogFormat    string        `json:"log_format"`
	StartTimeout time.Duration `json:"start_timeout"`
	StopTimeout  time.Duration `json:"stop_timeout"`
}

// DefaultBootstrapConfig 默认启动配置
func DefaultBootstrapConfig() *BootstrapConfig {
	return &BootstrapConfig{
		ConfigPath:   "config/kernel.yaml",
		LogLevel:     "info",
		LogFormat:    "text",
		StartTimeout: 30 * time.Second,
		StopTimeout:  10 * time.Second,
	}
}

// Bootstrap 微内核启动器
type Bootstrap struct {
	config *BootstrapConfig
	kernel Kernel
	logger *slog.Logger
}

// NewBootstrap 创建新的启动器
func NewBootstrap(config *BootstrapConfig) *Bootstrap {
	if config == nil {
		config = DefaultBootstrapConfig()
	}

	return &Bootstrap{
		config: config,
	}
}

// Initialize 初始化启动器
func (b *Bootstrap) Initialize() error {
	// 1. 初始化日志系统
	if err := b.initializeLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	b.logger.Info("Bootstrap initializing...")

	// 2. 创建微内核实例
	b.kernel = NewMicroKernel()

	// 3. 加载配置
	if err := b.loadConfiguration(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	b.logger.Info("Bootstrap initialized successfully")
	return nil
}

// Start 启动微内核
func (b *Bootstrap) Start(ctx context.Context) error {
	if b.kernel == nil {
		return fmt.Errorf("kernel not initialized")
	}

	b.logger.Info("Starting microkernel...")

	// 创建带超时的上下文
	startCtx, cancel := context.WithTimeout(ctx, b.config.StartTimeout)
	defer cancel()

	// 初始化内核
	if err := b.kernel.Initialize(startCtx); err != nil {
		return fmt.Errorf("failed to initialize kernel: %w", err)
	}

	// 启动内核
	if err := b.kernel.Start(startCtx); err != nil {
		return fmt.Errorf("failed to start kernel: %w", err)
	}

	b.logger.Info("Microkernel started successfully")
	return nil
}

// Stop 停止微内核
func (b *Bootstrap) Stop(ctx context.Context) error {
	if b.kernel == nil {
		return nil
	}

	b.logger.Info("Stopping microkernel...")

	// 创建带超时的上下文
	stopCtx, cancel := context.WithTimeout(ctx, b.config.StopTimeout)
	defer cancel()

	// 停止内核
	if err := b.kernel.Stop(stopCtx); err != nil {
		b.logger.Error("Failed to stop kernel gracefully", "error", err)
		return err
	}

	b.logger.Info("Microkernel stopped successfully")
	return nil
}

// Shutdown 关闭微内核
func (b *Bootstrap) Shutdown(ctx context.Context) error {
	if b.kernel == nil {
		return nil
	}

	b.logger.Info("Shutting down microkernel...")

	// 关闭内核
	if err := b.kernel.Shutdown(ctx); err != nil {
		b.logger.Error("Failed to shutdown kernel", "error", err)
		return err
	}

	b.logger.Info("Microkernel shutdown completed")
	return nil
}

// GetKernel 获取内核实例
func (b *Bootstrap) GetKernel() Kernel {
	return b.kernel
}

// GetLogger 获取日志器
func (b *Bootstrap) GetLogger() *slog.Logger {
	return b.logger
}

// initializeLogger 初始化日志系统
func (b *Bootstrap) initializeLogger() error {
	// 解析日志级别
	var level slog.Level
	switch b.config.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}



	// 创建处理器选项
	opts := &slog.HandlerOptions{
		Level: level,
		AddSource: true,
	}

	// 根据格式创建处理器
	var handler slog.Handler
	switch b.config.LogFormat {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		fallthrough
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// 创建日志器
	b.logger = slog.New(handler)

	return nil
}

// loadConfiguration 加载配置
func (b *Bootstrap) loadConfiguration() error {
	// 获取内核配置
	kernelConfig := b.kernel.GetConfig()

	// 加载配置文件（如果存在）
	if b.config.ConfigPath != "" {
		if _, err := os.Stat(b.config.ConfigPath); err == nil {
			if err := kernelConfig.Load(file.Provider(b.config.ConfigPath), yaml.Parser()); err != nil {
				b.logger.Warn("Failed to load config file", "path", b.config.ConfigPath, "error", err)
			}
		}
	}

	// 先设置默认配置
	b.setDefaultConfig(kernelConfig)

	// 然后加载环境变量（覆盖默认值）
	if err := kernelConfig.Load(env.Provider("MUSICFOX_", ".", func(s string) string {
		// 转换环境变量名：MUSICFOX_KERNEL_NAME -> kernel.name, MUSICFOX_KERNEL_LOG_LEVEL -> kernel.log_level
		key := strings.ToLower(strings.TrimPrefix(s, "MUSICFOX_"))
		// 只替换第一个下划线为点，保留其他下划线
		if idx := strings.Index(key, "_"); idx != -1 {
			key = key[:idx] + "." + key[idx+1:]
		}

		return key
	}), nil); err != nil {
		b.logger.Warn("Failed to load environment variables", "error", err)
	}


	b.logger.Info("Configuration loaded successfully")
	return nil
}

// setDefaultConfig 设置默认配置
func (b *Bootstrap) setDefaultConfig(config *koanf.Koanf) {
	// 内核默认配置
	defaults := map[string]interface{}{
		"kernel.version":     "1.0.0",
		"kernel.name":        "go-musicfox",
		"kernel.log_level":   "info", // 默认日志级别，可被环境变量覆盖
		"kernel.log_format":  b.config.LogFormat,
		"kernel.data_dir":    filepath.Join(os.TempDir(), "go-musicfox"),
		"kernel.plugin_dir":  "plugins",
		"kernel.config_dir":  "config",
		"plugins.enabled":    true,
		"plugins.auto_load": true,
		"plugins.scan_dirs": []string{"plugins"},
		"security.enabled":   true,
		"security.sandbox":   true,
		"registry.enabled":   true,
		"events.enabled":     true,
		"events.buffer_size": 1000,
	}

	// 设置默认值（只设置不存在的键）
	for key, value := range defaults {
		if !config.Exists(key) {
			config.Set(key, value)
		}
	}
}