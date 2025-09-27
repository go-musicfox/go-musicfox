package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
	plugincore "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	"github.com/go-musicfox/go-musicfox/v2/plugins/audio"
	"github.com/go-musicfox/go-musicfox/v2/plugins/playlist"
	"github.com/spf13/pflag"
)

// PluginContextImpl 插件上下文实现
type PluginContextImpl struct {
	ctx             context.Context
	eventBus        event.EventBus
	serviceRegistry plugincore.ServiceRegistry
	logger          *slog.Logger
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
	return &PluginConfigAdapter{config: make(map[string]interface{})}
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
		Source:    "core",
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
	// 简化实现
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
		Source:    "core",
		Data:      message,
		Timestamp: time.Now(),
	}
	return p.eventBus.Publish(context.Background(), event)
}

// GetContainer 获取容器（兼容接口）
func (p *PluginContextImpl) GetContainer() plugincore.ServiceRegistry {
	return p.serviceRegistry
}

// ServiceRegistryAdapter 适配器
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
		Source:    "core",
		Data:      data,
		Timestamp: time.Now(),
	}
	return e.eventBus.Publish(context.Background(), event)
}

// Subscribe 订阅事件
func (e *EventBusAdapter) Subscribe(eventType string, handler plugincore.EventHandler) error {
	// 简化实现
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
	return "core-plugin"
}

// GetName 获取名称
func (p *PluginConfigAdapter) GetName() string {
	return "Core Plugin"
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

// CoreApp 核心应用结构
type CoreApp struct {
	kernel        kernel.Kernel
	audioPlugin   *audio.AudioPlugin
	playlistPlugin *playlist.PlaylistPluginImpl
	eventBus      event.EventBus
	logger        *slog.Logger
	commandHandler *CommandHandler
	statusMonitor *StatusMonitor
}

// NewCoreApp 创建核心应用实例
func NewCoreApp(configPath, logLevel string) (*CoreApp, error) {
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

	// 创建应用实例
	app := &CoreApp{
		kernel: k,
		logger: logger,
	}

	return app, nil
}

// Run 运行应用
func (app *CoreApp) Run(ctx context.Context) error {
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

	app.logger.Info("MusicFox Core started successfully")

	// 处理命令行参数
	args := pflag.Args()
	if len(args) > 0 {
		return app.handleCommand(ctx, args)
	}

	// 如果没有命令参数，进入交互模式
	return app.runInteractiveMode(ctx)
}

// initializePlugins 初始化插件
func (app *CoreApp) initializePlugins(ctx context.Context) error {
	// 创建音频插件
	app.audioPlugin = audio.NewAudioPlugin(app.eventBus)

	// 创建播放列表插件
	app.playlistPlugin = playlist.NewPlaylistPlugin()

	// 注册插件到内核
	serviceRegistry := app.kernel.GetServiceRegistry()

	// 注册核心服务
	if err := serviceRegistry.Register(ctx, &kernel.ServiceInfo{
		ID:      "eventBus",
		Name:    "eventBus",
		Version: "1.0.0",
		Address: "localhost",
		Port:    8080,
	});	err != nil {
		return fmt.Errorf("failed to register eventBus service: %w", err)
	}

	if err := serviceRegistry.Register(ctx, &kernel.ServiceInfo{
		ID:      "logger",
		Name:    "logger",
		Version: "1.0.0",
		Address: "localhost",
		Port:    8081,
	});	err != nil {
		return fmt.Errorf("failed to register logger service: %w", err)
	}

	// 创建服务注册表适配器并注册服务
	serviceAdapter := &ServiceRegistryAdapter{
		registry: serviceRegistry,
		services: make(map[string]interface{}),
	}
	// 注册eventBus服务到适配器
	serviceAdapter.RegisterService("eventBus", app.eventBus)
	serviceAdapter.RegisterService("logger", app.logger)

	// 创建插件上下文
	audioCtx := &PluginContextImpl{
		ctx:             ctx,
		eventBus:        app.eventBus,
		serviceRegistry: serviceAdapter,
		logger:          app.logger,
	}
	if err := app.audioPlugin.Initialize(audioCtx); err != nil {
		return fmt.Errorf("failed to initialize audio plugin: %w", err)
	}

	playlistCtx := &PluginContextImpl{
		ctx:             ctx,
		eventBus:        app.eventBus,
		serviceRegistry: serviceAdapter,
		logger:          app.logger,
	}
	if err := app.playlistPlugin.Initialize(playlistCtx); err != nil {
		return fmt.Errorf("failed to initialize playlist plugin: %w", err)
	}

	// 启动插件
	if err := app.audioPlugin.Start(); err != nil {
		return fmt.Errorf("failed to start audio plugin: %w", err)
	}

	if err := app.playlistPlugin.Start(); err != nil {
		return fmt.Errorf("failed to start playlist plugin: %w", err)
	}

	app.logger.Info("All plugins initialized and started successfully")
	return nil
}

// handleCommand 处理单个命令
func (app *CoreApp) handleCommand(ctx context.Context, args []string) error {
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
func (app *CoreApp) runInteractiveMode(ctx context.Context) error {
	app.logger.Info("Entering interactive mode")
	fmt.Println("MusicFox Core - Interactive Mode")
	fmt.Println("Type 'help' for available commands, 'quit' to exit")
	fmt.Println()

	// 显示初始状态
	app.commandHandler.HandleStatus(ctx)
	fmt.Println()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Print("> ")
			var input string
			if _, err := fmt.Scanln(&input); err != nil {
				continue
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

			// 执行命令
			if err := app.handleCommand(ctx, args); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			fmt.Println()
		}
	}
}

// printInteractiveHelp 打印交互模式帮助
func (app *CoreApp) printInteractiveHelp() {
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
}

// Shutdown 关闭应用
func (app *CoreApp) Shutdown(ctx context.Context) error {
	app.logger.Info("Shutting down MusicFox Core...")

	// 停止状态监控
	if app.statusMonitor != nil {
		app.statusMonitor.Stop()
	}

	// 停止插件
	if app.audioPlugin != nil {
		app.audioPlugin.Stop()
	}

	if app.playlistPlugin != nil {
		app.playlistPlugin.Stop()
	}

	// 停止内核
	if app.kernel != nil {
		app.kernel.Shutdown(ctx)
	}

	app.logger.Info("MusicFox Core shutdown complete")
	return nil
}