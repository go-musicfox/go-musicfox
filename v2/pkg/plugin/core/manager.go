package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/plugin/loader"
)

// canUnloadFromState 检查插件状态是否可以卸载
func canUnloadFromState(state loader.PluginState) bool {
	switch state {
	case loader.PluginStateLoaded, loader.PluginStateStopped, loader.PluginStateError:
		return true
	case loader.PluginStateRunning, loader.PluginStateUnloading, loader.PluginStateUnloaded:
		return false
	default:
		return false
	}
}

// isValidStateTransition 检查状态转换是否有效
func isValidStateTransition(from, to loader.PluginState) bool {
	switch from {
	case loader.PluginStateLoaded:
		return to == loader.PluginStateUnloading || to == loader.PluginStateRunning || to == loader.PluginStateError
	case loader.PluginStateStopped:
		return to == loader.PluginStateUnloading || to == loader.PluginStateRunning || to == loader.PluginStateError
	case loader.PluginStateError:
		return to == loader.PluginStateUnloading || to == loader.PluginStateRunning
	case loader.PluginStateRunning:
		return to == loader.PluginStateStopped || to == loader.PluginStateError
	case loader.PluginStateUnloading:
		return to == loader.PluginStateUnloaded || to == loader.PluginStateError
	default:
		return false
	}
}

// HybridPluginManager 混合插件管理器
// 集成所有类型的插件加载器，提供统一的插件管理接口
type HybridPluginManager struct {
	// 插件加载器
	dynamicLoader   *loader.DynamicPluginLoader
	rpcLoader       *loader.RPCPluginLoader
	wasmLoader      *loader.WASMPluginLoader
	hotReloadLoader *loader.HotReloadPluginLoader

	// 核心组件
	eventBus        loader.EventBus
	serviceRegistry loader.ServiceRegistry
	securityManager loader.SecurityManager
	logger          *slog.Logger

	// 插件管理
	plugins map[string]*ManagedPlugin // pluginID -> ManagedPlugin
	mutex   sync.RWMutex

	// 配置
	config *ManagerConfig

	// 监控
	healthChecker HealthChecker
	monitor       *PluginMonitor

	// 生命周期
	ctx    context.Context
	cancel context.CancelFunc

	// 扩展功能
	pluginGroups    map[string][]*ManagedPlugin // 插件组管理
	circuitBreakers map[string]*CircuitBreaker   // 熔断器
	startQueue      chan *StartRequest           // 启动队列
	stopQueue       chan *StopRequest            // 停止队列
	workerPool      *WorkerPool                  // 工作池
}

// ManagedPlugin 被管理的插件
type ManagedPlugin struct {
	ID          string
	Path        string
	Type        loader.LoaderType
	Plugin      loader.Plugin
	Loader      loader.PluginLoader
	State       loader.PluginState
	LoadTime    time.Time
	StartTime   time.Time
	StopTime    time.Time
	Dependencies []string
	Metadata    map[string]interface{}
	Context     loader.PluginContext
	TempDir     string // 临时目录路径
	mutex       sync.RWMutex

	// 扩展字段
	Priority    int           // 启动优先级
	Group       string        // 插件组
	RetryCount  int           // 重试次数
	LastError   error         // 最后一次错误
	FailureCount int          // 失败次数
	CircuitBreakerOpen bool  // 熔断器状态
	Hooks       *PluginHooks  // 钩子函数
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	// 加载器配置
	MaxPlugins       int
	LoadTimeout      time.Duration
	StartTimeout     time.Duration
	StopTimeout      time.Duration
	HealthCheckInterval time.Duration

	// 安全配置
	EnableSecurity   bool
	AllowedPaths     []string
	ResourceLimits   *loader.ResourceLimits

	// 监控配置
	EnableMonitoring bool
	MetricsInterval  time.Duration

	// 热重载配置
	EnableHotReload  bool
	WatchInterval    time.Duration

	// 启动和停止配置
	MaxRetries       int           // 最大重试次数
	RetryDelay       time.Duration // 重试延迟
	EnableCircuitBreaker bool      // 启用熔断器
	CircuitBreakerThreshold int    // 熔断器阈值
	EnableGracefulShutdown bool    // 启用优雅关闭
	ShutdownTimeout  time.Duration // 关闭超时
}





// NewHybridPluginManager 创建新的混合插件管理器
func NewHybridPluginManager(
	eventBus loader.EventBus,
	serviceRegistry loader.ServiceRegistry,
	securityManager loader.SecurityManager,
	logger *slog.Logger,
	config *ManagerConfig,
) (*HybridPluginManager, error) {
	if eventBus == nil || serviceRegistry == nil || logger == nil {
		return nil, fmt.Errorf("required dependencies cannot be nil")
	}

	if config == nil {
		config = DefaultManagerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &HybridPluginManager{
		eventBus:        eventBus,
		serviceRegistry: serviceRegistry,
		securityManager: securityManager,
		logger:          logger,
		plugins:         make(map[string]*ManagedPlugin),
		config:          config,
		ctx:             ctx,
		cancel:          cancel,
	}

	// 初始化加载器
	if err := m.initializeLoaders(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize loaders: %w", err)
	}

	// 初始化监控组件
	if err := m.initializeMonitoring(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize monitoring: %w", err)
	}

	return m, nil
}

// DefaultManagerConfig 返回默认管理器配置
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		MaxPlugins:          100,
		LoadTimeout:         30 * time.Second,
		StartTimeout:        10 * time.Second,
		StopTimeout:         10 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		EnableSecurity:      true,
		AllowedPaths:        []string{},
		ResourceLimits: &loader.ResourceLimits{
			MaxMemoryMB:   256,
			MaxCPUPercent: 50.0,
			EnforceMode:   loader.EnforceModeLimit,
		},
		EnableMonitoring: true,
		MetricsInterval:  10 * time.Second,
		EnableHotReload:  false,
		WatchInterval:    1 * time.Second,
		// 启动和停止配置
		MaxRetries:       3,
		RetryDelay:       2 * time.Second,
		EnableCircuitBreaker: true,
		CircuitBreakerThreshold: 5,
		EnableGracefulShutdown: true,
		ShutdownTimeout:  30 * time.Second,
	}
}

// initializeLoaders 初始化所有插件加载器
func (m *HybridPluginManager) initializeLoaders() error {
	// 初始化动态库加载器
	m.dynamicLoader = loader.NewDynamicPluginLoader(m.securityManager, m.logger)
	if m.dynamicLoader == nil {
		return fmt.Errorf("failed to create dynamic plugin loader")
	}

	// 初始化RPC加载器
	m.rpcLoader = loader.NewRPCPluginLoaderWithSecurity(m.securityManager, m.logger)
	if m.rpcLoader == nil {
		return fmt.Errorf("failed to create RPC plugin loader")
	}

	// 初始化WASM加载器
	m.wasmLoader = loader.NewWASMPluginLoader(m.securityManager, m.logger)
	if m.wasmLoader == nil {
		return fmt.Errorf("failed to create WASM plugin loader")
	}

	// 初始化热重载加载器
	if m.config.EnableHotReload {
		m.hotReloadLoader = loader.NewHotReloadPluginLoaderWithSecurity(m.securityManager, m.logger)
		if m.hotReloadLoader == nil {
			return fmt.Errorf("failed to create hot reload plugin loader")
		}
	}

	// 验证所有加载器都已正确初始化
	if err := m.validateLoaders(); err != nil {
		return fmt.Errorf("loader validation failed: %w", err)
	}

	if m.logger != nil {
		m.logger.Info("All plugin loaders initialized successfully",
			slog.Int("dynamic_loader", 1),
			slog.Int("rpc_loader", 1),
			slog.Int("wasm_loader", 1),
			slog.Bool("hotreload_enabled", m.config.EnableHotReload),
		)
	}
	return nil
}

// UnloadProgress 卸载进度信息
type UnloadProgress struct {
	PluginID    string    `json:"plugin_id"`
	PluginName  string    `json:"plugin_name"`
	Stage       string    `json:"stage"`
	Progress    float64   `json:"progress"`    // 0.0 - 1.0
	Message     string    `json:"message"`
	StartTime   time.Time `json:"start_time"`
	ElapsedTime time.Duration `json:"elapsed_time"`
	Error       string    `json:"error,omitempty"`
}

// UnloadProgressCallback 卸载进度回调函数
type UnloadProgressCallback func(progress *UnloadProgress)

