package kernel

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
)

// PluginManager 插件管理器接口
type PluginManager interface {
	// 生命周期管理
	Initialize(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Shutdown(ctx context.Context) error

	// 插件管理
	LoadPlugin(pluginPath string, pluginType PluginType) error
	UnloadPlugin(pluginName string) error
	StartPlugin(pluginName string) error
	StopPlugin(pluginName string) error
	ReloadPlugin(pluginName string) error

	// 插件查询
	GetPlugin(pluginName string) (Plugin, error)
	GetLoadedPlugins() []Plugin
	GetLoadedPluginCount() int
	IsPluginLoaded(pluginName string) bool
	GetPluginInfo(pluginName string) (*PluginInfo, error)

	// 插件发现
	ScanPlugins(directory string) ([]*PluginInfo, error)
	RegisterPlugin(plugin Plugin) error
	UnregisterPlugin(pluginName string) error
}

// Plugin 插件基础接口
type Plugin interface {
	// 插件元信息
	GetInfo() *PluginInfo
	GetCapabilities() []string
	GetDependencies() []string

	// 生命周期管理
	Initialize(ctx PluginContext) error
	Start() error
	Stop() error
	Cleanup() error

	// 健康检查
	HealthCheck() error

	// 配置管理
	ValidateConfig(config map[string]interface{}) error
	UpdateConfig(config map[string]interface{}) error
}

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	License     string            `json:"license"`
	Homepage    string            `json:"homepage"`
	Tags        []string          `json:"tags"`
	Config      map[string]string `json:"config"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// PluginContext 插件上下文
type PluginContext interface {
	GetLogger() *slog.Logger
	GetEventBus() event.EventBus
	GetServiceRegistry() ServiceRegistry
	GetSecurityManager() SecurityManager
	GetConfig() map[string]interface{}
}

// PluginState 插件状态
type PluginState int

const (
	PluginStateUnloaded PluginState = iota
	PluginStateLoaded
	PluginStateStarting
	PluginStateRunning
	PluginStateStopping
	PluginStateStopped
	PluginStateError
)

// String returns the string representation of PluginState
func (ps PluginState) String() string {
	switch ps {
	case PluginStateUnloaded:
		return "unloaded"
	case PluginStateLoaded:
		return "loaded"
	case PluginStateStarting:
		return "starting"
	case PluginStateRunning:
		return "running"
	case PluginStateStopping:
		return "stopping"
	case PluginStateStopped:
		return "stopped"
	case PluginStateError:
		return "error"
	default:
		return "unknown"
	}
}

// LoadedPlugin 已加载的插件
type LoadedPlugin struct {
	Plugin      Plugin      `json:"-"`
	Info        *PluginInfo `json:"info"`
	Type        PluginType  `json:"type"`
	State       PluginState `json:"state"`
	LoadedAt    time.Time   `json:"loaded_at"`
	StartedAt   time.Time   `json:"started_at"`
	LastError   string      `json:"last_error"`
	mutex       sync.RWMutex `json:"-"`
}

// SetState 设置插件状态
func (lp *LoadedPlugin) SetState(state PluginState) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	lp.State = state
}

// GetState 获取插件状态
func (lp *LoadedPlugin) GetState() PluginState {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	return lp.State
}

// SetError 设置错误信息
func (lp *LoadedPlugin) SetError(err error) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	if err != nil {
		lp.LastError = err.Error()
		lp.State = PluginStateError
	} else {
		lp.LastError = ""
	}
}

// DefaultPluginManager 默认插件管理器实现
type DefaultPluginManager struct {
	logger          *slog.Logger
	eventBus        event.EventBus
	securityManager SecurityManager
	serviceRegistry ServiceRegistry

	// 插件管理
	plugins map[string]*LoadedPlugin
	mutex   sync.RWMutex

	// 生命周期管理
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
}

// NewPluginManager 创建新的插件管理器
func NewPluginManager(logger *slog.Logger, eventBus event.EventBus, securityManager SecurityManager, serviceRegistry ServiceRegistry) PluginManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &DefaultPluginManager{
		logger:          logger,
		eventBus:        eventBus,
		securityManager: securityManager,
		serviceRegistry: serviceRegistry,
		plugins:         make(map[string]*LoadedPlugin),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Initialize 初始化插件管理器
func (pm *DefaultPluginManager) Initialize(ctx context.Context) error {
	pm.logger.Info("Initializing plugin manager...")
	return nil
}

// Start 启动插件管理器
func (pm *DefaultPluginManager) Start(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.running {
		return fmt.Errorf("plugin manager already running")
	}

	pm.running = true
	pm.logger.Info("Plugin manager started")
	return nil
}

// Stop 停止插件管理器
func (pm *DefaultPluginManager) Stop(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.running {
		return fmt.Errorf("plugin manager not running")
	}

	// 停止所有插件
	for name, plugin := range pm.plugins {
		if plugin.GetState() == PluginStateRunning {
			if err := plugin.Plugin.Stop(); err != nil {
				pm.logger.Error("Failed to stop plugin", "name", name, "error", err)
			}
		}
	}

	pm.running = false
	pm.logger.Info("Plugin manager stopped")
	return nil
}

// Shutdown 关闭插件管理器
func (pm *DefaultPluginManager) Shutdown(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 卸载所有插件
	for name := range pm.plugins {
		if err := pm.unloadPluginUnsafe(name); err != nil {
			pm.logger.Error("Failed to unload plugin during shutdown", "name", name, "error", err)
		}
	}

	pm.cancel()
	pm.logger.Info("Plugin manager shutdown completed")
	return nil
}

// LoadPlugin 加载插件
func (pm *DefaultPluginManager) LoadPlugin(pluginPath string, pluginType PluginType) error {
	// TODO: 实现插件加载逻辑
	return fmt.Errorf("plugin loading not implemented yet")
}

// UnloadPlugin 卸载插件
func (pm *DefaultPluginManager) UnloadPlugin(pluginName string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	return pm.unloadPluginUnsafe(pluginName)
}

// unloadPluginUnsafe 卸载插件（不加锁版本）
func (pm *DefaultPluginManager) unloadPluginUnsafe(pluginName string) error {
	plugin, exists := pm.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	// 停止插件
	if plugin.GetState() == PluginStateRunning {
		if err := plugin.Plugin.Stop(); err != nil {
			pm.logger.Error("Failed to stop plugin during unload", "name", pluginName, "error", err)
		}
	}

	// 清理插件
	if err := plugin.Plugin.Cleanup(); err != nil {
		pm.logger.Error("Failed to cleanup plugin", "name", pluginName, "error", err)
	}

	delete(pm.plugins, pluginName)
	pm.logger.Info("Plugin unloaded", "name", pluginName)
	return nil
}

// StartPlugin 启动插件
func (pm *DefaultPluginManager) StartPlugin(pluginName string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, exists := pm.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	state := plugin.GetState()
	
	// 检查是否已经在启动过程中或已经运行
	if state == PluginStateStarting || state == PluginStateRunning {
		return nil // 已经启动或正在启动，直接返回成功
	}

	if state != PluginStateLoaded && state != PluginStateStopped {
		return fmt.Errorf("plugin %s not in a startable state: %s", pluginName, state.String())
	}

	// 创建插件上下文
	ctx := &defaultPluginContext{
		logger:          pm.logger,
		eventBus:        pm.eventBus,
		serviceRegistry: pm.serviceRegistry,
		securityManager: pm.securityManager,
		config:          make(map[string]interface{}),
	}

	// 初始化插件（如果还没有初始化）
	if state == PluginStateLoaded {
		if err := plugin.Plugin.Initialize(ctx); err != nil {
			plugin.SetError(err)
			return fmt.Errorf("failed to initialize plugin %s: %w", pluginName, err)
		}
	}

	plugin.SetState(PluginStateStarting)
	if err := plugin.Plugin.Start(); err != nil {
		plugin.SetError(err)
		return fmt.Errorf("failed to start plugin %s: %w", pluginName, err)
	}

	plugin.SetState(PluginStateRunning)
	plugin.StartedAt = time.Now()
	pm.logger.Info("Plugin started", "name", pluginName)
	return nil
}

// StopPlugin 停止插件
func (pm *DefaultPluginManager) StopPlugin(pluginName string) error {
	pm.mutex.RLock()
	plugin, exists := pm.plugins[pluginName]
	pm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	if plugin.GetState() != PluginStateRunning {
		return fmt.Errorf("plugin %s not running", pluginName)
	}

	plugin.SetState(PluginStateStopping)
	if err := plugin.Plugin.Stop(); err != nil {
		plugin.SetError(err)
		return fmt.Errorf("failed to stop plugin %s: %w", pluginName, err)
	}

	plugin.SetState(PluginStateStopped)
	pm.logger.Info("Plugin stopped", "name", pluginName)
	return nil
}

// ReloadPlugin 重新加载插件
func (pm *DefaultPluginManager) ReloadPlugin(pluginName string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, exists := pm.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	// 如果插件正在运行，先停止它
	if plugin.GetState() == PluginStateRunning {
		plugin.SetState(PluginStateStopping)
		if err := plugin.Plugin.Stop(); err != nil {
			plugin.SetError(err)
			pm.logger.Warn("Failed to stop plugin during reload", "name", pluginName, "error", err)
		}
	}

	// 清理插件
	if err := plugin.Plugin.Cleanup(); err != nil {
		pm.logger.Warn("Failed to cleanup plugin during reload", "name", pluginName, "error", err)
	}

	// 重置插件状态
	plugin.SetState(PluginStateLoaded)
	plugin.StartedAt = time.Time{}

	pm.logger.Info("Plugin reloaded", "name", pluginName)
	return nil
}

// GetPlugin 获取插件
func (pm *DefaultPluginManager) GetPlugin(pluginName string) (Plugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	return plugin.Plugin, nil
}

// GetLoadedPlugins 获取所有已加载的插件
func (pm *DefaultPluginManager) GetLoadedPlugins() []Plugin {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin.Plugin)
	}

	return plugins
}

// GetLoadedPluginCount 获取已加载插件数量
func (pm *DefaultPluginManager) GetLoadedPluginCount() int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return len(pm.plugins)
}

// IsPluginLoaded 检查插件是否已加载
func (pm *DefaultPluginManager) IsPluginLoaded(pluginName string) bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	_, exists := pm.plugins[pluginName]
	return exists
}

// GetPluginInfo 获取插件信息
func (pm *DefaultPluginManager) GetPluginInfo(pluginName string) (*PluginInfo, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	return plugin.Info, nil
}

// ScanPlugins 扫描插件目录
func (pm *DefaultPluginManager) ScanPlugins(directory string) ([]*PluginInfo, error) {
	// TODO: 实现插件扫描逻辑
	return nil, fmt.Errorf("plugin scanning not implemented yet")
}

// RegisterPlugin 注册插件
func (pm *DefaultPluginManager) RegisterPlugin(plugin Plugin) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	info := plugin.GetInfo()
	if info == nil {
		return fmt.Errorf("plugin info is nil")
	}

	if _, exists := pm.plugins[info.Name]; exists {
		return fmt.Errorf("plugin %s already registered", info.Name)
	}

	loadedPlugin := &LoadedPlugin{
		Plugin:   plugin,
		Info:     info,
		Type:     PluginTypeDynamicLibrary, // 默认类型
		State:    PluginStateLoaded,
		LoadedAt: time.Now(),
	}

	pm.plugins[info.Name] = loadedPlugin
	pm.logger.Info("Plugin registered", "name", info.Name)
	return nil
}

// UnregisterPlugin 注销插件
func (pm *DefaultPluginManager) UnregisterPlugin(pluginName string) error {
	return pm.UnloadPlugin(pluginName)
}

// defaultPluginContext 默认插件上下文实现
type defaultPluginContext struct {
	logger          *slog.Logger
	eventBus        event.EventBus
	serviceRegistry ServiceRegistry
	securityManager SecurityManager
	config          map[string]interface{}
}

// GetLogger 获取日志器
func (ctx *defaultPluginContext) GetLogger() *slog.Logger {
	return ctx.logger
}

// GetEventBus 获取事件总线
func (ctx *defaultPluginContext) GetEventBus() event.EventBus {
	return ctx.eventBus
}

// GetServiceRegistry 获取服务注册表
func (ctx *defaultPluginContext) GetServiceRegistry() ServiceRegistry {
	return ctx.serviceRegistry
}

// GetSecurityManager 获取安全管理器
func (ctx *defaultPluginContext) GetSecurityManager() SecurityManager {
	return ctx.securityManager
}

// GetConfig 获取配置
func (ctx *defaultPluginContext) GetConfig() map[string]interface{} {
	return ctx.config
}