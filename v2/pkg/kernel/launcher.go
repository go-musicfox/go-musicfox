package kernel

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LaunchMode 启动模式
type LaunchMode int

const (
	LaunchModeNormal LaunchMode = iota
	LaunchModeDaemon
	LaunchModeDebug
	LaunchModeTest
)

// String returns the string representation of LaunchMode
func (lm LaunchMode) String() string {
	switch lm {
	case LaunchModeNormal:
		return "normal"
	case LaunchModeDaemon:
		return "daemon"
	case LaunchModeDebug:
		return "debug"
	case LaunchModeTest:
		return "test"
	default:
		return "unknown"
	}
}

// LaunchConfig 启动配置
type LaunchConfig struct {
	// 基本配置
	Mode         LaunchMode    `json:"mode"`
	ConfigPath   string        `json:"config_path"`
	LogLevel     string        `json:"log_level"`
	LogFormat    string        `json:"log_format"`
	LogFile      string        `json:"log_file"`
	PidFile      string        `json:"pid_file"`
	WorkDir      string        `json:"work_dir"`

	// 超时配置
	StartTimeout    time.Duration `json:"start_timeout"`
	StopTimeout     time.Duration `json:"stop_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	// 功能开关
	EnableSignals   bool `json:"enable_signals"`
	EnableHotReload bool `json:"enable_hot_reload"`
	EnableMetrics   bool `json:"enable_metrics"`
	EnableProfile   bool `json:"enable_profile"`

	// 插件配置
	PluginDirs    []string `json:"plugin_dirs"`
	AutoLoadPlugins bool   `json:"auto_load_plugins"`

	// 调试配置
	DebugPort     int  `json:"debug_port"`
	ProfilePort   int  `json:"profile_port"`
	Verbose       bool `json:"verbose"`
}

// DefaultLaunchConfig 默认启动配置
func DefaultLaunchConfig() *LaunchConfig {
	return &LaunchConfig{
		Mode:            LaunchModeNormal,
		ConfigPath:      "config/kernel.yaml",
		LogLevel:        "info",
		LogFormat:       "text",
		LogFile:         "",
		PidFile:         "",
		WorkDir:         "",
		StartTimeout:    30 * time.Second,
		StopTimeout:     10 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		EnableSignals:   true,
		EnableHotReload: false,
		EnableMetrics:   false,
		EnableProfile:   false,
		PluginDirs:      []string{"plugins"},
		AutoLoadPlugins: true,
		DebugPort:       0,
		ProfilePort:     0,
		Verbose:         false,
	}
}

// Launcher 应用程序启动器
type Launcher struct {
	config    *LaunchConfig
	bootstrap *Bootstrap
	kernel    Kernel
	lifecycle LifecycleManager
	logger    *slog.Logger

	// 运行时信息
	pid       int
	startTime time.Time
	version   string
}

// NewLauncher 创建新的启动器
func NewLauncher() *Launcher {
	return &Launcher{
		config:  DefaultLaunchConfig(),
		pid:     os.Getpid(),
		version: "2.0.0", // TODO: 从构建信息获取
	}
}

// NewLauncherWithConfig 使用指定配置创建启动器
func NewLauncherWithConfig(config *LaunchConfig) *Launcher {
	if config == nil {
		config = DefaultLaunchConfig()
	}

	return &Launcher{
		config:  config,
		pid:     os.Getpid(),
		version: "2.0.0",
	}
}

// ParseFlags 解析命令行参数
func (l *Launcher) ParseFlags() error {
	// 定义命令行参数
	var (
		configPath   = flag.String("config", l.config.ConfigPath, "Configuration file path")
		logLevel     = flag.String("log-level", l.config.LogLevel, "Log level (debug, info, warn, error)")
		logFormat    = flag.String("log-format", l.config.LogFormat, "Log format (text, json)")
		logFile      = flag.String("log-file", l.config.LogFile, "Log file path (empty for stdout)")
		pidFile      = flag.String("pid-file", l.config.PidFile, "PID file path")
		workDir      = flag.String("work-dir", l.config.WorkDir, "Working directory")
		mode         = flag.String("mode", l.config.Mode.String(), "Launch mode (normal, daemon, debug, test)")
		verbose      = flag.Bool("verbose", l.config.Verbose, "Enable verbose logging")
		version      = flag.Bool("version", false, "Show version information")
		help         = flag.Bool("help", false, "Show help information")
		enableSignals = flag.Bool("enable-signals", l.config.EnableSignals, "Enable signal handling")
		enableHotReload = flag.Bool("enable-hot-reload", l.config.EnableHotReload, "Enable configuration hot reload")
		enableMetrics = flag.Bool("enable-metrics", l.config.EnableMetrics, "Enable metrics collection")
		enableProfile = flag.Bool("enable-profile", l.config.EnableProfile, "Enable profiling")
		debugPort    = flag.Int("debug-port", l.config.DebugPort, "Debug server port (0 to disable)")
		profilePort  = flag.Int("profile-port", l.config.ProfilePort, "Profile server port (0 to disable)")
	)

	// 自定义用法信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\ngo-musicfox v%s - A microkernel-based music player\n\n", l.version)
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --config config/production.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --mode daemon --pid-file /var/run/musicfox.pid\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --mode debug --log-level debug --verbose\n", os.Args[0])
	}

	// 解析参数
	flag.Parse()

	// 处理特殊参数
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *version {
		l.showVersion()
		os.Exit(0)
	}

	// 更新配置
	l.config.ConfigPath = *configPath
	l.config.LogLevel = *logLevel
	l.config.LogFormat = *logFormat
	l.config.LogFile = *logFile
	l.config.PidFile = *pidFile
	l.config.WorkDir = *workDir
	l.config.Verbose = *verbose
	l.config.EnableSignals = *enableSignals
	l.config.EnableHotReload = *enableHotReload
	l.config.EnableMetrics = *enableMetrics
	l.config.EnableProfile = *enableProfile
	l.config.DebugPort = *debugPort
	l.config.ProfilePort = *profilePort

	// 解析启动模式
	switch strings.ToLower(*mode) {
	case "normal":
		l.config.Mode = LaunchModeNormal
	case "daemon":
		l.config.Mode = LaunchModeDaemon
	case "debug":
		l.config.Mode = LaunchModeDebug
		l.config.LogLevel = "debug"
		l.config.Verbose = true
	case "test":
		l.config.Mode = LaunchModeTest
	default:
		return fmt.Errorf("invalid launch mode: %s", *mode)
	}

	// 验证配置
	return l.validateConfig()
}

// Initialize 初始化启动器
func (l *Launcher) Initialize() error {
	// 设置工作目录
	if l.config.WorkDir != "" {
		if err := os.Chdir(l.config.WorkDir); err != nil {
			return fmt.Errorf("failed to change working directory: %w", err)
		}
	}

	// 创建必要的目录
	if err := l.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// 写入PID文件
	if l.config.PidFile != "" {
		if err := l.writePidFile(); err != nil {
			return fmt.Errorf("failed to write PID file: %w", err)
		}
	}

	// 初始化日志
	if err := l.initializeLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	l.logger.Info("Launcher initialized",
		"mode", l.config.Mode.String(),
		"pid", l.pid,
		"version", l.version,
		"go_version", runtime.Version(),
		"arch", runtime.GOARCH,
		"os", runtime.GOOS)

	// 创建启动器组件
	bootstrapConfig := &BootstrapConfig{
		ConfigPath:   l.config.ConfigPath,
		LogLevel:     l.config.LogLevel,
		LogFormat:    l.config.LogFormat,
		StartTimeout: l.config.StartTimeout,
		StopTimeout:  l.config.StopTimeout,
	}

	l.bootstrap = NewBootstrap(bootstrapConfig)
	if err := l.bootstrap.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize bootstrap: %w", err)
	}

	// 获取内核实例
	l.kernel = l.bootstrap.GetKernel()

	// 创建生命周期管理器
	l.lifecycle = NewLifecycleManager(l.logger)

	// 配置生命周期阶段
	if err := l.setupLifecyclePhases(); err != nil {
		return fmt.Errorf("failed to setup lifecycle phases: %w", err)
	}

	// 启用信号处理
	if l.config.EnableSignals {
		if err := l.lifecycle.EnableSignalHandling(); err != nil {
			l.logger.Warn("Failed to enable signal handling", "error", err)
		}
	}

	return nil
}

// Run 运行应用程序
func (l *Launcher) Run() error {
	l.startTime = time.Now()
	l.logger.Info("Starting application...", "mode", l.config.Mode.String())

	ctx := context.Background()

	// 初始化生命周期管理器
	if err := l.lifecycle.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize lifecycle: %w", err)
	}

	// 启动生命周期管理器
	if err := l.lifecycle.Start(ctx); err != nil {
		return fmt.Errorf("failed to start lifecycle: %w", err)
	}

	l.logger.Info("Application started successfully",
		"uptime", time.Since(l.startTime),
		"kernel_state", l.kernel.GetStatus().State)

	// 根据模式运行
	switch l.config.Mode {
	case LaunchModeNormal:
		return l.runNormal()
	case LaunchModeDaemon:
		return l.runDaemon()
	case LaunchModeDebug:
		return l.runDebug()
	case LaunchModeTest:
		return l.runTest()
	default:
		return fmt.Errorf("unsupported launch mode: %s", l.config.Mode.String())
	}
}

// Shutdown 关闭应用程序
func (l *Launcher) Shutdown() error {
	l.logger.Info("Shutting down application...")

	ctx, cancel := context.WithTimeout(context.Background(), l.config.ShutdownTimeout)
	defer cancel()

	// 关闭生命周期管理器
	if l.lifecycle != nil {
		if err := l.lifecycle.Shutdown(ctx); err != nil {
			l.logger.Error("Failed to shutdown lifecycle", "error", err)
		}
	}

	// 清理PID文件
	if l.config.PidFile != "" {
		if err := os.Remove(l.config.PidFile); err != nil {
			l.logger.Warn("Failed to remove PID file", "error", err)
		}
	}

	uptime := time.Since(l.startTime)
	l.logger.Info("Application shutdown completed", "uptime", uptime)
	return nil
}

// GetConfig 获取启动配置
func (l *Launcher) GetConfig() *LaunchConfig {
	return l.config
}

// GetKernel 获取内核实例
func (l *Launcher) GetKernel() Kernel {
	return l.kernel
}

// GetLogger 获取日志器
func (l *Launcher) GetLogger() *slog.Logger {
	return l.logger
}

// validateConfig 验证配置
func (l *Launcher) validateConfig() error {
	// 验证日志级别
	validLogLevels := []string{"debug", "info", "warn", "error"}
	valid := false
	for _, level := range validLogLevels {
		if l.config.LogLevel == level {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log level: %s", l.config.LogLevel)
	}

	// 验证日志格式
	validLogFormats := []string{"text", "json"}
	valid = false
	for _, format := range validLogFormats {
		if l.config.LogFormat == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log format: %s", l.config.LogFormat)
	}

	// 验证超时配置
	if l.config.StartTimeout <= 0 {
		return fmt.Errorf("start timeout must be positive")
	}
	if l.config.StopTimeout <= 0 {
		return fmt.Errorf("stop timeout must be positive")
	}
	if l.config.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive")
	}

	// 验证端口配置
	if l.config.DebugPort < 0 || l.config.DebugPort > 65535 {
		return fmt.Errorf("invalid debug port: %d", l.config.DebugPort)
	}
	if l.config.ProfilePort < 0 || l.config.ProfilePort > 65535 {
		return fmt.Errorf("invalid profile port: %d", l.config.ProfilePort)
	}

	return nil
}

// createDirectories 创建必要的目录
func (l *Launcher) createDirectories() error {
	dirs := []string{
		filepath.Dir(l.config.ConfigPath),
	}

	if l.config.LogFile != "" {
		dirs = append(dirs, filepath.Dir(l.config.LogFile))
	}

	if l.config.PidFile != "" {
		dirs = append(dirs, filepath.Dir(l.config.PidFile))
	}

	for _, pluginDir := range l.config.PluginDirs {
		dirs = append(dirs, pluginDir)
	}

	for _, dir := range dirs {
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
	}

	return nil
}

// writePidFile 写入PID文件
func (l *Launcher) writePidFile() error {
	pidContent := fmt.Sprintf("%d\n", l.pid)
	return os.WriteFile(l.config.PidFile, []byte(pidContent), 0644)
}

// initializeLogger 初始化日志器
func (l *Launcher) initializeLogger() error {
	// 解析日志级别
	var level slog.Level
	switch l.config.LogLevel {
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
		Level:     level,
		AddSource: l.config.Verbose,
	}

	// 确定输出目标
	var output *os.File
	if l.config.LogFile != "" {
		var err error
		output, err = os.OpenFile(l.config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
	} else {
		output = os.Stdout
	}

	// 根据格式创建处理器
	var handler slog.Handler
	switch l.config.LogFormat {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text":
		fallthrough
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	// 创建日志器
	l.logger = slog.New(handler)

	return nil
}

// setupLifecyclePhases 设置生命周期阶段
func (l *Launcher) setupLifecyclePhases() error {
	// 初始化阶段
	initPhase := &LifecyclePhase{
		Name:        "kernel_init",
		Description: "Initialize kernel",
		Timeout:     l.config.StartTimeout,
		Handler: func(ctx context.Context) error {
			return l.kernel.Initialize(ctx)
		},
		Required: true,
		Order:    1,
	}

	// 启动阶段
	startPhase := &LifecyclePhase{
		Name:        "kernel_start",
		Description: "Start kernel",
		Timeout:     l.config.StartTimeout,
		Handler: func(ctx context.Context) error {
			return l.kernel.Start(ctx)
		},
		Required: true,
		Order:    1,
	}

	// 停止阶段
	stopPhase := &LifecyclePhase{
		Name:        "kernel_stop",
		Description: "Stop kernel",
		Timeout:     l.config.StopTimeout,
		Handler: func(ctx context.Context) error {
			return l.kernel.Stop(ctx)
		},
		Required: true,
		Order:    1,
	}

	// 关闭阶段
	shutdownPhase := &LifecyclePhase{
		Name:        "kernel_shutdown",
		Description: "Shutdown kernel",
		Timeout:     l.config.ShutdownTimeout,
		Handler: func(ctx context.Context) error {
			return l.kernel.Shutdown(ctx)
		},
		Required: true,
		Order:    1,
	}

	// 添加阶段
	if err := l.lifecycle.AddInitPhase(initPhase); err != nil {
		return err
	}
	if err := l.lifecycle.AddStartPhase(startPhase); err != nil {
		return err
	}
	if err := l.lifecycle.AddStopPhase(stopPhase); err != nil {
		return err
	}
	if err := l.lifecycle.AddShutdownPhase(shutdownPhase); err != nil {
		return err
	}

	return nil
}

// showVersion 显示版本信息
func (l *Launcher) showVersion() {
	fmt.Printf("go-musicfox v%s\n", l.version)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("Operating System: %s\n", runtime.GOOS)
	fmt.Printf("Build Date: %s\n", "unknown") // TODO: 从构建信息获取
}

// runNormal 正常模式运行
func (l *Launcher) runNormal() error {
	l.logger.Info("Running in normal mode")

	// 等待信号或生命周期结束
	for l.lifecycle.IsRunning() {
		time.Sleep(1 * time.Second)
	}

	return nil
}

// runDaemon 守护进程模式运行
func (l *Launcher) runDaemon() error {
	l.logger.Info("Running in daemon mode")

	// 守护进程逻辑
	// TODO: 实现守护进程化

	return l.runNormal()
}

// runDebug 调试模式运行
func (l *Launcher) runDebug() error {
	l.logger.Info("Running in debug mode")

	// 启动调试服务器
	if l.config.DebugPort > 0 {
		l.logger.Info("Debug server would start", "port", l.config.DebugPort)
		// TODO: 启动调试服务器
	}

	// 启动性能分析服务器
	if l.config.ProfilePort > 0 {
		l.logger.Info("Profile server would start", "port", l.config.ProfilePort)
		// TODO: 启动性能分析服务器
	}

	return l.runNormal()
}

// runTest 测试模式运行
func (l *Launcher) runTest() error {
	l.logger.Info("Running in test mode")

	// 测试模式逻辑
	// 运行一段时间后自动退出
	time.Sleep(5 * time.Second)

	l.logger.Info("Test mode completed")
	return nil
}