// UnloadPluginWithProgress 带进度监控的插件卸载
func (m *HybridPluginManager) UnloadPluginWithProgress(pluginID string, options *UnloadOptions, progressCallback UnloadProgressCallback) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 验证插件ID
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	// 验证选项
	if options == nil {
		options = &UnloadOptions{}
	}

	managedPlugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	managedPlugin.mutex.Lock()
	defer managedPlugin.mutex.Unlock()

	pluginInfo := managedPlugin.Plugin.GetInfo()
	pluginName := pluginID
	if pluginInfo != nil {
		pluginName = pluginInfo.Name
	}

	startTime := time.Now()

	// 创建进度报告函数
	reportProgress := func(stage string, progress float64, message string, err error) {
		if progressCallback != nil {
			progressInfo := &UnloadProgress{
				PluginID:    pluginID,
				PluginName:  pluginName,
				Stage:       stage,
				Progress:    progress,
				Message:     message,
				StartTime:   startTime,
				ElapsedTime: time.Since(startTime),
			}
			if err != nil {
				progressInfo.Error = err.Error()
			}
			progressCallback(progressInfo)
		}
	}

	// 开始卸载
	reportProgress("validation", 0.1, "Validating plugin state", nil)

	// 检查插件状态是否可以卸载
	if !canUnloadFromState(managedPlugin.State) && !options.ForceUnload {
		err := fmt.Errorf("plugin %s is not in a unloadable state: %s (use ForceUnload to override)", pluginID, managedPlugin.State.String())
		reportProgress("validation", 0.1, "Validation failed", err)
		return err
	}

	reportProgress("dependency_check", 0.2, "Checking dependencies", nil)

	// 检查依赖关系（除非强制卸载）
	if !options.ForceUnload {
		if err := m.checkDependentPluginsForUnload(managedPlugin); err != nil {
			reportProgress("dependency_check", 0.2, "Dependency check failed", err)
			return fmt.Errorf("dependency check failed for plugin %s: %w", pluginID, err)
		}
	}

	reportProgress("pre_hooks", 0.3, "Executing pre-unload hooks", nil)

	// 执行预卸载钩子
	if options.Hooks != nil && options.Hooks.PreUnload != nil {
		ctx := context.Background()
		if options.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, options.Timeout)
			defer cancel()
		}
		if err := options.Hooks.PreUnload(ctx, managedPlugin); err != nil {
			reportProgress("pre_hooks", 0.3, "Pre-unload hook failed", err)
			return fmt.Errorf("pre-unload hook failed for plugin %s: %w", pluginID, err)
		}
	}

	reportProgress("unloading", 0.4, "Starting plugin unload", nil)

	// 设置超时时间
	timeout := m.config.StopTimeout
	if options.Timeout > 0 {
		timeout = options.Timeout
	}

	// 重试逻辑
	retryCount := options.RetryCount
	if retryCount <= 0 {
		retryCount = 1
	}
	retryDelay := options.RetryDelay
	if retryDelay <= 0 {
		retryDelay = time.Second
	}

	var lastErr error
	for attempt := 0; attempt < retryCount; attempt++ {
		if attempt > 0 {
			reportProgress("unloading", 0.4+float64(attempt)*0.1, fmt.Sprintf("Retry attempt %d/%d", attempt+1, retryCount), nil)
			time.Sleep(retryDelay)
		}

		lastErr = m.doUnloadPluginWithProgress(managedPlugin, timeout, options, reportProgress)
		if lastErr == nil {
			break
		}

		// 执行错误钩子
		if options.Hooks != nil && options.Hooks.OnError != nil {
			options.Hooks.OnError(pluginID, lastErr)
		}
	}

	if lastErr != nil {
		reportProgress("recovery", 0.8, "Attempting error recovery", lastErr)
		// 尝试错误恢复
		if err := m.recoverFromUnloadError(managedPlugin, lastErr, options); err != nil {
			reportProgress("recovery", 0.8, "Error recovery failed", err)
			return fmt.Errorf("failed to unload plugin %s after %d attempts and recovery failed: original error: %w, recovery error: %v", pluginID, retryCount, lastErr, err)
		}
		reportProgress("recovery", 0.9, "Error recovery completed", nil)
	}

	reportProgress("post_hooks", 0.95, "Executing post-unload hooks", nil)

	// 执行后卸载钩子
	if options.Hooks != nil && options.Hooks.PostUnload != nil {
		ctx := context.Background()
		if options.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, options.Timeout)
			defer cancel()
		}
		if err := options.Hooks.PostUnload(ctx, managedPlugin); err != nil {
			reportProgress("post_hooks", 0.95, "Post-unload hook failed", err)
			m.logger.Warn("Post-unload hook failed",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	reportProgress("completed", 1.0, "Plugin unload completed successfully", nil)

	// 级联卸载依赖插件
	if options.CascadeUnload {
		if err := m.cascadeUnloadDependents(pluginID, options); err != nil {
			m.logger.Warn("Cascade unload failed",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	if lastErr != nil {
		return fmt.Errorf("plugin unload failed but recovered: %w", lastErr)
	}

	return nil
}

// doUnloadPluginWithProgress 带进度报告的实际卸载逻辑
func (m *HybridPluginManager) doUnloadPluginWithProgress(managedPlugin *ManagedPlugin, timeout time.Duration, options *UnloadOptions, reportProgress func(string, float64, string, error)) error {
	pluginID := managedPlugin.ID
	_ = managedPlugin.Plugin.GetInfo()

	reportProgress("stopping", 0.5, "Stopping plugin if running", nil)

	// 如果插件正在运行，先停止它
	if managedPlugin.State == loader.PluginStateRunning {
		// 临时解锁以调用StopPlugin
		managedPlugin.mutex.Unlock()
		m.mutex.Unlock()
		stopOptions := &StopOptions{
			Timeout:          timeout,
			ForceStop:        options.ForceUnload,
			GracefulShutdown: options.GracefulShutdown,
			SkipCleanup:      true, // 在卸载时统一清理
		}
		if err := m.StopPluginWithOptions(pluginID, stopOptions); err != nil {
			m.mutex.Lock()
			managedPlugin.mutex.Lock()
			if !options.ForceUnload {
				reportProgress("stopping", 0.5, "Failed to stop plugin", err)
				return fmt.Errorf("failed to stop plugin %s before unloading: %w", pluginID, err)
			}
			reportProgress("stopping", 0.5, "Failed to stop plugin gracefully, forcing unload", err)
		} else {
			m.mutex.Lock()
			managedPlugin.mutex.Lock()
		}
	}

	reportProgress("state_transition", 0.55, "Validating state transition", nil)

	// 验证状态转换
	if !isValidStateTransition(managedPlugin.State, loader.PluginStateUnloading) {
		if !options.ForceUnload {
			err := fmt.Errorf("invalid state transition from %s to unloading for plugin %s", managedPlugin.State.String(), pluginID)
			reportProgress("state_transition", 0.55, "Invalid state transition", err)
			return err
		}
		reportProgress("state_transition", 0.55, "Forcing invalid state transition", nil)
	}

	// 更新状态为卸载中
	managedPlugin.State = loader.PluginStateUnloading

	// 发布卸载开始事件
	m.publishUnloadEvent("plugin.unloading", managedPlugin, true, "")

	reportProgress("cleanup_monitoring", 0.6, "Stopping health check and monitoring", nil)

	// 停止健康检查
	if m.healthChecker != nil {
		m.healthChecker.StopMonitoring()
	}

	// 从监控中移除
	if m.monitor != nil {
		m.monitor.RemovePlugin(managedPlugin.ID)
	}

	reportProgress("plugin_cleanup", 0.65, "Executing plugin cleanup", nil)

	// 创建卸载上下文
	unloadCtx, unloadCancel := context.WithTimeout(context.Background(), timeout)
	defer unloadCancel()

	// 执行插件清理
	var cleanupErr error
	if managedPlugin.Plugin != nil {
		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					cleanupErr = fmt.Errorf("plugin cleanup panicked: %v", r)
				}
			}()
			if err := managedPlugin.Plugin.Cleanup(); err != nil {
				cleanupErr = fmt.Errorf("plugin cleanup failed: %w", err)
			}
		}()

		select {
		case <-unloadCtx.Done():
			if !options.ForceUnload {
				managedPlugin.State = loader.PluginStateError
				err := fmt.Errorf("plugin cleanup timeout for %s: %w", pluginID, unloadCtx.Err())
				reportProgress("plugin_cleanup", 0.65, "Plugin cleanup timeout", err)
				return err
			}
			reportProgress("plugin_cleanup", 0.65, "Plugin cleanup timeout, forcing cleanup", nil)
		case <-done:
			if cleanupErr != nil && !options.ForceUnload {
				managedPlugin.State = loader.PluginStateError
				reportProgress("plugin_cleanup", 0.65, "Plugin cleanup failed", cleanupErr)
				return cleanupErr
			}
		}
	}

	if cleanupErr != nil {
		reportProgress("plugin_cleanup", 0.65, "Plugin cleanup failed, continuing", cleanupErr)
	}

	reportProgress("service_cleanup", 0.7, "Unregistering plugin services", nil)

	// 注销插件服务
	m.unregisterPluginServices(managedPlugin)

	reportProgress("resource_cleanup", 0.75, "Cleaning up plugin resources", nil)

	// 清理插件资源（除非跳过）
	if !options.SkipCleanup {
		// 更新状态为卸载中（继续卸载过程）
		managedPlugin.State = loader.PluginStateUnloading

		// 执行清理钩子
		if options.Hooks != nil && options.Hooks.OnCleanup != nil {
			if err := options.Hooks.OnCleanup(unloadCtx, managedPlugin); err != nil {
				reportProgress("resource_cleanup", 0.75, "Cleanup hook failed", err)
			}
		}

		if err := m.cleanupPluginResourcesAdvanced(managedPlugin, unloadCtx); err != nil {
			if !options.ForceUnload {
				managedPlugin.State = loader.PluginStateError
				reportProgress("resource_cleanup", 0.75, "Resource cleanup failed", err)
				return fmt.Errorf("failed to cleanup plugin resources: %w", err)
			}
			reportProgress("resource_cleanup", 0.75, "Resource cleanup failed, continuing", err)
		}

		// 执行安全清理
		if err := m.secureCleanupPlugin(managedPlugin, unloadCtx); err != nil {
			reportProgress("resource_cleanup", 0.75, "Secure cleanup failed, continuing", err)
		}
	}

	reportProgress("loader_cleanup", 0.8, "Using loader to unload plugin", nil)

	// 使用对应的加载器卸载插件
	pluginLoader := m.selectLoader(managedPlugin.Type)
	if pluginLoader == nil {
		reportProgress("loader_cleanup", 0.8, "No suitable loader found, skipping loader unload", nil)
	} else {
		if err := pluginLoader.UnloadPlugin(unloadCtx, pluginID); err != nil {
			if !options.ForceUnload {
				managedPlugin.State = loader.PluginStateError
				reportProgress("loader_cleanup", 0.8, "Loader unload failed", err)
				return fmt.Errorf("loader unload failed: %w", err)
			}
			reportProgress("loader_cleanup", 0.8, "Loader unload failed, continuing", err)
		}
	}

	reportProgress("finalization", 0.9, "Finalizing plugin unload", nil)

	// 更新状态为已卸载
	managedPlugin.State = loader.PluginStateUnloaded

	// 从管理器中移除
	delete(m.plugins, pluginID)

	// 从插件组中移除
	if managedPlugin.Group != "" {
		m.removeFromPluginGroup(managedPlugin.Group, managedPlugin)
	}

	// 发布卸载完成事件
	m.publishUnloadEvent("plugin.unloaded", managedPlugin, true, "")

	reportProgress("finalization", 0.95, "Plugin unload finalized successfully", nil)

	return nil
}

// checkDependentPluginsForUnload 检查依赖此插件的其他插件
func (m *HybridPluginManager) checkDependentPluginsForUnload(targetPlugin *ManagedPlugin) error {
	targetInfo := targetPlugin.Plugin.GetInfo()
	if targetInfo == nil {
		return fmt.Errorf("target plugin info is nil")
	}

	var dependentPlugins []*ManagedPlugin

	// 遍历所有插件，查找依赖目标插件的插件
	for _, plugin := range m.plugins {
		if plugin.ID == targetPlugin.ID {
			continue // 跳过自己
		}

		// 检查插件的依赖列表
		for _, dep := range plugin.Dependencies {
			if dep == targetInfo.Name || dep == targetPlugin.ID {
				dependentPlugins = append(dependentPlugins, plugin)
				break
			}
		}
	}

	if len(dependentPlugins) > 0 {
		dependentNames := make([]string, len(dependentPlugins))
		for i, plugin := range dependentPlugins {
			pluginInfo := plugin.Plugin.GetInfo()
			if pluginInfo != nil {
				dependentNames[i] = fmt.Sprintf("%s(%s)", pluginInfo.Name, plugin.ID)
			} else {
				dependentNames[i] = plugin.ID
			}
		}
		return fmt.Errorf("cannot unload plugin %s: it is required by %d other plugin(s): %s",
			targetInfo.Name, len(dependentPlugins), strings.Join(dependentNames, ", "))
	}

	return nil
}

// cascadeUnloadDependents 级联卸载依赖插件
func (m *HybridPluginManager) cascadeUnloadDependents(pluginID string, options *UnloadOptions) error {
	targetPlugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	targetInfo := targetPlugin.Plugin.GetInfo()
	if targetInfo == nil {
		return fmt.Errorf("target plugin info is nil")
	}

	// 查找依赖目标插件的插件
	var dependentPlugins []*ManagedPlugin
	for _, plugin := range m.plugins {
		if plugin.ID == pluginID {
			continue // 跳过自己
		}

		// 检查插件的依赖列表
		for _, dep := range plugin.Dependencies {
			if dep == targetInfo.Name || dep == targetPlugin.ID {
				dependentPlugins = append(dependentPlugins, plugin)
				break
			}
		}
	}

	if len(dependentPlugins) == 0 {
		return nil // 没有依赖插件
	}

	m.logger.Info("Starting cascade unload of dependent plugins",
		slog.String("target_plugin", pluginID),
		slog.Int("dependent_count", len(dependentPlugins)),
	)

	// 按优先级排序（高优先级的插件先卸载）
	sort.Slice(dependentPlugins, func(i, j int) bool {
		return dependentPlugins[i].Priority > dependentPlugins[j].Priority
	})

	var errors []error
	for _, plugin := range dependentPlugins {
		pluginInfo := plugin.Plugin.GetInfo()
		pluginName := plugin.ID
		if pluginInfo != nil {
			pluginName = pluginInfo.Name
		}

		m.logger.Info("Cascade unloading dependent plugin",
			slog.String("plugin_id", plugin.ID),
			slog.String("plugin_name", pluginName),
		)

		// 创建级联卸载选项
		cascadeOptions := &UnloadOptions{
			Timeout:          options.Timeout,
			ForceUnload:      options.ForceUnload,
			GracefulShutdown: options.GracefulShutdown,
			SkipCleanup:      options.SkipCleanup,
			CascadeUnload:    true, // 继续级联
			RetryCount:       options.RetryCount,
			RetryDelay:       options.RetryDelay,
		}

		if err := m.UnloadPluginWithOptions(plugin.ID, cascadeOptions); err != nil {
			errors = append(errors, fmt.Errorf("failed to cascade unload plugin %s: %w", plugin.ID, err))
			if !options.ForceUnload {
				break // 如果不是强制卸载，遇到错误就停止
			}
		}
	}

	if len(errors) > 0 {
		errorMsgs := make([]string, len(errors))
		for i, err := range errors {
			errorMsgs[i] = err.Error()
		}
		return fmt.Errorf("cascade unload errors: %s", strings.Join(errorMsgs, "; "))
	}

	m.logger.Info("Cascade unload completed successfully",
		slog.String("target_plugin", pluginID),
		slog.Int("unloaded_count", len(dependentPlugins)),
	)

	return nil
}

// removeFromPluginGroup 从插件组中移除插件
func (m *HybridPluginManager) removeFromPluginGroup(groupName string, plugin *ManagedPlugin) {
	if m.pluginGroups == nil {
		return
	}

	plugins, exists := m.pluginGroups[groupName]
	if !exists {
		return
	}

	// 查找并移除插件
	for i, p := range plugins {
		if p.ID == plugin.ID {
			// 移除插件
		m.pluginGroups[groupName] = append(plugins[:i], plugins[i+1:]...)
			break
		}
	}

	// 如果组为空，删除组
	if len(m.pluginGroups[groupName]) == 0 {
		delete(m.pluginGroups, groupName)
	}
}

// BatchUnloadPlugins 批量卸载插件
func (m *HybridPluginManager) BatchUnloadPlugins(pluginIDs []string, options *UnloadOptions) map[string]error {
	results := make(map[string]error)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, pluginID := range pluginIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			err := m.UnloadPluginWithOptions(id, options)
			mu.Lock()
			results[id] = err
			mu.Unlock()
		}(pluginID)
	}

	wg.Wait()
	return results
}

