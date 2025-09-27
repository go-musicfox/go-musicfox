package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/pflag"
)

// Version 应用版本
var Version = "2.0.0-mvp"

// BuildTime 构建时间
var BuildTime = "unknown"

// GitCommit Git提交哈希
var GitCommit = "unknown"

func main() {
	// 正常的主程序逻辑
	runMVP()
}

// MVPApp 应用实例
type MVPApp struct {
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *slog.Logger
	pluginManager *PluginManager
	tuiPlugin     Plugin
	running       bool
}

func runMVP() {
	// 解析命令行参数
	configPath := pflag.StringP("config", "c", getDefaultConfigPath(), "配置文件路径")
	version := pflag.BoolP("version", "v", false, "显示版本信息")
	help := pflag.BoolP("help", "h", false, "显示帮助信息")
	logLevel := pflag.StringP("log-level", "l", "info", "日志级别 (debug, info, warn, error)")
	pflag.Parse()

	// 显示版本信息
	if *version {
		fmt.Printf("MusicFox MVP v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// 显示帮助信息
	if *help {
		printUsage()
		os.Exit(0)
	}

	// 创建应用实例
	app, err := NewMVPApp(*configPath, *logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create application: %v\n", err)
		os.Exit(1)
	}

	// 设置信号处理
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nReceived interrupt signal, shutting down...")
		app.Shutdown()
		os.Exit(0)
	}()

	// 运行应用
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("MusicFox MVP shutdown complete")
}

// NewMVPApp 创建新的MVP应用实例
func NewMVPApp(configPath, logLevel string) (*MVPApp, error) {
	// 创建日志器
	logger := createLogger(logLevel)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建插件管理器
	pluginManager := NewPluginManager(logger)

	app := &MVPApp{
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		pluginManager: pluginManager,
	}

	return app, nil
}

// Run 运行应用
func (app *MVPApp) Run() error {
	app.logger.Info("Starting MusicFox MVP", "version", Version)

	// 初始化插件系统
	if err := app.initializePluginSystem(); err != nil {
		return fmt.Errorf("failed to initialize plugin system: %w", err)
	}

	app.running = true
	app.logger.Info("MusicFox MVP started successfully")
	app.logger.Info("Real plugin system is now running")
	app.logger.Info("All plugins are loaded and active")

	// 让TUI插件接管程序控制
	if app.tuiPlugin != nil {
		app.logger.Info("Transferring control to TUI plugin")
		return app.runWithTUI()
	}

	// 如果没有TUI插件，等待上下文取消
	app.logger.Info("Running in headless mode")
	<-app.ctx.Done()
	return nil
}

// initializePluginSystem 初始化插件系统
func (app *MVPApp) initializePluginSystem() error {
	app.logger.Info("Initializing real plugin system...")

	// 启动插件管理器
	if err := app.pluginManager.Start(); err != nil {
		return fmt.Errorf("failed to start plugin manager: %w", err)
	}

	// 加载插件
	if err := app.loadPlugins(); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// 启动插件
	if err := app.startPlugins(); err != nil {
		return fmt.Errorf("failed to start plugins: %w", err)
	}

	app.logger.Info("Real plugin system initialized successfully")
	return nil
}

// loadPlugins 加载插件
func (app *MVPApp) loadPlugins() error {
	app.logger.Info("Loading real plugins...")

	// 1. 加载音频插件
	if err := app.loadAudioPlugin(); err != nil {
		app.logger.Error("Failed to load audio plugin", "error", err)
		return err
	}

	// 2. 加载播放列表插件
	if err := app.loadPlaylistPlugin(); err != nil {
		app.logger.Error("Failed to load playlist plugin", "error", err)
		return err
	}

	// 3. 加载网易云插件
	if err := app.loadNeteasePlugin(); err != nil {
		app.logger.Error("Failed to load netease plugin", "error", err)
		return err
	}

	// 4. 加载TUI插件
	if err := app.loadTUIPlugin(); err != nil {
		app.logger.Error("Failed to load TUI plugin", "error", err)
		return err
	}

	app.logger.Info("All real plugins loaded successfully")
	return nil
}

// startPlugins 启动所有插件
func (app *MVPApp) startPlugins() error {
	app.logger.Info("Starting plugins in dependency order")

	// 按依赖顺序启动插件
	order := []string{"audio", "playlist", "netease", "tui"}
	for _, pluginID := range order {
		if plugin, err := app.pluginManager.GetPlugin(pluginID); err == nil && plugin != nil {
			app.logger.Info("Starting plugin", "id", pluginID)
			if err := plugin.Start(); err != nil {
				return fmt.Errorf("failed to start plugin %s: %w", pluginID, err)
			}
			app.logger.Info("Plugin started successfully", "id", pluginID)
			
			// 保存TUI插件引用
			if pluginID == "tui" {
				app.tuiPlugin = plugin
			}
		}
	}

	return nil
}

// Shutdown 关闭应用
func (app *MVPApp) Shutdown() {
	if !app.running {
		return
	}

	app.logger.Info("Shutting down MusicFox MVP")
	app.running = false

	// 停止插件
	app.stopPlugins()

	// 停止插件管理器
	if app.pluginManager != nil {
		if err := app.pluginManager.Stop(); err != nil {
			app.logger.Error("Failed to stop plugin manager", "error", err)
		}
	}

	// 取消上下文
	app.cancel()

	app.logger.Info("MusicFox MVP shutdown complete")
}

// runWithTUI 让TUI插件接管程序控制
func (app *MVPApp) runWithTUI() error {
	app.logger.Info("TUI plugin is now controlling the application")
	
	// 检查TUI插件是否为RealTUIPluginWrapper类型
	if wrapper, ok := app.tuiPlugin.(*RealTUIPluginWrapper); ok {
		// 等待TUI插件运行完成或上下文取消
		select {
		case <-wrapper.WaitForCompletion():
			app.logger.Info("TUI plugin finished, shutting down application")
			// TUI程序结束，主动关闭应用
			app.Shutdown()
			return nil
		case <-app.ctx.Done():
			app.logger.Info("Application context cancelled")
			return nil
		}
	} else {
		// 如果不是预期的TUI插件类型，等待上下文取消
		app.logger.Warn("TUI plugin is not RealTUIPluginWrapper, waiting for context cancellation")
		<-app.ctx.Done()
		return nil
	}
}

// stopPlugins 停止所有插件
func (app *MVPApp) stopPlugins() {
	app.logger.Info("Stopping all plugins")

	// 按相反顺序停止插件
	order := []string{"tui", "netease", "playlist", "audio"}
	for _, pluginID := range order {
		if plugin, err := app.pluginManager.GetPlugin(pluginID); err == nil && plugin != nil {
			app.logger.Info("Stopping plugin", "id", pluginID)
			if err := plugin.Stop(); err != nil {
				app.logger.Error("Failed to stop plugin", "id", pluginID, "error", err)
			} else {
				app.logger.Info("Plugin stopped successfully", "id", pluginID)
			}
		}
	}

	app.logger.Info("All plugins stopped")
}

// createLogger 创建日志器
func createLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
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

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

// getDefaultConfigPath 获取默认配置文件路径
func getDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(homeDir, ".musicfox", "config.yaml")
}

func printUsage() {
	fmt.Printf(`MusicFox MVP v%s - 基于真实插件系统的音乐播放器

`, Version)
	fmt.Println("Usage:")
	fmt.Println("  musicfox-mvp [flags]")
	fmt.Println("  musicfox-mvp debug-login        # 调试登录功能")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  - 真实的插件系统架构")
	fmt.Println("  - 完整的TUI用户界面")
	fmt.Println("  - 网易云音乐集成")
	fmt.Println("  - 音乐搜索和播放")
	fmt.Println("  - 用户登录和个人音乐库")
	fmt.Println("  - 播放列表管理")
	fmt.Println("  - 音频播放控制")
	fmt.Println()
	fmt.Println("Flags:")
	pflag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  musicfox-mvp                    # 启动完整插件系统")
	fmt.Println("  musicfox-mvp --config config.yaml  # 使用指定配置文件")
	fmt.Println("  musicfox-mvp --log-level debug  # 启用调试日志")
	fmt.Println("  musicfox-mvp debug-login        # 调试登录问题")
}
