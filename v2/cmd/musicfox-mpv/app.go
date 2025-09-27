package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
	plugincore "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	"github.com/go-musicfox/go-musicfox/v2/plugins/audio"
	"github.com/go-musicfox/go-musicfox/v2/plugins/playlist"
	"github.com/go-musicfox/go-musicfox/v2/plugins/tui"
	"github.com/spf13/pflag"
)

// MPVApp MPV专用应用结构
type MPVApp struct {
	kernel         kernel.Kernel
	audioPlugin    *audio.AudioPlugin
	playlistPlugin *playlist.PlaylistPluginImpl
	neteasePlugin  plugincore.Plugin
	tuiPlugin      plugincore.Plugin
	eventBus       event.EventBus
	logger         *slog.Logger
	commandHandler *CommandHandler
	statusMonitor  *StatusMonitor
	config         *MPVConfig
}

// NewMPVApp 创建MPV应用实例
func NewMPVApp(configPath, logLevel string) (*MPVApp, error) {
	// 检查MPV是否可用
	if err := checkMPVAvailability(); err != nil {
		return nil, fmt.Errorf("MPV not available: %w", err)
	}

	// 创建微内核
	var k kernel.Kernel
	var err error

	if configPath != "" {
		k, err = kernel.NewMicroKernelWithConfig(configPath)
	} else {
		k = kernel.NewMicroKernel()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create kernel: %w", err)
	}

	// 获取日志器
	logger := k.GetLogger()

	// 加载MPV配置
	config, err := LoadMPVConfig(configPath)
	if err != nil {
		logger.Warn("Failed to load config, using defaults", "error", err)
		config = DefaultMPVConfig()
	}

	// 创建应用实例
	app := &MPVApp{
		kernel: k,
		logger: logger,
		config: config,
	}

	return app, nil
}

// Run 运行应用
func (app *MPVApp) Run(ctx context.Context) error {
	// 初始化内核
	if err := app.kernel.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize kernel: %w", err)
	}

	// 启动内核
	if err := app.kernel.Start(ctx); err != nil {
		return fmt.Errorf("failed to start kernel: %w", err)
	}

	// 获取事件总线
	app.eventBus = app.kernel.GetEventBus()

	// 初始化插件
	if err := app.initializePlugins(ctx); err != nil {
		return fmt.Errorf("failed to initialize plugins: %w", err)
	}

	// 创建命令处理器
	app.commandHandler = NewCommandHandler(app.audioPlugin, app.playlistPlugin, app.logger)

	// 创建状态监控器
	app.statusMonitor = NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)

	// 启动状态监控
	if err := app.statusMonitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start status monitor: %w", err)
	}

	app.logger.Info("MusicFox MPV started successfully", "backend", "mpv")

	// 处理命令行参数
	args := pflag.Args()
	if len(args) > 0 {
		return app.handleCommand(ctx, args)
	}

	// 如果没有命令参数，进入交互模式
	return app.runInteractiveMode(ctx)
}