// UnloadPluginGroup 卸载插件组
func (m *HybridPluginManager) UnloadPluginGroup(groupName string, options *UnloadOptions) error {
	m.mutex.RLock()
	plugins, exists := m.pluginGroups[groupName]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin group not found: %s", groupName)
	}

	if len(plugins) == 0 {
		return nil // 空组
	}

	m.logger.Info("Unloading plugin group",
		slog.String("group_name", groupName),
		slog.Int("plugin_count", len(plugins)),
	)

	// 按优先级排序（高优先级的插件先卸载）
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Priority > plugins[j].Priority
	})

	var errors []error
	for _, plugin := range plugins {
		if err := m.UnloadPluginWithOptions(plugin.ID, options); err != nil {
			errors = append(errors, fmt.Errorf("failed to unload plugin %s: %w", plugin.ID, err))
			if !options.ForceUnload {
				break // 如果不是强制卸载，遇到错误就停止
			}
		}
	}

	if len(errors) > 0 {
		errorMsgs := make([]string, len(errors))
		for i, err := range errors {
			errorMsgs[i] = err.Error()
		}
		return fmt.Errorf("plugin group unload errors: %s", strings.Join(errorMsgs, "; "))
	}

	m.logger.Info("Plugin group unloaded successfully",
		slog.String("group_name", groupName),
		slog.Int("unloaded_count", len(plugins)),
	)

	return nil
}

// cleanupPluginContext 清理插件上下文资源
func (m *HybridPluginManager) cleanupPluginContext(managedPlugin *ManagedPlugin, ctx context.Context) error {
	if managedPlugin.Context == nil {
		return nil
	}

	// 尝试类型断言以获取清理方法
	if contextCleaner, ok := managedPlugin.Context.(interface{ Cleanup() error }); ok {
		if err := contextCleaner.Cleanup(); err != nil {
			return fmt.Errorf("context cleanup failed: %w", err)
		}
	}

	// 清理上下文中的资源引用
	managedPlugin.Context = nil
	return nil
}

// cleanupEventListeners 清理事件监听器
func (m *HybridPluginManager) cleanupEventListeners(managedPlugin *ManagedPlugin, ctx context.Context) error {
	if m.eventBus == nil {
		return nil
	}

	// 尝试清理插件相关的事件监听器
	pluginInfo := managedPlugin.Plugin.GetInfo()
	if pluginInfo != nil {
		// 构造插件相关的事件主题
		topics := []string{
			fmt.Sprintf("plugin.%s.*", pluginInfo.Name),
			fmt.Sprintf("plugin.%s.event.*", pluginInfo.Name),
			fmt.Sprintf("plugin.%s.status.*", pluginInfo.Name),
		}

		// 如果事件总线支持批量取消订阅
		if unsubscriber, ok := m.eventBus.(interface{ UnsubscribeAll(topics []string) error }); ok {
			if err := unsubscriber.UnsubscribeAll(topics); err != nil {
				return fmt.Errorf("failed to unsubscribe event listeners: %w", err)
			}
		}
	}

	return nil
}

// cleanupTemporaryFiles 清理临时文件和缓存
func (m *HybridPluginManager) cleanupTemporaryFiles(managedPlugin *ManagedPlugin, ctx context.Context) error {
	var errors []error

	// 清理插件临时目录
	if managedPlugin.TempDir != "" {
		if err := os.RemoveAll(managedPlugin.TempDir); err != nil {
			errors = append(errors, fmt.Errorf("failed to remove temp directory %s: %w", managedPlugin.TempDir, err))
		} else {
			m.logger.Debug("Removed plugin temp directory",
				slog.String("plugin_id", managedPlugin.ID),
				slog.String("temp_dir", managedPlugin.TempDir),
			)
		}
		managedPlugin.TempDir = ""
	}

	// 清理插件缓存文件
	pluginInfo := managedPlugin.Plugin.GetInfo()
	if pluginInfo != nil {
		cacheDir := filepath.Join(os.TempDir(), "go-musicfox", "plugins", pluginInfo.Name)
		if _, err := os.Stat(cacheDir); err == nil {
			if err := os.RemoveAll(cacheDir); err != nil {
				errors = append(errors, fmt.Errorf("failed to remove cache directory %s: %w", cacheDir, err))
			} else {
				m.logger.Debug("Removed plugin cache directory",
					slog.String("plugin_id", managedPlugin.ID),
					slog.String("cache_dir", cacheDir),
				)
			}
		}
	}

	if len(errors) > 0 {
		errorMsgs := make([]string, len(errors))
		for i, err := range errors {
			errorMsgs[i] = err.Error()
		}
		return fmt.Errorf("temporary files cleanup errors: %s", strings.Join(errorMsgs, "; "))
	}

	return nil
}

// cleanupNetworkConnections 清理网络连接
func (m *HybridPluginManager) cleanupNetworkConnections(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// 对于RPC插件，需要关闭网络连接
	if managedPlugin.Type == loader.LoaderTypeRPC {
		if rpcCleaner, ok := managedPlugin.Plugin.(interface{ CloseConnections() error }); ok {
			if err := rpcCleaner.CloseConnections(); err != nil {
				return fmt.Errorf("failed to close RPC connections: %w", err)
			}
			m.logger.Debug("Closed RPC connections",
				slog.String("plugin_id", managedPlugin.ID),
			)
		}
	}

	// 清理插件可能打开的其他网络连接
	if networkCleaner, ok := managedPlugin.Plugin.(interface{ CleanupNetworkResources() error }); ok {
		if err := networkCleaner.CleanupNetworkResources(); err != nil {
			return fmt.Errorf("failed to cleanup network resources: %w", err)
		}
	}

	return nil
}

// cleanupMemoryResources 清理内存资源
func (m *HybridPluginManager) cleanupMemoryResources(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// 清理插件元数据
	if managedPlugin.Metadata != nil {
		for k := range managedPlugin.Metadata {
			delete(managedPlugin.Metadata, k)
		}
		managedPlugin.Metadata = nil
	}

	// 清理插件依赖引用
	if managedPlugin.Dependencies != nil {
		managedPlugin.Dependencies = nil
	}

	// 清理插件钩子函数
	if managedPlugin.Hooks != nil {
		managedPlugin.Hooks = nil
	}

	// 如果插件支持内存清理
	if memoryCleaner, ok := managedPlugin.Plugin.(interface{ CleanupMemory() error }); ok {
		if err := memoryCleaner.CleanupMemory(); err != nil {
			return fmt.Errorf("failed to cleanup plugin memory: %w", err)
		}
	}

	// 强制垃圾回收（谨慎使用）
	runtime.GC()

	m.logger.Debug("Cleaned up memory resources",
		slog.String("plugin_id", managedPlugin.ID),
	)

	return nil
}

// cleanupFileHandles 清理文件句柄
func (m *HybridPluginManager) cleanupFileHandles(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// 如果插件支持文件句柄清理
	if fileCleaner, ok := managedPlugin.Plugin.(interface{ CloseFiles() error }); ok {
		if err := fileCleaner.CloseFiles(); err != nil {
			return fmt.Errorf("failed to close plugin files: %w", err)
		}
		m.logger.Debug("Closed plugin file handles",
			slog.String("plugin_id", managedPlugin.ID),
		)
	}

	return nil
}

// cleanupSystemResources 清理系统资源
func (m *HybridPluginManager) cleanupSystemResources(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// 清理系统级资源（如信号量、互斥锁等）
	if systemCleaner, ok := managedPlugin.Plugin.(interface{ CleanupSystemResources() error }); ok {
		if err := systemCleaner.CleanupSystemResources(); err != nil {
			return fmt.Errorf("failed to cleanup system resources: %w", err)
		}
		m.logger.Debug("Cleaned up system resources",
			slog.String("plugin_id", managedPlugin.ID),
		)
	}

	return nil
}

// cleanupPluginSpecificResources 清理插件特定资源
func (m *HybridPluginManager) cleanupPluginSpecificResources(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// 根据插件类型进行特定清理
	switch managedPlugin.Type {
	case loader.LoaderTypeDynamic:
		return m.cleanupDynamicLibraryResources(managedPlugin, ctx)
	case loader.LoaderTypeRPC:
		return m.cleanupRPCResources(managedPlugin, ctx)
	case loader.LoaderTypeWASM:
		return m.cleanupWASMResources(managedPlugin, ctx)
	case loader.LoaderTypeHotReload:
		return m.cleanupHotReloadResources(managedPlugin, ctx)
	default:
		m.logger.Debug("No specific cleanup for plugin type",
			slog.String("plugin_id", managedPlugin.ID),
			slog.String("type", managedPlugin.Type.String()),
		)
	}

	return nil
}

// cleanupDynamicLibraryResources 清理动态库资源
func (m *HybridPluginManager) cleanupDynamicLibraryResources(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// 动态库特定的清理逻辑
	if m.dynamicLoader != nil {
		if err := m.dynamicLoader.UnloadPlugin(ctx, managedPlugin.ID); err != nil {
			return fmt.Errorf("dynamic library cleanup failed: %w", err)
		}
	}
	return nil
}

// cleanupRPCResources 清理RPC资源
func (m *HybridPluginManager) cleanupRPCResources(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// RPC特定的清理逻辑
	if m.rpcLoader != nil {
		if err := m.rpcLoader.UnloadPlugin(ctx, managedPlugin.ID); err != nil {
			return fmt.Errorf("RPC cleanup failed: %w", err)
		}
	}
	return nil
}

