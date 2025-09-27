package kernel

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/knadh/koanf/v2"
	"go.uber.org/dig"
)

// Kernel 微内核接口
type Kernel interface {
	// 生命周期管理
	Initialize(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Shutdown(ctx context.Context) error

	// 组件管理
	GetPluginManager() PluginManager
	GetEventBus() event.EventBus
	GetServiceRegistry() ServiceRegistry
	GetSecurityManager() SecurityManager
	GetConfig() *koanf.Koanf
	GetLogger() *slog.Logger
	GetContainer() *dig.Container

	// 状态查询
	IsRunning() bool
	GetStatus() KernelStatus
}

// KernelStatus 内核状态
type KernelStatus struct {
	State       KernelState `json:"state"`
	StartedAt   int64       `json:"started_at"`
	Uptime      int64       `json:"uptime"`
	Version     string      `json:"version"`
	PluginCount int         `json:"plugin_count"`
}

// KernelState 内核状态枚举
type KernelState int

const (
	KernelStateUninitialized KernelState = iota
	KernelStateInitialized
	KernelStateStarting
	KernelStateRunning
	KernelStateStopping
	KernelStateStopped
	KernelStateError
)

// String returns the string representation of KernelState
func (ks KernelState) String() string {
	switch ks {
	case KernelStateUninitialized:
		return "uninitialized"
	case KernelStateInitialized:
		return "initialized"
	case KernelStateStarting:
		return "starting"
	case KernelStateRunning:
		return "running"
	case KernelStateStopping:
		return "stopping"
	case KernelStateStopped:
		return "stopped"
	case KernelStateError:
		return "error"
	default:
		return "unknown"
	}
}

// MicroKernel 微内核实现
type MicroKernel struct {
	// 核心组件
	eventBus        event.EventBus
	serviceRegistry ServiceRegistry
	securityManager SecurityManager
	pluginManager   PluginManager

	// 依赖注入容器
	container *dig.Container

	// 配置和日志
	config *koanf.Koanf
	logger *slog.Logger

	// 生命周期管理
	state     KernelState
	startedAt time.Time
	ctx       context.Context
	cancel    context.CancelFunc
	mutex     sync.RWMutex
}

// NewMicroKernel 创建新的微内核实例
func NewMicroKernel() Kernel {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建配置管理器
	cfg := koanf.New(".")

	// 设置默认日志器
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// 创建依赖注入容器
	container := dig.New()

	return &MicroKernel{
		container: container,
		config:    cfg,
		logger:    logger,
		state:     KernelStateUninitialized,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// NewMicroKernelWithConfig 使用指定配置创建微内核实例
func NewMicroKernelWithConfig(configPath string) (Kernel, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建配置管理器
	cfg := koanf.New(".")
	
	// 加载配置文件
	if configPath != "" {
		// 这里可以根据需要添加配置文件加载逻辑
		// 目前使用默认配置
	}

	// 根据配置设置日志级别
	logLevel := slog.LevelInfo
	if level := cfg.String("kernel.log_level"); level != "" {
		switch level {
		case "debug":
			logLevel = slog.LevelDebug
		case "info":
			logLevel = slog.LevelInfo
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// 创建依赖注入容器
	container := dig.New()

	return &MicroKernel{
		container: container,
		config:    cfg,
		logger:    logger,
		state:     KernelStateUninitialized,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Initialize 初始化微内核
func (mk *MicroKernel) Initialize(ctx context.Context) error {
	mk.mutex.Lock()
	defer mk.mutex.Unlock()

	if mk.state != KernelStateUninitialized {
		return fmt.Errorf("kernel already initialized")
	}

	mk.logger.Info("Initializing microkernel...")

	// 1. 初始化事件总线
	mk.eventBus = event.NewEventBus(mk.logger)
	if err := mk.eventBus.Start(ctx); err != nil {
		mk.state = KernelStateError
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// 2. 初始化服务注册表
	mk.serviceRegistry = NewServiceRegistry(mk.logger)
	if err := mk.serviceRegistry.Initialize(ctx); err != nil {
		mk.state = KernelStateError
		return fmt.Errorf("failed to initialize service registry: %w", err)
	}

	// 5. 初始化安全管理器
	// 创建一个空的koanf实例用于安全管理器
	securityConfig := koanf.New(".")
	mk.securityManager = NewSecurityManager(securityConfig, mk.logger)
	if err := mk.securityManager.Initialize(ctx); err != nil {
		mk.state = KernelStateError
		return fmt.Errorf("failed to initialize security manager: %w", err)
	}

	// 4. 初始化插件管理器
	mk.pluginManager = NewPluginManager(mk.logger, mk.eventBus, mk.securityManager, mk.serviceRegistry)
	if err := mk.pluginManager.Initialize(ctx); err != nil {
		mk.state = KernelStateError
		return fmt.Errorf("failed to initialize plugin manager: %w", err)
	}

	// 5. 设置依赖注入容器
	if err := mk.setupContainer(); err != nil {
		mk.state = KernelStateError
		return fmt.Errorf("failed to setup dependency injection container: %w", err)
	}

	mk.state = KernelStateInitialized
	mk.logger.Info("Microkernel initialized successfully")
	return nil
}

// Start 启动微内核
func (mk *MicroKernel) Start(ctx context.Context) error {
	mk.mutex.Lock()
	defer mk.mutex.Unlock()

	if mk.state != KernelStateInitialized {
		return fmt.Errorf("kernel not initialized")
	}

	mk.state = KernelStateStarting
	mk.logger.Info("Starting microkernel...")

	// 启动服务注册表
	if err := mk.serviceRegistry.Start(ctx); err != nil {
		mk.state = KernelStateError
		return fmt.Errorf("failed to start service registry: %w", err)
	}

	// 启动安全管理器
	if err := mk.securityManager.Start(ctx); err != nil {
		mk.state = KernelStateError
		return fmt.Errorf("failed to start security manager: %w", err)
	}

	// 启动插件管理器
	if err := mk.pluginManager.Start(ctx); err != nil {
		mk.state = KernelStateError
		return fmt.Errorf("failed to start plugin manager: %w", err)
	}

	mk.state = KernelStateRunning
	mk.startedAt = time.Now()
	mk.logger.Info("Microkernel started successfully")
	return nil
}

// Stop 停止微内核
func (mk *MicroKernel) Stop(ctx context.Context) error {
	mk.mutex.Lock()
	defer mk.mutex.Unlock()

	if mk.state != KernelStateRunning {
		return fmt.Errorf("kernel not running")
	}

	mk.state = KernelStateStopping
	mk.logger.Info("Stopping microkernel...")

	// 按相反顺序停止组件
	if mk.pluginManager != nil {
		if err := mk.pluginManager.Stop(ctx); err != nil {
			mk.logger.Error("Failed to stop plugin manager", "error", err)
		}
	}

	if mk.securityManager != nil {
		if err := mk.securityManager.Stop(ctx); err != nil {
			mk.logger.Error("Failed to stop security manager", "error", err)
		}
	}

	if mk.serviceRegistry != nil {
		if err := mk.serviceRegistry.Stop(ctx); err != nil {
			mk.logger.Error("Failed to stop service registry", "error", err)
		}
	}

	if mk.eventBus != nil {
		if err := mk.eventBus.Stop(ctx); err != nil {
			mk.logger.Error("Failed to stop event bus", "error", err)
		}
	}

	mk.state = KernelStateStopped
	mk.logger.Info("Microkernel stopped successfully")
	return nil
}

// Shutdown 关闭微内核
func (mk *MicroKernel) Shutdown(ctx context.Context) error {
	mk.mutex.Lock()
	defer mk.mutex.Unlock()

	mk.logger.Info("Shutting down microkernel...")

	// 如果还在运行，先停止
	if mk.state == KernelStateRunning {
		mk.mutex.Unlock()
		if err := mk.Stop(ctx); err != nil {
			mk.logger.Error("Failed to stop kernel during shutdown", "error", err)
		}
		mk.mutex.Lock()
	}

	// 关闭组件
	if mk.pluginManager != nil {
		if err := mk.pluginManager.Shutdown(ctx); err != nil {
			mk.logger.Error("Failed to shutdown plugin manager", "error", err)
		}
	}

	if mk.securityManager != nil {
		if err := mk.securityManager.Shutdown(ctx); err != nil {
			mk.logger.Error("Failed to shutdown security manager", "error", err)
		}
	}

	if mk.serviceRegistry != nil {
		if err := mk.serviceRegistry.Shutdown(ctx); err != nil {
			mk.logger.Error("Failed to shutdown service registry", "error", err)
		}
	}

	// 取消上下文
	mk.cancel()

	mk.logger.Info("Microkernel shutdown completed")
	return nil
}

// GetPluginManager 获取插件管理器
func (mk *MicroKernel) GetPluginManager() PluginManager {
	mk.mutex.RLock()
	defer mk.mutex.RUnlock()
	return mk.pluginManager
}

// GetEventBus 获取事件总线
func (mk *MicroKernel) GetEventBus() event.EventBus {
	mk.mutex.RLock()
	defer mk.mutex.RUnlock()
	return mk.eventBus
}

// GetServiceRegistry 获取服务注册表
func (mk *MicroKernel) GetServiceRegistry() ServiceRegistry {
	mk.mutex.RLock()
	defer mk.mutex.RUnlock()
	return mk.serviceRegistry
}

// GetSecurityManager 获取安全管理器
func (mk *MicroKernel) GetSecurityManager() SecurityManager {
	mk.mutex.RLock()
	defer mk.mutex.RUnlock()
	return mk.securityManager
}

// GetConfig 获取配置
func (mk *MicroKernel) GetConfig() *koanf.Koanf {
	mk.mutex.RLock()
	defer mk.mutex.RUnlock()
	return mk.config
}

// GetLogger 获取日志器
func (mk *MicroKernel) GetLogger() *slog.Logger {
	mk.mutex.RLock()
	defer mk.mutex.RUnlock()
	return mk.logger
}

// IsRunning 检查内核是否正在运行
func (mk *MicroKernel) IsRunning() bool {
	mk.mutex.RLock()
	defer mk.mutex.RUnlock()
	return mk.state == KernelStateRunning
}

// GetStatus 获取内核状态
func (mk *MicroKernel) GetStatus() KernelStatus {
	mk.mutex.RLock()
	defer mk.mutex.RUnlock()

	status := KernelStatus{
		State:   mk.state,
		Version: "1.0.0", // TODO: 从配置或构建信息获取版本
	}

	if !mk.startedAt.IsZero() {
		status.StartedAt = mk.startedAt.Unix()
		// 使用毫秒来避免精度问题
		status.Uptime = int64(time.Since(mk.startedAt).Milliseconds())
	}

	if mk.pluginManager != nil {
		status.PluginCount = mk.pluginManager.GetLoadedPluginCount()
	}

	return status
}

// GetContainer 获取依赖注入容器
func (mk *MicroKernel) GetContainer() *dig.Container {
	mk.mutex.RLock()
	defer mk.mutex.RUnlock()
	return mk.container
}

// setupContainer 设置依赖注入容器
func (mk *MicroKernel) setupContainer() error {
	// 注册核心组件到容器
	if err := mk.container.Provide(func() *slog.Logger { return mk.logger }); err != nil {
		return fmt.Errorf("failed to provide logger: %w", err)
	}

	if err := mk.container.Provide(func() event.EventBus { return mk.eventBus }); err != nil {
		return fmt.Errorf("failed to provide event bus: %w", err)
	}

	if err := mk.container.Provide(func() ServiceRegistry { return mk.serviceRegistry }); err != nil {
		return fmt.Errorf("failed to provide service registry: %w", err)
	}

	if err := mk.container.Provide(func() SecurityManager { return mk.securityManager }); err != nil {
		return fmt.Errorf("failed to provide security manager: %w", err)
	}

	if err := mk.container.Provide(func() PluginManager { return mk.pluginManager }); err != nil {
		return fmt.Errorf("failed to provide plugin manager: %w", err)
	}

	if err := mk.container.Provide(func() *koanf.Koanf { return mk.config }); err != nil {
		return fmt.Errorf("failed to provide config: %w", err)
	}

	if err := mk.container.Provide(func() Kernel { return mk }); err != nil {
		return fmt.Errorf("failed to provide kernel: %w", err)
	}

	mk.logger.Info("Dependency injection container setup completed")
	return nil
}