// initializePlugins 初始化插件
func (app *MPVApp) initializePlugins(ctx context.Context) error {
	// 创建带超时的上下文
	initCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	app.logger.Info("Initializing plugins...")

	// 创建音频插件
	app.logger.Debug("Creating audio plugin")
	app.audioPlugin = audio.NewAudioPlugin(app.eventBus)

	// 创建播放列表插件
	app.logger.Debug("Creating playlist plugin")
	app.playlistPlugin = playlist.NewPlaylistPlugin()

	// 创建TUI插件
	app.logger.Debug("Creating TUI plugin")
	app.tuiPlugin = tui.NewTUIPlugin()

	// 创建Netease插件（临时实现）
	app.logger.Debug("Creating Netease plugin")
	app.neteasePlugin = NewNeteasePluginWrapper()

	// 注册插件到内核
	serviceRegistry := app.kernel.GetServiceRegistry()

	// 注册核心服务
	if err := serviceRegistry.Register(ctx, &kernel.ServiceInfo{
		ID:      "eventBus",
		Name:    "eventBus",
		Version: "1.0.0",
		Address: "localhost",
		Port:    8080,
	}); err != nil {
		return fmt.Errorf("failed to register eventBus service: %w", err)
	}

	if err := serviceRegistry.Register(ctx, &kernel.ServiceInfo{
		ID:      "logger",
		Name:    "logger",
		Version: "1.0.0",
		Address: "localhost",
		Port:    8081,
	}); err != nil {
		return fmt.Errorf("failed to register logger service: %w", err)
	}

	// 创建服务注册表适配器并注册服务
	serviceAdapter := &ServiceRegistryAdapter{
		registry: serviceRegistry,
		services: make(map[string]interface{}),
	}
	// 注册服务到适配器
	serviceAdapter.RegisterService("eventBus", app.eventBus)
	serviceAdapter.RegisterService("logger", app.logger)

	// 创建插件上下文，强制使用MPV配置
	audioConfig := map[string]interface{}{
		"backend":     "mpv",
		"mpv_path":    app.config.Audio.MPVPath,
		"mpv_args":    app.config.Audio.MPVArgs,
		"buffer_size": app.config.Audio.BufferSize,
		"volume":      app.config.Audio.Volume,
	}

	// 初始化插件
	pluginInits := []struct {
		name   string
		plugin interface{ Initialize(plugincore.PluginContext) error }
		config map[string]interface{}
	}{
		{"audio", app.audioPlugin, audioConfig},
		{"playlist", app.playlistPlugin, nil},
		{"netease", app.neteasePlugin, map[string]interface{}{
			"enabled":   app.config.Netease.Enabled,
			"cache_dir": app.config.Netease.CacheDir,
		}},
		{"tui", app.tuiPlugin, map[string]interface{}{
			"enabled":     app.config.TUI.Enabled,
			"theme":       app.config.TUI.Theme,
			"auto_start":  app.config.TUI.AutoStart,
			"full_screen": app.config.TUI.FullScreen,
		}},
	}

	for _, init := range pluginInits {
		app.logger.Debug("Initializing plugin", "name", init.name)
		pluginCtx := &PluginContextImpl{
			ctx:             initCtx,
			eventBus:        app.eventBus,
			serviceRegistry: serviceAdapter,
			logger:          app.logger,
			config:          init.config,
		}
		if err := init.plugin.Initialize(pluginCtx); err != nil {
			return fmt.Errorf("failed to initialize %s plugin: %w", init.name, err)
		}
	}

	// 按依赖顺序启动插件：audio -> playlist -> netease -> tui
	pluginStarts := []struct {
		name   string
		plugin interface{ Start() error }
	}{
		{"audio", app.audioPlugin},
		{"playlist", app.playlistPlugin},
		{"netease", app.neteasePlugin},
		{"tui", app.tuiPlugin},
	}

	for _, start := range pluginStarts {
		app.logger.Debug("Starting plugin", "name", start.name)
		if err := start.plugin.Start(); err != nil {
			return fmt.Errorf("failed to start %s plugin: %w", start.name, err)
		}
		
		// 特殊处理：音频插件启动后强制切换到MPV后端
		if start.name == "audio" {
			app.logger.Debug("Switching to MPV backend")
			if err := app.audioPlugin.SwitchBackend("mpv"); err != nil {
				return fmt.Errorf("failed to switch to MPV backend: %w", err)
			}
		}
	}

	app.logger.Info("All plugins initialized with MPV backend", "plugins", []string{"audio", "playlist", "netease", "tui"})
	return nil
}