// cleanupWASMResources 清理WASM资源
func (m *HybridPluginManager) cleanupWASMResources(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// WASM特定的清理逻辑
	if m.wasmLoader != nil {
		if err := m.wasmLoader.UnloadPlugin(ctx, managedPlugin.ID); err != nil {
			return fmt.Errorf("WASM cleanup failed: %w", err)
		}
	}
	return nil
}

// cleanupHotReloadResources 清理热重载资源
func (m *HybridPluginManager) cleanupHotReloadResources(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// 热重载特定的清理逻辑
	if m.hotReloadLoader != nil {
		if err := m.hotReloadLoader.UnloadPlugin(ctx, managedPlugin.ID); err != nil {
			return fmt.Errorf("hot reload cleanup failed: %w", err)
		}
	}
	return nil
}

// doUnloadPlugin 执行实际的插件卸载逻辑
func (m *HybridPluginManager) doUnloadPlugin(managedPlugin *ManagedPlugin, timeout time.Duration, options *UnloadOptions) error {
	pluginID := managedPlugin.ID
	pluginInfo := managedPlugin.Plugin.GetInfo()

	// 如果插件正在运行，先停止它
	if managedPlugin.State == loader.PluginStateRunning {
		// 临时解锁以调用StopPlugin
		managedPlugin.mutex.Unlock()
		m.mutex.Unlock()
		stopOptions := &StopOptions{
			Timeout:          timeout,
			ForceStop:        options.ForceUnload,
			GracefulShutdown: options.GracefulShutdown,
			SkipCleanup:      true, // 在卸载时统一清理
		}
		if err := m.StopPluginWithOptions(pluginID, stopOptions); err != nil {
			m.mutex.Lock()
			managedPlugin.mutex.Lock()
			if !options.ForceUnload {
				return fmt.Errorf("failed to stop plugin %s before unloading: %w", pluginID, err)
			}
			m.logger.Warn("Failed to stop plugin gracefully, forcing unload",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		} else {
			m.mutex.Lock()
			managedPlugin.mutex.Lock()
		}
	}

	// 验证状态转换
	if !isValidStateTransition(managedPlugin.State, loader.PluginStateUnloading) {
		if !options.ForceUnload {
			return fmt.Errorf("invalid state transition from %s to unloading for plugin %s", managedPlugin.State.String(), pluginID)
		}
		m.logger.Warn("Forcing invalid state transition",
			slog.String("plugin_id", pluginID),
			slog.String("from_state", managedPlugin.State.String()),
			slog.String("to_state", loader.PluginStateUnloading.String()),
		)
	}

	// 更新状态为卸载中
	managedPlugin.State = loader.PluginStateUnloading

	// 发布卸载开始事件
	m.publishUnloadEvent("plugin.unloading", managedPlugin, true, "")

	// 停止健康检查
	if m.healthChecker != nil {
		m.healthChecker.StopMonitoring()
	}

	// 从监控中移除
	if m.monitor != nil {
		m.monitor.RemovePlugin(managedPlugin.ID)
	}

	// 创建卸载上下文
	unloadCtx, unloadCancel := context.WithTimeout(context.Background(), timeout)
	defer unloadCancel()

	// 执行插件清理
	var cleanupErr error
	if managedPlugin.Plugin != nil {
		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					cleanupErr = fmt.Errorf("plugin cleanup panicked: %v", r)
				}
			}()
			if err := managedPlugin.Plugin.Cleanup(); err != nil {
				cleanupErr = fmt.Errorf("plugin cleanup failed: %w", err)
			}
		}()

		select {
		case <-unloadCtx.Done():
			if !options.ForceUnload {
				managedPlugin.State = loader.PluginStateError
				return fmt.Errorf("plugin cleanup timeout for %s: %w", pluginID, unloadCtx.Err())
			}
			m.logger.Warn("Plugin cleanup timeout, forcing cleanup",
				slog.String("plugin_id", pluginID),
			)
		case <-done:
			if cleanupErr != nil && !options.ForceUnload {
				managedPlugin.State = loader.PluginStateError
				return cleanupErr
			}
		}
	}

	if cleanupErr != nil {
		m.logger.Warn("Plugin cleanup failed, continuing with unload",
			slog.String("plugin_id", pluginID),
			slog.String("error", cleanupErr.Error()),
		)
	}

	// 注销插件服务
	m.unregisterPluginServices(managedPlugin)

	// 清理插件资源（除非跳过）
	if !options.SkipCleanup {
		// 更新状态为卸载中（继续卸载过程）
		managedPlugin.State = loader.PluginStateUnloading

		// 执行清理钩子
		if options.Hooks != nil && options.Hooks.OnCleanup != nil {
			if err := options.Hooks.OnCleanup(unloadCtx, managedPlugin); err != nil {
				m.logger.Warn("Cleanup hook failed",
					slog.String("plugin_id", pluginID),
					slog.String("error", err.Error()),
				)
			}
		}

		if err := m.cleanupPluginResourcesAdvanced(managedPlugin, unloadCtx); err != nil {
			if !options.ForceUnload {
				managedPlugin.State = loader.PluginStateError
				return fmt.Errorf("failed to cleanup plugin resources: %w", err)
			}
			m.logger.Warn("Failed to cleanup plugin resources, continuing with unload",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}

		// 执行安全清理
		if err := m.secureCleanupPlugin(managedPlugin, unloadCtx); err != nil {
			m.logger.Warn("Secure cleanup failed, continuing with unload",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 使用对应的加载器卸载插件
	pluginLoader := m.selectLoader(managedPlugin.Type)
	if pluginLoader == nil {
		m.logger.Warn("No suitable loader found for plugin type, skipping loader unload",
			slog.String("plugin_id", pluginID),
			slog.String("type", managedPlugin.Type.String()),
		)
	} else {
		if err := pluginLoader.UnloadPlugin(unloadCtx, pluginID); err != nil {
			if !options.ForceUnload {
				managedPlugin.State = loader.PluginStateError
				return fmt.Errorf("loader unload failed: %w", err)
			}
			m.logger.Warn("Loader unload failed, continuing with manager cleanup",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 更新状态为已卸载
	managedPlugin.State = loader.PluginStateUnloaded

	// 从管理器中移除
	delete(m.plugins, pluginID)

	// 从插件组中移除
	if managedPlugin.Group != "" {
		m.removeFromPluginGroup(managedPlugin.Group, managedPlugin)
	}

	// 发布卸载完成事件
	m.publishUnloadEvent("plugin.unloaded", managedPlugin, true, "")

	if m.logger != nil {
		m.logger.Info("Plugin unloaded successfully",
			slog.String("plugin_id", pluginID),
			slog.String("name", pluginInfo.Name),
			slog.String("type", managedPlugin.Type.String()),
		)
	}

	return nil
}

// initializeMonitoring 初始化监控组件
func (m *HybridPluginManager) initializeMonitoring() error {
	if m.config.EnableMonitoring {
		m.monitor = NewPluginMonitor(m.logger, m.config.MetricsInterval)
		if m.monitor == nil {
			return fmt.Errorf("failed to create plugin monitor")
		}
	}

	m.healthChecker = NewHealthChecker(m.logger, m.config.HealthCheckInterval)
	if m.healthChecker == nil {
		return fmt.Errorf("failed to create health checker")
	}

	if m.logger != nil {
		m.logger.Info("Monitoring components initialized successfully")
	}
	return nil
}

// LoadPlugin 加载插件
func (m *HybridPluginManager) LoadPlugin(pluginPath string, pluginType loader.LoaderType) (*ManagedPlugin, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 验证输入参数
	if pluginPath == "" {
		return nil, fmt.Errorf("plugin path cannot be empty")
	}

	// 检查插件是否已加载
	for _, existingPlugin := range m.plugins {
		if existingPlugin.Path == pluginPath {
			return nil, fmt.Errorf("plugin already loaded from path: %s", pluginPath)
		}
	}

	// 检查插件数量限制
	if len(m.plugins) >= m.config.MaxPlugins {
		return nil, fmt.Errorf("maximum number of plugins (%d) reached", m.config.MaxPlugins)
	}

	// 安全验证
	if m.config.EnableSecurity && m.securityManager != nil {
		if err := m.securityManager.ValidatePlugin(pluginPath); err != nil {
			return nil, fmt.Errorf("security validation failed: %w", err)
		}
	}

	// 选择合适的加载器
	pluginLoader := m.selectLoader(pluginType)
	if pluginLoader == nil {
		return nil, fmt.Errorf("no suitable loader found for plugin type %s", pluginType)
	}

	// 预验证插件
	if err := pluginLoader.ValidatePlugin(pluginPath); err != nil {
		return nil, fmt.Errorf("plugin validation failed: %w", err)
	}

	// 创建加载上下文
	loadCtx, loadCancel := context.WithTimeout(m.ctx, m.config.LoadTimeout)
	defer loadCancel()

	// 加载插件
	var loaderPlugin loader.Plugin
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)
		loaderPlugin, err = pluginLoader.LoadPlugin(loadCtx, pluginPath)
	}()

	select {
	case <-loadCtx.Done():
		return nil, fmt.Errorf("plugin load timeout: %w", loadCtx.Err())
	case <-done:
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin: %w", err)
		}
	}

	// 验证插件实例
	if loaderPlugin == nil {
		return nil, fmt.Errorf("loaded plugin instance is nil")
	}

	// 获取插件信息
	pluginInfo := loaderPlugin.GetInfo()
	if pluginInfo == nil {
		return nil, fmt.Errorf("plugin info is nil")
	}

	// 生成插件ID
	pluginID := m.generatePluginID(pluginPath, pluginType)
	if pluginInfo.ID != "" {
		pluginID = pluginInfo.ID
	}

	// 检查ID冲突
	if _, exists := m.plugins[pluginID]; exists {
		return nil, fmt.Errorf("plugin ID conflict: %s", pluginID)
	}

	// 创建插件上下文
	pluginCtx := m.createPluginContext(pluginPath)

	// 获取插件依赖
	dependencies := loaderPlugin.GetDependencies()
	if dependencies == nil {
		dependencies = []string{}
	}

	// 创建管理插件实例
	managedPlugin := &ManagedPlugin{
		ID:          pluginID,
		Path:        pluginPath,
		Type:        pluginType,
		Plugin:      loaderPlugin,
		Loader:      pluginLoader,
		State:       loader.PluginStateLoaded,
		LoadTime:    time.Now(),
		Dependencies: dependencies,
		Metadata:    make(map[string]interface{}),
		Context:     pluginCtx,
	}

	// 添加插件元数据
	managedPlugin.Metadata["loader_type"] = string(pluginType)
	managedPlugin.Metadata["capabilities"] = loaderPlugin.GetCapabilities()
	managedPlugin.Metadata["load_timestamp"] = time.Now().Unix()

	// 注册插件
	m.plugins[managedPlugin.ID] = managedPlugin

	// 发布加载事件
	m.publishEvent("plugin.loaded", map[string]interface{}{
		"plugin_id": managedPlugin.ID,
		"path":      pluginPath,
		"type":      string(pluginType),
		"info":      pluginInfo,
	})

	// 添加到监控
	if m.monitor != nil {
		m.monitor.AddPlugin(managedPlugin)
	}

	// 添加到健康检查
	if m.healthChecker != nil {
		m.healthChecker.AddPlugin(nil)
	}

	if m.logger != nil {
		m.logger.Info("Plugin loaded successfully",
			slog.String("plugin_id", managedPlugin.ID),
			slog.String("name", pluginInfo.Name),
			slog.String("version", pluginInfo.Version),
			slog.String("path", pluginPath),
			slog.String("type", string(pluginType)),
			slog.Int("dependencies", len(dependencies)),
		)
	}

	return managedPlugin, nil
}

// StartPlugin 启动插件
func (m *HybridPluginManager) StartPlugin(pluginID string) error {
	return m.StartPluginWithOptions(pluginID, &StartOptions{})
}

// StartOptions 插件启动选项
type StartOptions struct {
	ForceStart     bool          // 强制启动，忽略依赖检查
	RetryCount     int           // 重试次数
	RetryDelay     time.Duration // 重试延迟
	Timeout        time.Duration // 自定义超时时间
	SkipHealthCheck bool         // 跳过健康检查
	Hooks          *StartHooks   // 启动钩子
}

// StartHooks 启动钩子
type StartHooks struct {
	PreStart  func(pluginID string) error
	PostStart func(pluginID string) error
	OnError   func(pluginID string, err error)
}

// StopOptions 停止选项
type StopOptions struct {
	Timeout     time.Duration
	Force       bool
	Cleanup     bool
	Hooks       *StopHooks
	ForceStop       bool          // 强制停止，忽略依赖检查
	GracefulShutdown bool         // 优雅关闭
	SkipCleanup     bool          // 跳过资源清理
}

// StopHooks 停止钩子
type StopHooks struct {
	PreStop  func(ctx context.Context, plugin *ManagedPlugin) error
	PostStop func(ctx context.Context, plugin *ManagedPlugin) error
}

// UnloadOptions 卸载选项
type UnloadOptions struct {
	Timeout         time.Duration // 自定义超时时间
	ForceUnload     bool          // 强制卸载，忽略依赖检查
	GracefulShutdown bool         // 优雅关闭
	SkipCleanup     bool          // 跳过资源清理
	CascadeUnload   bool          // 级联卸载依赖插件
	Hooks           *UnloadHooks  // 卸载钩子
	RetryCount      int           // 重试次数
	RetryDelay      time.Duration // 重试延迟
}

// UnloadHooks 卸载钩子
type UnloadHooks struct {
	PreUnload  func(ctx context.Context, plugin *ManagedPlugin) error
	PostUnload func(ctx context.Context, plugin *ManagedPlugin) error
	OnError    func(pluginID string, err error)
	OnCleanup  func(ctx context.Context, plugin *ManagedPlugin) error
}

// PluginHooks 插件钩子函数
type PluginHooks struct {
	PreLoad   func(ctx context.Context, path string) error
	PostLoad  func(ctx context.Context, plugin *ManagedPlugin) error
	PreStart  func(ctx context.Context, plugin *ManagedPlugin) error
	PostStart func(ctx context.Context, plugin *ManagedPlugin) error
	PreStop   func(ctx context.Context, plugin *ManagedPlugin) error
	PostStop  func(ctx context.Context, plugin *ManagedPlugin) error
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	FailureThreshold int
	RecoveryTimeout  time.Duration
	FailureCount     int
	LastFailureTime  time.Time
	State           CircuitBreakerState
	mutex           sync.RWMutex
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// StartRequest 启动请求
type StartRequest struct {
	PluginID string
	Options  *StartOptions
	Result   chan error
}

// StopRequest 停止请求
type StopRequest struct {
	PluginID string
	Options  *StopOptions
	Result   chan error
}

// WorkerPool 工作池
type WorkerPool struct {
	workers   int
	taskQueue chan func()
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// StartPluginWithOptions 使用选项启动插件
func (m *HybridPluginManager) StartPluginWithOptions(pluginID string, options *StartOptions) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 验证插件ID
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	managedPlugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	managedPlugin.mutex.Lock()
	defer managedPlugin.mutex.Unlock()

	// 检查插件状态
	if managedPlugin.State != loader.PluginStateLoaded && managedPlugin.State != loader.PluginStateStopped {
		return fmt.Errorf("plugin is not in a startable state: %s (current: %s)", pluginID, managedPlugin.State.String())
	}

	// 检查依赖（除非强制启动）
	if !options.ForceStart {
		if err := m.checkDependencies(managedPlugin); err != nil {
			return fmt.Errorf("dependency check failed for plugin %s: %w", pluginID, err)
		}
	}

	// 执行预启动钩子
	if options.Hooks != nil && options.Hooks.PreStart != nil {
		if err := options.Hooks.PreStart(pluginID); err != nil {
			return fmt.Errorf("pre-start hook failed for plugin %s: %w", pluginID, err)
		}
	}

	// 设置超时时间
	timeout := m.config.StartTimeout
	if options.Timeout > 0 {
		timeout = options.Timeout
	}

	// 重试逻辑
	retryCount := options.RetryCount
	if retryCount <= 0 {
		retryCount = 1
	}
	retryDelay := options.RetryDelay
	if retryDelay <= 0 {
		retryDelay = time.Second
	}

	var lastErr error
	for attempt := 0; attempt < retryCount; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
			m.logger.Info("Retrying plugin start",
				slog.String("plugin_id", pluginID),
				slog.Int("attempt", attempt+1),
				slog.Int("max_attempts", retryCount),
			)
		}

		lastErr = m.doStartPlugin(managedPlugin, timeout)
		if lastErr == nil {
			break
		}

		// 执行错误钩子
		if options.Hooks != nil && options.Hooks.OnError != nil {
			options.Hooks.OnError(pluginID, lastErr)
		}
	}

	if lastErr != nil {
		return fmt.Errorf("failed to start plugin %s after %d attempts: %w", pluginID, retryCount, lastErr)
	}

	// 启动健康检查（除非跳过）
	if !options.SkipHealthCheck && m.healthChecker != nil {
		m.healthChecker.StartMonitoring()
	}

	// 执行后启动钩子
	if options.Hooks != nil && options.Hooks.PostStart != nil {
		if err := options.Hooks.PostStart(pluginID); err != nil {
			m.logger.Warn("Post-start hook failed",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

// doStartPlugin 执行实际的插件启动逻辑
func (m *HybridPluginManager) doStartPlugin(managedPlugin *ManagedPlugin, timeout time.Duration) error {
	// 创建启动上下文
	startCtx, startCancel := context.WithTimeout(m.ctx, timeout)
	defer startCancel()

	// 更新状态为初始化中
	managedPlugin.State = loader.PluginStateInitialized

	// 初始化插件
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("plugin initialization panicked: %v", r)
			}
		}()
		err = managedPlugin.Plugin.Initialize(managedPlugin.Context)
	}()

	select {
	case <-startCtx.Done():
		managedPlugin.State = loader.PluginStateError
		return fmt.Errorf("plugin initialization timeout: %w", startCtx.Err())
	case <-done:
		if err != nil {
			managedPlugin.State = loader.PluginStateError
			return fmt.Errorf("failed to initialize plugin: %w", err)
		}
	}

	// 启动插件
	done = make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("plugin start panicked: %v", r)
			}
		}()
		err = managedPlugin.Plugin.Start()
	}()

	select {
	case <-startCtx.Done():
		managedPlugin.State = loader.PluginStateError
		return fmt.Errorf("plugin start timeout: %w", startCtx.Err())
	case <-done:
		if err != nil {
			managedPlugin.State = loader.PluginStateError
			return fmt.Errorf("failed to start plugin: %w", err)
		}
	}

	// 更新状态和时间戳
	managedPlugin.State = loader.PluginStateRunning
	managedPlugin.StartTime = time.Now()

	// 注册插件服务
	if err := m.registerPluginServices(managedPlugin); err != nil {
		m.logger.Warn("Failed to register plugin services",
			slog.String("plugin_id", managedPlugin.ID),
			slog.String("error", err.Error()),
		)
	}

	// 发布启动事件
	m.publishEvent("plugin.started", map[string]interface{}{
		"plugin_id": managedPlugin.ID,
		"start_time": managedPlugin.StartTime,
		"info": managedPlugin.Plugin.GetInfo(),
	})

	if m.logger != nil {
		m.logger.Info("Plugin started successfully",
			slog.String("plugin_id", managedPlugin.ID),
			slog.String("name", managedPlugin.Plugin.GetInfo().Name),
			slog.Duration("startup_time", time.Since(managedPlugin.LoadTime)),
		)
	}

	return nil
}

// StopPlugin 停止插件
func (m *HybridPluginManager) StopPlugin(pluginID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 验证插件ID
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	managedPlugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	managedPlugin.mutex.Lock()
	defer managedPlugin.mutex.Unlock()

	// 检查插件状态
	if managedPlugin.State != loader.PluginStateRunning {
		return fmt.Errorf("plugin is not running: %s (current: %s)", pluginID, managedPlugin.State.String())
	}

	// 停止健康检查
	if m.healthChecker != nil {
		m.healthChecker.StopMonitoring()
	}

	// 创建停止上下文
	stopCtx, stopCancel := context.WithTimeout(m.ctx, m.config.StopTimeout)
	defer stopCancel()

	// 更新状态为停止中
	managedPlugin.State = loader.PluginStateStopping

	// 停止插件
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("plugin stop panicked: %v", r)
			}
		}()
		err = managedPlugin.Plugin.Stop()
	}()

	select {
	case <-stopCtx.Done():
		managedPlugin.State = loader.PluginStateError
		return fmt.Errorf("plugin stop timeout for %s: %w", pluginID, stopCtx.Err())
	case <-done:
		if err != nil {
			m.logger.Warn("Plugin stop returned error, but continuing cleanup",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 注销插件服务
	m.unregisterPluginServices(managedPlugin)

	// 清理插件资源
	if err := m.cleanupPluginResources(managedPlugin); err != nil {
		m.logger.Warn("Failed to cleanup plugin resources",
			slog.String("plugin_id", pluginID),
			slog.String("error", err.Error()),
		)
	}

	// 更新状态
	managedPlugin.State = loader.PluginStateStopped
	stopTime := time.Now()
	managedPlugin.StopTime = stopTime

	// 发布停止事件
	m.publishEvent("plugin.stopped", map[string]interface{}{
		"plugin_id": pluginID,
		"stop_time": stopTime,
		"runtime": stopTime.Sub(managedPlugin.StartTime),
		"info": managedPlugin.Plugin.GetInfo(),
	})

	if m.logger != nil {
		m.logger.Info("Plugin stopped successfully",
			slog.String("plugin_id", pluginID),
			slog.String("name", managedPlugin.Plugin.GetInfo().Name),
			slog.Duration("runtime", stopTime.Sub(managedPlugin.StartTime)),
		)
	}

	return nil
}

// StartPluginGroup 启动插件组
func (m *HybridPluginManager) StartPluginGroup(groupName string, options *StartOptions) error {
	m.mutex.RLock()
	plugins, exists := m.pluginGroups[groupName]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin group not found: %s", groupName)
	}

	// 按优先级排序
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Priority > plugins[j].Priority
	})

	// 依次启动插件
	for _, plugin := range plugins {
		if err := m.StartPluginWithOptions(plugin.ID, options); err != nil {
			return fmt.Errorf("failed to start plugin %s in group %s: %w", plugin.ID, groupName, err)
		}
	}

	return nil
}

// StopPluginGroup 停止插件组
func (m *HybridPluginManager) StopPluginGroup(groupName string, options *StopOptions) error {
	m.mutex.RLock()
	plugins, exists := m.pluginGroups[groupName]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin group not found: %s", groupName)
	}

	// 按优先级逆序排序（先停止低优先级的）
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Priority < plugins[j].Priority
	})

	// 依次停止插件
	for _, plugin := range plugins {
		if err := m.StopPluginWithOptions(plugin.ID, options); err != nil {
			return fmt.Errorf("failed to stop plugin %s in group %s: %w", plugin.ID, groupName, err)
		}
	}

	return nil
}

// BatchStartPlugins 批量启动插件
func (m *HybridPluginManager) BatchStartPlugins(pluginIDs []string, options *StartOptions) map[string]error {
	results := make(map[string]error)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, pluginID := range pluginIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			err := m.StartPluginWithOptions(id, options)
			mu.Lock()
			results[id] = err
			mu.Unlock()
		}(pluginID)
	}

	wg.Wait()
	return results
}