// handleCommand 处理单个命令
func (app *MPVApp) handleCommand(ctx context.Context, args []string) error {
	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "play":
		return app.commandHandler.HandlePlay(ctx, commandArgs)
	case "pause":
		return app.commandHandler.HandlePause(ctx)
	case "resume":
		return app.commandHandler.HandleResume(ctx)
	case "stop":
		return app.commandHandler.HandleStop(ctx)
	case "next":
		return app.commandHandler.HandleNext(ctx)
	case "prev":
		return app.commandHandler.HandlePrev(ctx)
	case "volume":
		return app.commandHandler.HandleVolume(ctx, commandArgs)
	case "status":
		return app.commandHandler.HandleStatus(ctx)
	case "playlist":
		return app.commandHandler.HandlePlaylist(ctx, commandArgs)
	case "interactive":
		return app.runInteractiveMode(ctx)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// runInteractiveMode 运行交互模式
func (app *MPVApp) runInteractiveMode(ctx context.Context) error {
	app.logger.Info("Entering interactive mode")
	fmt.Println("MusicFox MPV - Interactive Mode")
	fmt.Println("Type 'help' for available commands, 'quit' to exit")
	fmt.Println("Backend: MPV Player")
	fmt.Println()

	// 显示初始状态
	if err := app.commandHandler.HandleStatus(ctx); err != nil {
		app.logger.Warn("Failed to get initial status", "error", err)
	}
	fmt.Println()

	// 创建输入channel避免阻塞
	inputChan := make(chan string, 1)
	go func() {
		defer close(inputChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Print("> ")
				var input string
				if _, err := fmt.Scanln(&input); err != nil {
					// 输入错误，可能是EOF或其他问题
					if err.Error() == "EOF" {
						return
					}
					continue
				}
				select {
				case inputChan <- input:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case input, ok := <-inputChan:
			if !ok {
				// 输入channel关闭
				return nil
			}

			input = strings.TrimSpace(input)
			if input == "" {
				continue
			}

			if input == "quit" || input == "exit" {
				fmt.Println("Goodbye!")
				return nil
			}

			if input == "help" {
				app.printInteractiveHelp()
				continue
			}

			// 解析命令
			args := strings.Fields(input)
			if len(args) == 0 {
				continue
			}

			// 执行命令，添加超时控制
			cmdCtx, cmdCancel := context.WithTimeout(ctx, 10*time.Second)
			if err := app.handleCommand(cmdCtx, args); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			cmdCancel()
			fmt.Println()
		}
	}
}

// printInteractiveHelp 打印交互模式帮助
func (app *MPVApp) printInteractiveHelp() {
	fmt.Println("Available Commands:")
	fmt.Println("  play <url>          播放指定歌曲")
	fmt.Println("  pause               暂停播放")
	fmt.Println("  resume              恢复播放")
	fmt.Println("  stop                停止播放")
	fmt.Println("  next                下一首")
	fmt.Println("  prev                上一首")
	fmt.Println("  volume <level>      设置音量 (0-100)")
	fmt.Println("  status              显示播放状态")
	fmt.Println("  playlist create <name>     创建播放列表")
	fmt.Println("  playlist list              列出播放列表")
	fmt.Println("  playlist show <id>         显示播放列表")
	fmt.Println("  playlist add <id> <url>    添加歌曲到播放列表")
	fmt.Println("  help                显示此帮助")
	fmt.Println("  quit/exit           退出程序")
	fmt.Println()
	fmt.Println("Note: Using MPV player backend")
}

// Shutdown 关闭应用
func (app *MPVApp) Shutdown(ctx context.Context) error {
	app.logger.Info("Shutting down MusicFox MPV...")

	// 创建带超时的上下文
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 停止状态监控
	if app.statusMonitor != nil {
		app.logger.Debug("Stopping status monitor")
		app.statusMonitor.Stop()
	}

	// 按相反顺序停止插件：tui -> netease -> playlist -> audio
	plugins := []struct {
		name   string
		plugin interface{ Stop() error }
	}{
		{"tui", app.tuiPlugin},
		{"netease", app.neteasePlugin},
		{"playlist", app.playlistPlugin},
		{"audio", app.audioPlugin},
	}

	for _, p := range plugins {
		if p.plugin != nil {
			app.logger.Debug("Stopping plugin", "name", p.name)
			if err := p.plugin.Stop(); err != nil {
				app.logger.Warn("Failed to stop plugin", "name", p.name, "error", err)
			}
		}
	}

	// 停止内核
	if app.kernel != nil {
		app.logger.Debug("Shutting down kernel")
		if err := app.kernel.Shutdown(shutdownCtx); err != nil {
			app.logger.Warn("Failed to shutdown kernel", "error", err)
			return fmt.Errorf("kernel shutdown failed: %w", err)
		}
	}

	app.logger.Info("MusicFox MPV shutdown complete")
	return nil
}

// checkMPVAvailability 检查MPV是否可用
func checkMPVAvailability() error {
	_, err := exec.LookPath("mpv")
	if err != nil {
		return fmt.Errorf("mpv not found in PATH, please install MPV player")
	}
	return nil
}

// 适配器实现 - 复用musicfox-core的实现

// PluginContextImpl 插件上下文实现
type PluginContextImpl struct {
	ctx             context.Context
	eventBus        event.EventBus
	serviceRegistry plugincore.ServiceRegistry
	logger          *slog.Logger
	config          map[string]interface{}
}

// GetContext 获取上下文
func (p *PluginContextImpl) GetContext() context.Context {
	return p.ctx
}

// GetEventBus 获取事件总线
func (p *PluginContextImpl) GetEventBus() plugincore.EventBus {
	return &EventBusAdapter{eventBus: p.eventBus}
}

// GetServiceRegistry 获取服务注册表
func (p *PluginContextImpl) GetServiceRegistry() plugincore.ServiceRegistry {
	return p.serviceRegistry
}

// GetLogger 获取日志器
func (p *PluginContextImpl) GetLogger() plugincore.Logger {
	return &LoggerAdapter{logger: p.logger}
}

// GetDataDir 获取数据目录
func (p *PluginContextImpl) GetDataDir() string {
	return "./data"
}

// GetIsolationGroup 获取隔离组
func (p *PluginContextImpl) GetIsolationGroup() *plugincore.IsolationGroup {
	return nil
}

// GetPluginConfig 获取插件配置
func (p *PluginContextImpl) GetPluginConfig() plugincore.PluginConfig {
	return &PluginConfigAdapter{config: p.config}
}

// GetResourceMonitor 获取资源监控器
func (p *PluginContextImpl) GetResourceMonitor() *plugincore.ResourceMonitor {
	return nil
}

// GetSecurityManager 获取安全管理器
func (p *PluginContextImpl) GetSecurityManager() *plugincore.SecurityManager {
	return nil
}

// GetTempDir 获取临时目录
func (p *PluginContextImpl) GetTempDir() string {
	return "./tmp"
}

// SendMessage 发送消息
func (p *PluginContextImpl) SendMessage(topic string, data interface{}) error {
	event := &event.BaseEvent{
		ID:        "message-" + time.Now().Format("20060102150405"),
		Type:      event.EventType(topic),
		Source:    "mpv",
		Data:      data,
		Timestamp: time.Now(),
	}
	return p.eventBus.Publish(context.Background(), event)
}

// Shutdown 关闭上下文
func (p *PluginContextImpl) Shutdown() error {
	return nil
}

// Subscribe 订阅事件
func (p *PluginContextImpl) Subscribe(topic string, handler plugincore.EventHandler) error {
	return nil
}

// Unsubscribe 取消订阅
func (p *PluginContextImpl) Unsubscribe(topic string, handler plugincore.EventHandler) error {
	return nil
}

// UpdateConfig 更新配置
func (p *PluginContextImpl) UpdateConfig(config plugincore.PluginConfig) error {
	return nil
}

// BroadcastMessage 广播消息
func (p *PluginContextImpl) BroadcastMessage(message interface{}) error {
	event := &event.BaseEvent{
		ID:        "broadcast-" + time.Now().Format("20060102150405"),
		Type:      "plugin.broadcast",
		Source:    "mpv",
		Data:      message,
		Timestamp: time.Now(),
	}
	return p.eventBus.Publish(context.Background(), event)
}

// GetContainer 获取容器（兼容接口）
func (p *PluginContextImpl) GetContainer() plugincore.ServiceRegistry {
	return p.serviceRegistry
}

// ServiceRegistryAdapter 服务注册表适配器
type ServiceRegistryAdapter struct {
	registry kernel.ServiceRegistry
	services map[string]interface{}
	mu       sync.RWMutex
}

// RegisterService 注册服务
func (s *ServiceRegistryAdapter) RegisterService(name string, service interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.services == nil {
		s.services = make(map[string]interface{})
	}
	s.services[name] = service
	return nil
}

// GetService 获取服务
func (s *ServiceRegistryAdapter) GetService(name string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.services == nil {
		return nil, fmt.Errorf("service not found: %s", name)
	}
	if service, ok := s.services[name]; ok {
		return service, nil
	}
	return nil, fmt.Errorf("service not found: %s", name)
}

// UnregisterService 注销服务
func (s *ServiceRegistryAdapter) UnregisterService(name string) error {
	return nil
}

// ListServices 列出所有服务
func (s *ServiceRegistryAdapter) ListServices() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.services == nil {
		return []string{}
	}
	services := make([]string, 0, len(s.services))
	for name := range s.services {
		services = append(services, name)
	}
	return services
}

// HasService 检查服务是否存在
func (s *ServiceRegistryAdapter) HasService(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.services == nil {
		return false
	}
	_, exists := s.services[name]
	return exists
}

// EventBusAdapter 事件总线适配器
type EventBusAdapter struct {
	eventBus event.EventBus
}

// Publish 发布事件
func (e *EventBusAdapter) Publish(eventType string, data interface{}) error {
	event := &event.BaseEvent{
		ID:        "event-" + time.Now().Format("20060102150405"),
		Type:      event.EventType(eventType),
		Source:    "mpv",
		Data:      data,
		Timestamp: time.Now(),
	}
	return e.eventBus.Publish(context.Background(), event)
}

// Subscribe 订阅事件
func (e *EventBusAdapter) Subscribe(eventType string, handler plugincore.EventHandler) error {
	return nil
}

// Unsubscribe 取消订阅
func (e *EventBusAdapter) Unsubscribe(eventType string, handler plugincore.EventHandler) error {
	return nil
}

// GetSubscriberCount 获取订阅者数量
func (e *EventBusAdapter) GetSubscriberCount(eventType string) int {
	return 0
}

// LoggerAdapter 日志适配器
type LoggerAdapter struct {
	logger *slog.Logger
}

// Debug 调试日志
func (l *LoggerAdapter) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Info 信息日志
func (l *LoggerAdapter) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Warn 警告日志
func (l *LoggerAdapter) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error 错误日志
func (l *LoggerAdapter) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// PluginConfigAdapter 插件配置适配器
type PluginConfigAdapter struct {
	config map[string]interface{}
}

// GetString 获取字符串配置
func (p *PluginConfigAdapter) GetString(key string) string {
	if val, ok := p.config[key].(string); ok {
		return val
	}
	return ""
}

// GetInt 获取整数配置
func (p *PluginConfigAdapter) GetInt(key string) int {
	if val, ok := p.config[key].(int); ok {
		return val
	}
	return 0
}

// GetBool 获取布尔配置
func (p *PluginConfigAdapter) GetBool(key string) bool {
	if val, ok := p.config[key].(bool); ok {
		return val
	}
	return false
}

// GetCustomConfig 获取自定义配置
func (p *PluginConfigAdapter) GetCustomConfig() map[string]interface{} {
	return p.config
}

// GetDependencies 获取依赖
func (p *PluginConfigAdapter) GetDependencies() []string {
	return []string{}
}

// GetEnabled 获取启用状态
func (p *PluginConfigAdapter) GetEnabled() bool {
	return true
}

// GetID 获取ID
func (p *PluginConfigAdapter) GetID() string {
	return "mpv-plugin"
}

// GetName 获取名称
func (p *PluginConfigAdapter) GetName() string {
	return "MPV Plugin"
}

// GetPriority 获取优先级
func (p *PluginConfigAdapter) GetPriority() plugincore.PluginPriority {
	return plugincore.PluginPriority(0)
}

// GetResourceLimits 获取资源限制
func (p *PluginConfigAdapter) GetResourceLimits() *plugincore.ResourceLimits {
	return nil
}

// GetSecurityConfig 获取安全配置
func (p *PluginConfigAdapter) GetSecurityConfig() *plugincore.SecurityConfig {
	return nil
}

// GetVersion 获取版本
func (p *PluginConfigAdapter) GetVersion() string {
	return "1.0.0"
}

// Validate 验证配置
func (p *PluginConfigAdapter) Validate() error {
	return nil
}

// NeteasePluginWrapper Netease插件包装器
type NeteasePluginWrapper struct {
	*plugincore.BasePlugin
	logger *slog.Logger
	isRunning bool
}

// NewNeteasePluginWrapper 创建Netease插件包装器
func NewNeteasePluginWrapper() plugincore.Plugin {
	info := &plugincore.PluginInfo{
		ID:          "netease-music",
		Name:        "Netease Music",
		Version:     "1.0.0",
		Description: "Netease Cloud Music source plugin wrapper",
		Author:      "go-musicfox team",
		Type:        "music-source",
	}

	return &NeteasePluginWrapper{
		BasePlugin: plugincore.NewBasePlugin(info),
	}
}

// Initialize 初始化插件
func (p *NeteasePluginWrapper) Initialize(ctx plugincore.PluginContext) error {
	if err := p.BasePlugin.Initialize(ctx); err != nil {
		return err
	}

	p.logger = ctx.GetLogger().(*LoggerAdapter).logger
	p.logger.Info("Netease plugin wrapper initialized")
	return nil
}

// Start 启动插件
func (p *NeteasePluginWrapper) Start() error {
	if err := p.BasePlugin.Start(); err != nil {
		return err
	}

	p.isRunning = true
	p.logger.Info("Netease plugin wrapper started")
	return nil
}

// Stop 停止插件
func (p *NeteasePluginWrapper) Stop() error {
	if err := p.BasePlugin.Stop(); err != nil {
		return err
	}

	p.isRunning = false
	p.logger.Info("Netease plugin wrapper stopped")
	return nil
}

// IsRunning 检查插件是否运行
func (p *NeteasePluginWrapper) IsRunning() bool {
	return p.isRunning
}