// BatchStopPlugins 批量停止插件
func (m *HybridPluginManager) BatchStopPlugins(pluginIDs []string, options *StopOptions) map[string]error {
	results := make(map[string]error)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, pluginID := range pluginIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			err := m.StopPluginWithOptions(id, options)
			mu.Lock()
			results[id] = err
			mu.Unlock()
		}(pluginID)
	}

	wg.Wait()
	return results
}

// AddPluginToGroup 将插件添加到组
func (m *HybridPluginManager) AddPluginToGroup(pluginID, groupName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	managedPlugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 更新插件的组信息
	managedPlugin.Group = groupName

	// 添加到组中
	if m.pluginGroups == nil {
		m.pluginGroups = make(map[string][]*ManagedPlugin)
	}

	// 检查是否已存在
	for _, plugin := range m.pluginGroups[groupName] {
		if plugin.ID == pluginID {
			return nil // 已存在
		}
	}

	m.pluginGroups[groupName] = append(m.pluginGroups[groupName], managedPlugin)
	return nil
}

// RemovePluginFromGroup 从组中移除插件
func (m *HybridPluginManager) RemovePluginFromGroup(pluginID, groupName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	managedPlugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 清除插件的组信息
	managedPlugin.Group = ""

	// 从组中移除
	plugins, exists := m.pluginGroups[groupName]
	if !exists {
		return fmt.Errorf("plugin group not found: %s", groupName)
	}

	for i, plugin := range plugins {
		if plugin.ID == pluginID {
			m.pluginGroups[groupName] = append(plugins[:i], plugins[i+1:]...)
			break
		}
	}

	return nil
}

// GetPluginGroupNames 获取所有插件组名称
func (m *HybridPluginManager) GetPluginGroupNames() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var names []string
	for groupName := range m.pluginGroups {
		names = append(names, groupName)
	}

	return names
}

// checkDependentPlugins 检查依赖此插件的其他插件
func (m *HybridPluginManager) checkDependentPlugins(managedPlugin *ManagedPlugin) error {
	var dependentPlugins []string

	for _, plugin := range m.plugins {
		for _, dep := range plugin.Dependencies {
			if dep == managedPlugin.ID {
				if plugin.State == loader.PluginStateRunning {
					dependentPlugins = append(dependentPlugins, plugin.ID)
				}
				break
			}
		}
	}

	if len(dependentPlugins) > 0 {
		return fmt.Errorf("cannot stop plugin %s: dependent plugins are still running: %v", managedPlugin.ID, dependentPlugins)
	}

	return nil
}

// UnloadPlugin 卸载插件（基础方法）
func (m *HybridPluginManager) UnloadPlugin(pluginID string) error {
	return m.UnloadPluginWithOptions(pluginID, &UnloadOptions{})
}

// UnloadPluginWithOptions 使用选项卸载插件
func (m *HybridPluginManager) UnloadPluginWithOptions(pluginID string, options *UnloadOptions) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 验证插件ID
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	// 验证选项
	if options == nil {
		options = &UnloadOptions{}
	}

	managedPlugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	managedPlugin.mutex.Lock()
	defer managedPlugin.mutex.Unlock()

	_ = managedPlugin.Plugin.GetInfo()

	// 检查插件状态是否可以卸载
	if !canUnloadFromState(managedPlugin.State) && !options.ForceUnload {
		return fmt.Errorf("plugin %s is not in a unloadable state: %s (use ForceUnload to override)", pluginID, managedPlugin.State.String())
	}

	// 检查依赖关系（除非强制卸载）
	if !options.ForceUnload {
		if err := m.checkDependentPluginsForUnload(managedPlugin); err != nil {
			return fmt.Errorf("dependency check failed for plugin %s: %w", pluginID, err)
		}
	}

	// 执行预卸载钩子
	if options.Hooks != nil && options.Hooks.PreUnload != nil {
		ctx := context.Background()
		if options.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, options.Timeout)
			defer cancel()
		}
		if err := options.Hooks.PreUnload(ctx, managedPlugin); err != nil {
			return fmt.Errorf("pre-unload hook failed for plugin %s: %w", pluginID, err)
		}
	}

	// 设置超时时间
	timeout := m.config.StopTimeout
	if options.Timeout > 0 {
		timeout = options.Timeout
	}

	// 重试逻辑
	retryCount := options.RetryCount
	if retryCount <= 0 {
		retryCount = 1
	}
	retryDelay := options.RetryDelay
	if retryDelay <= 0 {
		retryDelay = time.Second
	}

	var lastErr error
	for attempt := 0; attempt < retryCount; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
			m.logger.Info("Retrying plugin unload",
				slog.String("plugin_id", pluginID),
				slog.Int("attempt", attempt+1),
				slog.Int("max_attempts", retryCount),
			)
		}

		lastErr = m.doUnloadPlugin(managedPlugin, timeout, options)
		if lastErr == nil {
			break
		}

		// 执行错误钩子
		if options.Hooks != nil && options.Hooks.OnError != nil {
			options.Hooks.OnError(pluginID, lastErr)
		}
	}

	if lastErr != nil {
		// 尝试错误恢复
		if err := m.recoverFromUnloadError(managedPlugin, lastErr, options); err != nil {
			m.logger.Error("Failed to recover from unload error",
				slog.String("plugin_id", pluginID),
				slog.String("original_error", lastErr.Error()),
				slog.String("recovery_error", err.Error()),
			)
			return fmt.Errorf("failed to unload plugin %s after %d attempts and recovery failed: original error: %w, recovery error: %v", pluginID, retryCount, lastErr, err)
		}
		// 错误恢复成功，返回原始错误但标记为已恢复
		m.logger.Info("Successfully recovered from unload error",
			slog.String("plugin_id", pluginID),
			slog.String("original_error", lastErr.Error()),
		)
		return fmt.Errorf("plugin unload failed but recovered: %w", lastErr)
	}

	// 执行后卸载钩子
	if options.Hooks != nil && options.Hooks.PostUnload != nil {
		ctx := context.Background()
		if options.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, options.Timeout)
			defer cancel()
		}
		if err := options.Hooks.PostUnload(ctx, managedPlugin); err != nil {
			m.logger.Warn("Post-unload hook failed",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 级联卸载依赖插件
	if options.CascadeUnload {
		if err := m.cascadeUnloadDependents(pluginID, options); err != nil {
			m.logger.Warn("Cascade unload failed",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

// GetPlugin 获取插件
func (m *HybridPluginManager) GetPlugin(pluginID string) (*ManagedPlugin, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	managedPlugin, exists := m.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return managedPlugin, nil
}

// ListPlugins 列出所有插件
func (m *HybridPluginManager) ListPlugins() []*ManagedPlugin {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plugins := make([]*ManagedPlugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// GetPluginsByState 根据状态获取插件
func (m *HybridPluginManager) GetPluginsByState(state loader.PluginState) []*ManagedPlugin {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var plugins []*ManagedPlugin
	for _, plugin := range m.plugins {
		plugin.mutex.RLock()
		if plugin.State == state {
			plugins = append(plugins, plugin)
		}
		plugin.mutex.RUnlock()
	}

	return plugins
}

// GetPluginsByType 根据类型获取插件
func (m *HybridPluginManager) GetPluginsByType(pluginType loader.LoaderType) []*ManagedPlugin {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var plugins []*ManagedPlugin
	for _, plugin := range m.plugins {
		if plugin.Type == pluginType {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}

// Shutdown 关闭管理器
func (m *HybridPluginManager) Shutdown() error {
	m.logger.Info("Shutting down plugin manager")

	// 停止所有运行中的插件
	runningPlugins := m.GetPluginsByState(loader.PluginStateRunning)
	for _, plugin := range runningPlugins {
		if err := m.StopPlugin(plugin.ID); err != nil {
			m.logger.Warn("Failed to stop plugin during shutdown",
				slog.String("plugin_id", plugin.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 卸载所有插件
	allPlugins := m.ListPlugins()
	for _, plugin := range allPlugins {
		if err := m.UnloadPlugin(plugin.ID); err != nil {
			m.logger.Warn("Failed to unload plugin during shutdown",
				slog.String("plugin_id", plugin.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 停止监控组件
	if m.monitor != nil {
		m.monitor.Stop()
	}
	if m.healthChecker != nil {
		m.healthChecker.Stop()
	}

	// 取消上下文
	m.cancel()

	m.logger.Info("Plugin manager shutdown completed")
	return nil
}

// selectLoader 选择合适的插件加载器
func (m *HybridPluginManager) selectLoader(pluginType loader.LoaderType) loader.PluginLoader {
	switch pluginType {
	case loader.LoaderTypeDynamic:
		return m.dynamicLoader
	case loader.LoaderTypeRPC:
		return m.rpcLoader
	case loader.LoaderTypeWASM:
		return m.wasmLoader
	case loader.LoaderTypeHotReload:
		return m.hotReloadLoader
	default:
		return nil
	}
}

// generatePluginID 生成插件ID
func (m *HybridPluginManager) generatePluginID(pluginPath string, pluginType loader.LoaderType) string {
	return fmt.Sprintf("%s_%s_%d", pluginType, pluginPath, time.Now().UnixNano())
}

// createPluginContext 创建插件上下文
func (m *HybridPluginManager) createPluginContext(pluginPath string) loader.PluginContext {
	// 创建基础配置
	config := &BasePluginConfig{
		ID:           m.generatePluginID(pluginPath, loader.LoaderTypeDynamic),
		Name:         filepath.Base(pluginPath),
		Version:      "1.0.0",
		Type:         PluginTypeDynamicLibrary,
		Enabled:      true,
		Priority:     PluginPriorityNormal,
		CustomConfig: make(map[string]interface{}),
	}
	
	// 创建插件上下文
	ctx := NewPluginContext(m.ctx, config)
	return &PluginContextAdapter{ctx: ctx}
}

// checkDependencies 检查插件依赖
func (m *HybridPluginManager) checkDependencies(plugin *ManagedPlugin) error {
	for _, depID := range plugin.Dependencies {
		dep, exists := m.plugins[depID]
		if !exists {
			return fmt.Errorf("dependency not found: %s", depID)
		}
		dep.mutex.RLock()
		if dep.State != loader.PluginStateRunning {
			dep.mutex.RUnlock()
			return fmt.Errorf("dependency not running: %s (state: %s)", depID, dep.State.String())
		}
		dep.mutex.RUnlock()
	}
	return nil
}

// publishEvent 发布事件
func (m *HybridPluginManager) publishEvent(eventType string, data map[string]interface{}) {
	if m.eventBus != nil {
		if err := m.eventBus.Publish(eventType, data); err != nil {
			m.logger.Warn("Failed to publish event",
				slog.String("event_type", eventType),
				slog.String("error", err.Error()),
			)
		}
	}
}

// publishPluginEvent 发布插件相关事件
func (m *HybridPluginManager) publishPluginEvent(eventType string, plugin *ManagedPlugin, additionalData map[string]interface{}) {
	pluginInfo := plugin.Plugin.GetInfo()
	data := map[string]interface{}{
		"plugin_id":   plugin.ID,
		"plugin_path": plugin.Path,
		"plugin_type": plugin.Type.String(),
		"plugin_state": plugin.State.String(),
		"timestamp":   time.Now(),
	}

	if pluginInfo != nil {
		data["plugin_name"] = pluginInfo.Name
		data["plugin_version"] = pluginInfo.Version
		data["plugin_author"] = pluginInfo.Author
	}

	// 合并额外数据
	for k, v := range additionalData {
		data[k] = v
	}

	m.publishEvent(eventType, data)
}

// publishUnloadEvent 发布卸载事件
func (m *HybridPluginManager) publishUnloadEvent(eventType string, plugin *ManagedPlugin, success bool, errorMsg string) {
	additionalData := map[string]interface{}{
		"success": success,
	}

	if errorMsg != "" {
		additionalData["error"] = errorMsg
	}

	m.publishPluginEvent(eventType, plugin, additionalData)
}

// publishCleanupEvent 发布清理事件
func (m *HybridPluginManager) publishCleanupEvent(eventType string, plugin *ManagedPlugin, cleanupType string, success bool, errorMsg string) {
	additionalData := map[string]interface{}{
		"cleanup_type": cleanupType,
		"success":      success,
	}

	if errorMsg != "" {
		additionalData["error"] = errorMsg
	}

	m.publishPluginEvent(eventType, plugin, additionalData)
}

// publishDependencyEvent 发布依赖事件
func (m *HybridPluginManager) publishDependencyEvent(eventType string, targetPlugin *ManagedPlugin, dependentPlugins []*ManagedPlugin) {
	dependentIDs := make([]string, len(dependentPlugins))
	dependentNames := make([]string, len(dependentPlugins))

	for i, plugin := range dependentPlugins {
		dependentIDs[i] = plugin.ID
		pluginInfo := plugin.Plugin.GetInfo()
		if pluginInfo != nil {
			dependentNames[i] = pluginInfo.Name
		} else {
			dependentNames[i] = plugin.ID
		}
	}

	additionalData := map[string]interface{}{
		"dependent_plugin_ids":   dependentIDs,
		"dependent_plugin_names": dependentNames,
		"dependent_count":        len(dependentPlugins),
	}

	m.publishPluginEvent(eventType, targetPlugin, additionalData)
}

// secureCleanupPlugin 安全清理插件
func (m *HybridPluginManager) secureCleanupPlugin(managedPlugin *ManagedPlugin, ctx context.Context) error {
	pluginID := managedPlugin.ID
	m.logger.Info("Starting secure plugin cleanup",
		slog.String("plugin_id", pluginID),
	)

	// 1. 清理敏感数据
	if err := m.cleanupSensitiveData(managedPlugin, ctx); err != nil {
		m.logger.Warn("Failed to cleanup sensitive data",
			slog.String("plugin_id", pluginID),
			slog.String("error", err.Error()),
		)
		// 继续清理，不因敏感数据清理失败而停止
	}

	// 2. 清理安全上下文
	if err := m.cleanupSecurityContext(managedPlugin, ctx); err != nil {
		m.logger.Warn("Failed to cleanup security context",
			slog.String("plugin_id", pluginID),
			slog.String("error", err.Error()),
		)
	}

	// 3. 清理权限和访问控制
	if err := m.cleanupPermissions(managedPlugin, ctx); err != nil {
		m.logger.Warn("Failed to cleanup permissions",
			slog.String("plugin_id", pluginID),
			slog.String("error", err.Error()),
		)
	}

	// 4. 清理加密密钥
	if err := m.cleanupEncryptionKeys(managedPlugin, ctx); err != nil {
		m.logger.Warn("Failed to cleanup encryption keys",
			slog.String("plugin_id", pluginID),
			slog.String("error", err.Error()),
		)
	}

	// 5. 清理沙箱环境
	if err := m.cleanupSandboxEnvironment(managedPlugin, ctx); err != nil {
		m.logger.Warn("Failed to cleanup sandbox environment",
			slog.String("plugin_id", pluginID),
			slog.String("error", err.Error()),
		)
	}

	m.logger.Info("Secure plugin cleanup completed",
		slog.String("plugin_id", pluginID),
	)

	return nil
}

// cleanupSensitiveData 清理敏感数据
func (m *HybridPluginManager) cleanupSensitiveData(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// 清理插件可能存储的敏感数据
	if sensitiveDataCleaner, ok := managedPlugin.Plugin.(interface{ CleanupSensitiveData() error }); ok {
		if err := sensitiveDataCleaner.CleanupSensitiveData(); err != nil {
			return fmt.Errorf("failed to cleanup plugin sensitive data: %w", err)
		}
	}

	// 清理插件元数据中的敏感信息
	if managedPlugin.Metadata != nil {
		sensitiveKeys := []string{"password", "token", "key", "secret", "credential"}
		for _, key := range sensitiveKeys {
			for metaKey := range managedPlugin.Metadata {
				if strings.Contains(strings.ToLower(metaKey), key) {
					delete(managedPlugin.Metadata, metaKey)
				}
			}
		}
	}

	return nil
}

// cleanupSecurityContext 清理安全上下文
func (m *HybridPluginManager) cleanupSecurityContext(managedPlugin *ManagedPlugin, ctx context.Context) error {
	if m.securityManager != nil {
		// 清理插件的安全上下文 - 如果SecurityManager支持的话
		m.logger.Debug("Cleaning up security context", slog.String("plugin_id", managedPlugin.ID))
	}
	return nil
}

// cleanupPermissions 清理权限和访问控制
func (m *HybridPluginManager) cleanupPermissions(managedPlugin *ManagedPlugin, ctx context.Context) error {
	if m.securityManager != nil {
		// 撤销插件的所有权限 - 如果SecurityManager支持的话
		m.logger.Debug("Revoking permissions", slog.String("plugin_id", managedPlugin.ID))
	}
	return nil
}

// cleanupEncryptionKeys 清理加密密钥
func (m *HybridPluginManager) cleanupEncryptionKeys(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// 清理插件相关的加密密钥
	if keyCleaner, ok := managedPlugin.Plugin.(interface{ CleanupEncryptionKeys() error }); ok {
		if err := keyCleaner.CleanupEncryptionKeys(); err != nil {
			return fmt.Errorf("failed to cleanup encryption keys: %w", err)
		}
	}
	return nil
}

// cleanupSandboxEnvironment 清理沙箱环境
func (m *HybridPluginManager) cleanupSandboxEnvironment(managedPlugin *ManagedPlugin, ctx context.Context) error {
	// 对于WASM插件，清理沙箱环境
	if managedPlugin.Type == loader.LoaderTypeWASM {
		if m.wasmLoader != nil {
			// 清理WASM沙箱环境 - 使用UnloadPlugin作为清理方法
			if err := m.wasmLoader.UnloadPlugin(ctx, managedPlugin.ID); err != nil {
				return fmt.Errorf("failed to cleanup WASM sandbox: %w", err)
			}
		}
	}

	// 清理其他类型插件的沙箱环境
	if sandboxCleaner, ok := managedPlugin.Plugin.(interface{ CleanupSandbox() error }); ok {
		if err := sandboxCleaner.CleanupSandbox(); err != nil {
			return fmt.Errorf("failed to cleanup plugin sandbox: %w", err)
		}
	}

	return nil
}

// recoverFromUnloadError 从卸载错误中恢复
func (m *HybridPluginManager) recoverFromUnloadError(managedPlugin *ManagedPlugin, originalError error, options *UnloadOptions) error {
	pluginID := managedPlugin.ID
	m.logger.Warn("Attempting to recover from unload error",
		slog.String("plugin_id", pluginID),
		slog.String("original_error", originalError.Error()),
	)

	// 发布错误恢复开始事件
	m.publishUnloadEvent("plugin.unload.recovery.started", managedPlugin, false, originalError.Error())

	// 1. 尝试强制停止插件
	if managedPlugin.State == loader.PluginStateRunning {
		if err := m.forceStopPlugin(managedPlugin); err != nil {
			m.logger.Warn("Failed to force stop plugin during recovery",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 2. 标记插件为错误状态
	managedPlugin.State = loader.PluginStateError
	managedPlugin.LastError = originalError
	managedPlugin.FailureCount++

	// 3. 强制清理资源
	if err := m.forceCleanupResources(managedPlugin); err != nil {
		m.logger.Warn("Failed to force cleanup resources during recovery",
			slog.String("plugin_id", pluginID),
			slog.String("error", err.Error()),
		)
	}

	// 4. 从管理器中移除插件
	delete(m.plugins, pluginID)

	// 5. 发布错误恢复完成事件
	m.publishUnloadEvent("plugin.unload.recovery.completed", managedPlugin, true, "")

	m.logger.Info("Recovery from unload error completed",
		slog.String("plugin_id", pluginID),
	)

	return nil
}

// forceStopPlugin 强制停止插件
func (m *HybridPluginManager) forceStopPlugin(managedPlugin *ManagedPlugin) error {
	pluginID := managedPlugin.ID
	m.logger.Info("Force stopping plugin",
		slog.String("plugin_id", pluginID),
	)

	// 创建短超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试正常停止
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("plugin stop panicked: %v", r)
			}
		}()
		done <- managedPlugin.Plugin.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			m.logger.Warn("Plugin stop returned error during force stop",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	case <-ctx.Done():
		m.logger.Warn("Plugin stop timeout during force stop",
			slog.String("plugin_id", pluginID),
		)
	}

	// 无论如何都更新状态
	managedPlugin.State = loader.PluginStateStopped
	return nil
}

// forceCleanupResources 强制清理资源
func (m *HybridPluginManager) forceCleanupResources(managedPlugin *ManagedPlugin) error {
	pluginID := managedPlugin.ID
	m.logger.Info("Force cleaning up plugin resources",
		slog.String("plugin_id", pluginID),
	)

	// 创建短超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 强制清理各种资源
	var errors []error

	// 1. 强制清理插件实例
	if managedPlugin.Plugin != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					errors = append(errors, fmt.Errorf("plugin cleanup panicked: %v", r))
				}
			}()
			if err := managedPlugin.Plugin.Cleanup(); err != nil {
				errors = append(errors, fmt.Errorf("plugin cleanup failed: %w", err))
			}
		}()
	}

	// 2. 强制清理高级资源
	if err := m.cleanupPluginResourcesAdvanced(managedPlugin, ctx); err != nil {
		errors = append(errors, fmt.Errorf("advanced cleanup failed: %w", err))
	}

	// 3. 强制安全清理
	if err := m.secureCleanupPlugin(managedPlugin, ctx); err != nil {
		errors = append(errors, fmt.Errorf("secure cleanup failed: %w", err))
	}

	// 4. 注销服务
	m.unregisterPluginServices(managedPlugin)

	// 5. 停止健康检查
	if m.healthChecker != nil {
		m.healthChecker.StopMonitoring()
	}

	// 6. 从监控中移除
	if m.monitor != nil {
		m.monitor.RemovePlugin(managedPlugin.ID)
	}

	if len(errors) > 0 {
		errorMsgs := make([]string, len(errors))
		for i, err := range errors {
			errorMsgs[i] = err.Error()
		}
		m.logger.Warn("Force cleanup completed with errors",
			slog.String("plugin_id", pluginID),
			slog.Int("error_count", len(errors)),
		)
		return fmt.Errorf("force cleanup errors: %s", strings.Join(errorMsgs, "; "))
	}

	m.logger.Info("Force cleanup completed successfully",
		slog.String("plugin_id", pluginID),
	)

	return nil
}

// registerPluginServices 注册插件服务
func (m *HybridPluginManager) registerPluginServices(managedPlugin *ManagedPlugin) error {
	if m.serviceRegistry == nil {
		return nil
	}

	pluginInfo := managedPlugin.Plugin.GetInfo()
	serviceName := fmt.Sprintf("plugin.%s", pluginInfo.Name)
	
	// 注册插件实例
	if err := m.serviceRegistry.RegisterService(serviceName, managedPlugin.Plugin); err != nil {
		return fmt.Errorf("failed to register plugin service %s: %w", serviceName, err)
	}

	// 注册插件上下文
	contextServiceName := fmt.Sprintf("plugin.%s.context", pluginInfo.Name)
	if err := m.serviceRegistry.RegisterService(contextServiceName, managedPlugin.Context); err != nil {
		m.logger.Warn("Failed to register plugin context service",
			slog.String("service", contextServiceName),
			slog.String("error", err.Error()),
		)
	}

	return nil
}

// unregisterPluginServices 注销插件服务
func (m *HybridPluginManager) unregisterPluginServices(managedPlugin *ManagedPlugin) {
	if m.serviceRegistry == nil {
		return
	}

	pluginInfo := managedPlugin.Plugin.GetInfo()
	serviceName := fmt.Sprintf("plugin.%s", pluginInfo.Name)
	contextServiceName := fmt.Sprintf("plugin.%s.context", pluginInfo.Name)

	// 注销服务
	m.serviceRegistry.UnregisterService(serviceName)
	m.serviceRegistry.UnregisterService(contextServiceName)
}

// cleanupPluginResources 清理插件资源（基础版本）
func (m *HybridPluginManager) cleanupPluginResources(managedPlugin *ManagedPlugin) error {
	return m.cleanupPluginResourcesAdvanced(managedPlugin, context.Background())
}

// cleanupPluginResourcesAdvanced 高级插件资源清理
func (m *HybridPluginManager) cleanupPluginResourcesAdvanced(managedPlugin *ManagedPlugin, ctx context.Context) error {
	var errors []error
	pluginID := managedPlugin.ID

	m.logger.Info("Starting advanced plugin resource cleanup",
		slog.String("plugin_id", pluginID),
	)

	// 1. 清理插件上下文资源
	if err := m.cleanupPluginContext(managedPlugin, ctx); err != nil {
		errors = append(errors, fmt.Errorf("context cleanup failed: %w", err))
	}

	// 2. 清理事件监听器
	if err := m.cleanupEventListeners(managedPlugin, ctx); err != nil {
		errors = append(errors, fmt.Errorf("event listeners cleanup failed: %w", err))
	}

	// 3. 清理临时文件和缓存
	if err := m.cleanupTemporaryFiles(managedPlugin, ctx); err != nil {
		errors = append(errors, fmt.Errorf("temporary files cleanup failed: %w", err))
	}

	// 4. 清理网络连接
	if err := m.cleanupNetworkConnections(managedPlugin, ctx); err != nil {
		errors = append(errors, fmt.Errorf("network connections cleanup failed: %w", err))
	}

	// 5. 清理内存资源
	if err := m.cleanupMemoryResources(managedPlugin, ctx); err != nil {
		errors = append(errors, fmt.Errorf("memory resources cleanup failed: %w", err))
	}

	// 6. 清理文件句柄
	if err := m.cleanupFileHandles(managedPlugin, ctx); err != nil {
		errors = append(errors, fmt.Errorf("file handles cleanup failed: %w", err))
	}

	// 7. 清理系统资源
	if err := m.cleanupSystemResources(managedPlugin, ctx); err != nil {
		errors = append(errors, fmt.Errorf("system resources cleanup failed: %w", err))
	}

	// 8. 清理插件特定资源
	if err := m.cleanupPluginSpecificResources(managedPlugin, ctx); err != nil {
		errors = append(errors, fmt.Errorf("plugin specific resources cleanup failed: %w", err))
	}

	// 合并错误
	if len(errors) > 0 {
		errorMsgs := make([]string, len(errors))
		for i, err := range errors {
			errorMsgs[i] = err.Error()
		}
		m.logger.Warn("Plugin resource cleanup completed with errors",
			slog.String("plugin_id", pluginID),
			slog.Int("error_count", len(errors)),
		)
		return fmt.Errorf("cleanup errors: %s", strings.Join(errorMsgs, "; "))
	}

	m.logger.Info("Plugin resource cleanup completed successfully",
		slog.String("plugin_id", pluginID),
	)
	return nil
}

// validateLoaders 验证所有加载器是否正确初始化
func (m *HybridPluginManager) validateLoaders() error {
	if m.dynamicLoader == nil {
		return fmt.Errorf("dynamic loader is not initialized")
	}
	if m.rpcLoader == nil {
		return fmt.Errorf("RPC loader is not initialized")
	}
	if m.wasmLoader == nil {
		return fmt.Errorf("WASM loader is not initialized")
	}
	if m.config.EnableHotReload && m.hotReloadLoader == nil {
		return fmt.Errorf("hot reload loader is not initialized")
	}
	return nil
}

// GetPluginInfo 获取插件信息
func (m *HybridPluginManager) GetPluginInfo(pluginID string) (*loader.PluginInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	managedPlugin, exists := m.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	if managedPlugin.Plugin == nil {
		return nil, fmt.Errorf("plugin instance is nil")
	}

	return managedPlugin.Plugin.GetInfo(), nil
}

// GetPluginState 获取插件状态
func (m *HybridPluginManager) GetPluginState(pluginID string) (loader.PluginState, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	managedPlugin, exists := m.plugins[pluginID]
	if !exists {
		return loader.PluginStateUnknown, fmt.Errorf("plugin not found: %s", pluginID)
	}

	managedPlugin.mutex.RLock()
	defer managedPlugin.mutex.RUnlock()
	return managedPlugin.State, nil
}

// StopPluginWithOptions 使用选项停止插件
func (m *HybridPluginManager) StopPluginWithOptions(pluginID string, options *StopOptions) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 验证插件ID
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	managedPlugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	managedPlugin.mutex.Lock()
	defer managedPlugin.mutex.Unlock()

	// 检查插件状态
	if managedPlugin.State != loader.PluginStateRunning {
		return fmt.Errorf("plugin is not running: %s (current: %s)", pluginID, managedPlugin.State.String())
	}

	// 检查依赖（除非强制停止）
	if !options.ForceStop {
		if err := m.checkDependentPlugins(managedPlugin); err != nil {
			return fmt.Errorf("dependency check failed for plugin %s: %w", pluginID, err)
		}
	}

	// 执行预停止钩子
	if options.Hooks != nil && options.Hooks.PreStop != nil {
		ctx := context.Background()
		if options.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, options.Timeout)
			defer cancel()
		}
		if err := options.Hooks.PreStop(ctx, managedPlugin); err != nil {
			return fmt.Errorf("pre-stop hook failed for plugin %s: %w", pluginID, err)
		}
	}

	// 停止健康检查
	if m.healthChecker != nil {
		m.healthChecker.StopMonitoring()
	}

	// 设置超时时间
	timeout := m.config.StopTimeout
	if options.Timeout > 0 {
		timeout = options.Timeout
	}

	// 创建停止上下文
	stopCtx, stopCancel := context.WithTimeout(m.ctx, timeout)
	defer stopCancel()

	// 更新状态为停止中
	managedPlugin.State = loader.PluginStateStopping

	// 停止插件
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("plugin stop panicked: %v", r)
			}
		}()
		err = managedPlugin.Plugin.Stop()
	}()

	select {
	case <-stopCtx.Done():
		if options.Force {
			m.logger.Warn("Plugin stop timeout, forcing cleanup",
				slog.String("plugin_id", pluginID),
			)
		} else {
			managedPlugin.State = loader.PluginStateError
			return fmt.Errorf("plugin stop timeout for %s: %w", pluginID, stopCtx.Err())
		}
	case <-done:
		if err != nil {
			m.logger.Warn("Plugin stop returned error, but continuing cleanup",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 注销插件服务
	m.unregisterPluginServices(managedPlugin)

	// 清理插件资源（除非跳过）
	if !options.SkipCleanup {
		if err := m.cleanupPluginResources(managedPlugin); err != nil {
			m.logger.Warn("Failed to cleanup plugin resources",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 更新状态
	managedPlugin.State = loader.PluginStateStopped
	stopTime := time.Now()
	managedPlugin.StopTime = stopTime

	// 执行后停止钩子
	if options.Hooks != nil && options.Hooks.PostStop != nil {
		ctx := context.Background()
		if options.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, options.Timeout)
			defer cancel()
		}
		if err := options.Hooks.PostStop(ctx, managedPlugin); err != nil {
			m.logger.Warn("Post-stop hook failed",
				slog.String("plugin_id", pluginID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 发布停止事件
	m.publishEvent("plugin.stopped", map[string]interface{}{
		"plugin_id": pluginID,
		"stop_time": stopTime,
		"runtime": stopTime.Sub(managedPlugin.StartTime),
		"info": managedPlugin.Plugin.GetInfo(),
	})

	if m.logger != nil {
		m.logger.Info("Plugin stopped successfully",
			slog.String("plugin_id", pluginID),
			slog.String("name", managedPlugin.Plugin.GetInfo().Name),
			slog.Duration("runtime", stopTime.Sub(managedPlugin.StartTime)),
		)
	}

	return nil
}

// processStartQueue 处理启动队列
func (m *HybridPluginManager) processStartQueue() {
	for {
		select {
		case req := <-m.startQueue:
			err := m.StartPluginWithOptions(req.PluginID, req.Options)
			req.Result <- err
		case <-m.ctx.Done():
			return
		}
	}
}

// processStopQueue 处理停止队列
func (m *HybridPluginManager) processStopQueue() {
	for {
		select {
		case req := <-m.stopQueue:
			err := m.StopPluginWithOptions(req.PluginID, req.Options)
			req.Result <- err
		case <-m.ctx.Done():
			return
		}
	}
}

// StartPluginAsync 异步启动插件
func (m *HybridPluginManager) StartPluginAsync(pluginID string, options *StartOptions) <-chan error {
	req := &StartRequest{
		PluginID: pluginID,
		Options:  options,
		Result:   make(chan error, 1),
	}

	select {
	case m.startQueue <- req:
		return req.Result
	default:
		// 队列满了，直接返回错误
		result := make(chan error, 1)
		result <- fmt.Errorf("start queue is full")
		return result
	}
}

// StopPluginAsync 异步停止插件
func (m *HybridPluginManager) StopPluginAsync(pluginID string, options *StopOptions) <-chan error {
	req := &StopRequest{
		PluginID: pluginID,
		Options:  options,
		Result:   make(chan error, 1),
	}

	select {
	case m.stopQueue <- req:
		return req.Result
	default:
		// 队列满了，直接返回错误
		result := make(chan error, 1)
		result <- fmt.Errorf("stop queue is full")
		return result
	}
}

// GetPluginGroups 获取所有插件组
func (m *HybridPluginManager) GetPluginGroups() map[string][]*ManagedPlugin {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 创建副本以避免并发问题
	groups := make(map[string][]*ManagedPlugin)
	for name, plugins := range m.pluginGroups {
		groupCopy := make([]*ManagedPlugin, len(plugins))
		copy(groupCopy, plugins)
		groups[name] = groupCopy
	}

	return groups
}

// GetRunningPlugins 获取所有运行中的插件
func (m *HybridPluginManager) GetRunningPlugins() []*ManagedPlugin {
	return m.GetPluginsByState(loader.PluginStateRunning)
}

// GetStoppedPlugins 获取所有已停止的插件
func (m *HybridPluginManager) GetStoppedPlugins() []*ManagedPlugin {
	return m.GetPluginsByState(loader.PluginStateStopped)
}

// GetErrorPlugins 获取所有错误状态的插件
func (m *HybridPluginManager) GetErrorPlugins() []*ManagedPlugin {
	return m.GetPluginsByState(loader.PluginStateError)
}

// GetPluginStatistics 获取插件统计信息
func (m *HybridPluginManager) GetPluginStatistics() map[string]int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]int)
	for _, plugin := range m.plugins {
		plugin.mutex.RLock()
		stateStr := plugin.State.String()
		stats[stateStr]++
		plugin.mutex.RUnlock()
	}

	stats["total"] = len(m.plugins)
	return stats
}