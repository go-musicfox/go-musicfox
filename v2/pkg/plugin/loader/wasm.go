package loader

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v14"
)



// PluginType is defined in types.go







// WASMPluginLoader WebAssembly 插件加载器
// 负责加载和管理 WebAssembly 插件，提供安全的沙箱执行环境
type WASMPluginLoader struct {
	// loadedModules 已加载的 WASM 模块映射表
	loadedModules map[string]*WASMModule
	
	// mutex 读写锁，保护并发访问
	mutex sync.RWMutex
	
	// engine WASM 引擎实例
	engine *wasmtime.Engine
	
	// store WASM 存储实例
	store *wasmtime.Store
	
	// config 加载器配置
	config *WASMLoaderConfig
	
	// logger 日志记录器
	logger *slog.Logger
	
	// securityManager 安全管理器
	securityManager SecurityManager
	
	// resourceMonitor 资源监控器
	resourceMonitor *ResourceMonitor
	
	// ctx 上下文
	ctx context.Context
	
	// cancel 取消函数
	cancel context.CancelFunc
}

// WASMModule 已加载的 WASM 模块信息
type WASMModule struct {
	// ID 插件唯一标识符
	ID string
	
	// Path 插件文件路径
	Path string
	
	// Module WASM 模块实例
	Module *wasmtime.Module
	
	// Instance WASM 实例
	Instance *wasmtime.Instance
	
	// Store WASM 存储
	Store *wasmtime.Store
	
	// PluginInstance 插件实例包装器
	PluginInstance Plugin
	
	// Exports 导出函数映射
	Exports map[string]*wasmtime.Func
	
	// Memory WASM 内存
	Memory *wasmtime.Memory
	
	// State 插件状态
	State PluginState
	
	// LoadTime 加载时间
	LoadTime time.Time
	
	// LastAccess 最后访问时间
	LastAccess time.Time
	
	// ResourceUsage 资源使用情况
	ResourceUsage *ResourceUsage
	
	// Metadata 元数据
	Metadata map[string]interface{}
	
	// sandbox 沙箱环境
	sandbox *WASMSandbox
	
	LoadedAt time.Time
}

// WASMLoaderConfig WASM 加载器配置
type WASMLoaderConfig struct {
	// MaxModules 最大模块数量
	MaxModules int
	
	// LoadTimeout 加载超时时间
	LoadTimeout time.Duration
	
	// ExecutionTimeout 执行超时时间
	ExecutionTimeout time.Duration
	
	// MaxMemorySize 最大内存大小（字节）
	MaxMemorySize uint64
	
	// MaxStackSize 最大栈大小（字节）
	MaxStackSize uint64
	
	// EnableSandbox 是否启用沙箱
	EnableSandbox bool
	
	// AllowedPaths 允许的插件路径列表
	AllowedPaths []string
	
	// ResourceLimits 资源限制
	ResourceLimits *ResourceLimits
	
	// SecurityPolicy 安全策略
	SecurityPolicy *SecurityPolicy
	
	// EnableOptimization 是否启用优化
	EnableOptimization bool
	
	// EnableDebug 是否启用调试
	EnableDebug bool
}

// WASMSandbox WASM 沙箱环境
type WASMSandbox struct {
	// limits 资源限制
	limits *SandboxLimits
	
	// monitor 监控器
	monitor *SandboxMonitor
	
	// policy 安全策略
	policy *SecurityPolicy
	
	// startTime 启动时间
	startTime time.Time
	
	// active 是否激活
	active bool
}

// SandboxLimits 沙箱资源限制
type SandboxLimits struct {
	// MaxMemory 最大内存（字节）
	MaxMemory uint64
	
	// MaxCPUTime CPU 时间限制
	MaxCPUTime time.Duration
	
	// MaxExecutionTime 最大执行时间
	MaxExecutionTime time.Duration
	
	// MaxFileSize 最大文件大小
	MaxFileSize uint64
	
	// MaxNetworkConnections 最大网络连接数
	MaxNetworkConnections int
}

// SandboxMonitor 沙箱监控器
type SandboxMonitor struct {
	// memoryUsage 内存使用量
	memoryUsage uint64
	
	// cpuTime CPU 时间
	cpuTime time.Duration
	
	// executionTime 执行时间
	executionTime time.Duration
	
	// networkConnections 网络连接数
	networkConnections int
	
	// lastUpdate 最后更新时间
	lastUpdate time.Time
}

// SecurityPolicy 安全策略
type SecurityPolicy struct {
	// AllowFileAccess 是否允许文件访问
	AllowFileAccess bool
	
	// AllowNetworkAccess 是否允许网络访问
	AllowNetworkAccess bool
	
	// AllowSystemCalls 是否允许系统调用
	AllowSystemCalls bool
	
	// AllowedDomains 允许的域名列表
	AllowedDomains []string
	
	// BlockedFunctions 禁用的函数列表
	BlockedFunctions []string
}



// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	// modules 监控的模块列表
	modules map[string]*WASMModule
	
	// mutex 锁
	mutex sync.RWMutex
	
	// ticker 定时器
	ticker *time.Ticker
	
	// stopCh 停止通道
	stopCh chan struct{}
	
	// logger 日志记录器
	logger *slog.Logger
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	// MemoryUsed 已使用内存
	MemoryUsed uint64
	
	// CPUTime CPU 时间
	CPUTime time.Duration
	
	// ExecutionCount 执行次数
	ExecutionCount int64
	
	// LastExecution 最后执行时间
	LastExecution time.Time
	
	// ErrorCount 错误次数
	ErrorCount int64
}

// NewWASMPluginLoader 创建新的 WASM 插件加载器
func NewWASMPluginLoader(securityManager SecurityManager, logger *slog.Logger) *WASMPluginLoader {
	// 验证必需参数 - 使用反射检查接口是否为nil
	if securityManager == nil || logger == nil {
		return nil
	}
	
	// 检查接口是否包含nil值
	if reflect.ValueOf(securityManager).IsNil() {
		return nil
	}
	
	config := &WASMLoaderConfig{
		MaxModules:         10,
		LoadTimeout:        30 * time.Second,
		ExecutionTimeout:   5 * time.Second,
		MaxMemorySize:      64 * 1024 * 1024, // 64MB
		MaxStackSize:       1024 * 1024,      // 1MB
		EnableSandbox:      true,
		EnableOptimization: true,
		EnableDebug:        false,
		ResourceLimits: &ResourceLimits{
			MaxMemoryMB:      64,                    // 64MB
			MaxCPUPercent:    50.0,                  // 50% CPU
			MaxGoroutines:    100,                   // 100个协程
			MaxFileHandles:   50,                    // 50个文件句柄
			MaxNetworkConn:   10,                    // 10个网络连接
			ExecutionTimeout: time.Second * 30,      // 30秒执行超时
			IdleTimeout:      time.Minute * 5,       // 5分钟空闲超时
			StartupTimeout:   time.Second * 10,      // 10秒启动超时
			ShutdownTimeout:  time.Second * 5,       // 5秒关闭超时
			Enabled:          true,                  // 启用限制
			EnforceMode:      EnforceModeLimit, // 限制模式
		},
		SecurityPolicy: &SecurityPolicy{
			AllowFileAccess:    false,
			AllowNetworkAccess: false,
			AllowSystemCalls:   false,
		},
	}
	
	return NewWASMPluginLoaderWithConfig(config, securityManager, logger)
}

// NewWASMPluginLoaderWithConfig 使用指定配置创建 WASM 插件加载器
func NewWASMPluginLoaderWithConfig(config *WASMLoaderConfig, securityManager SecurityManager, logger *slog.Logger) *WASMPluginLoader {
	// 验证必需参数
	if config == nil || securityManager == nil || logger == nil {
		return nil
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// 创建 WASM 引擎配置
	engineConfig := wasmtime.NewConfig()
	engineConfig.SetConsumeFuel(true) // 启用燃料消耗以限制执行时间
	if config.EnableOptimization {
		// engineConfig.SetOptLevel(wasmtime.OptLevelSpeed) // 该方法在当前版本中不可用
	}
	if config.EnableDebug {
		engineConfig.SetDebugInfo(true)
	}
	
	// 创建引擎和存储
	engine := wasmtime.NewEngineWithConfig(engineConfig)
	store := wasmtime.NewStore(engine)
	
	// 设置燃料限制
	store.AddFuel(1000000) // 设置初始燃料
	
	loader := &WASMPluginLoader{
		loadedModules:   make(map[string]*WASMModule),
		engine:          engine,
		store:           store,
		config:          config,
		logger:          logger,
		securityManager: securityManager,
		ctx:             ctx,
		cancel:          cancel,
	}
	
	// 创建资源监控器
	loader.resourceMonitor = NewResourceMonitor(logger)
	
	return loader
}

// LoadPlugin 加载 WASM 插件
func (l *WASMPluginLoader) LoadPlugin(ctx context.Context, pluginPath string) (Plugin, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// 检查是否已加载
	if module, exists := l.loadedModules[pluginPath]; exists {
		if module.State == PluginStateRunning {
			return module.PluginInstance, nil
		}
	}
	
	// 检查模块数量限制
	if len(l.loadedModules) >= l.config.MaxModules {
		return nil, fmt.Errorf("maximum number of WASM modules (%d) exceeded", l.config.MaxModules)
	}
	
	// 验证插件路径
	if err := l.validatePluginPath(pluginPath); err != nil {
		return nil, fmt.Errorf("plugin path validation failed: %w", err)
	}
	
	// 安全验证
	if err := l.securityManager.ValidatePlugin(pluginPath); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}
	
	// 创建加载上下文
	loadCtx, loadCancel := context.WithTimeout(ctx, l.config.LoadTimeout)
	defer loadCancel()
	
	// 加载 WASM 模块
	wasmModule, err := l.loadWASMModule(loadCtx, pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load WASM module: %w", err)
	}
	
	// 创建沙箱环境
	if l.config.EnableSandbox {
		// 简化沙箱创建，实际应该调用安全管理器
		wasmModule.sandbox = &WASMSandbox{
			limits: &SandboxLimits{
				MaxMemory:             l.config.MaxMemorySize,
				MaxExecutionTime:      l.config.ExecutionTimeout,
				MaxCPUTime:            l.config.ExecutionTimeout,
				MaxFileSize:           1024 * 1024, // 1MB
				MaxNetworkConnections: 0,
			},
			policy:    l.config.SecurityPolicy,
			startTime: time.Now(),
			active:    true,
		}
	}
	
	// 初始化插件
	if err := l.initializePlugin(wasmModule); err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}
	
	// 创建插件包装器
	pluginWrapper := l.createWASMPluginWrapper(wasmModule)
	wasmModule.PluginInstance = pluginWrapper
	wasmModule.State = PluginStateRunning
	wasmModule.LoadTime = time.Now()
	
	// 注册到已加载模块列表
	l.loadedModules[pluginPath] = wasmModule
	
	// 启动资源监控
	l.resourceMonitor.AddModule(wasmModule)
	
	l.logger.Info("WASM plugin loaded successfully", "path", pluginPath, "id", wasmModule.ID)
	
	return pluginWrapper, nil
}

// UnloadPlugin 卸载 WASM 插件
func (l *WASMPluginLoader) UnloadPlugin(ctx context.Context, pluginID string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// 查找插件模块
	var targetModule *WASMModule
	var targetPath string
	
	for path, module := range l.loadedModules {
		if module.ID == pluginID {
			targetModule = module
			targetPath = path
			break
		}
	}
	
	if targetModule == nil {
		return fmt.Errorf("WASM plugin not found: %s", pluginID)
	}
	
	// 停止插件
	if err := l.stopPlugin(targetModule); err != nil {
		return fmt.Errorf("failed to stop WASM plugin: %w", err)
	}
	
	// 清理资源
	if err := l.cleanupModule(targetModule); err != nil {
		l.logger.Warn("failed to cleanup WASM module resources", "error", err)
	}
	
	// 从已加载模块列表中移除
	delete(l.loadedModules, targetPath)
	
	// 停止资源监控
	l.resourceMonitor.RemoveModule(pluginID)
	
	l.logger.Info("WASM plugin unloaded successfully", "id", pluginID)
	return nil
}

// GetLoadedPlugins 获取已加载的插件列表（实现PluginLoader接口）
func (l *WASMPluginLoader) GetLoadedPlugins() map[string]Plugin {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	result := make(map[string]Plugin)
	for _, module := range l.loadedModules {
		if module.State == PluginStateRunning {
			result[module.ID] = module.PluginInstance
		}
	}
	
	return result
}

// IsPluginLoaded 检查插件是否已加载
func (l *WASMPluginLoader) IsPluginLoaded(pluginID string) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	for _, module := range l.loadedModules {
		if module.ID == pluginID && module.State == PluginStateRunning {
			return true
		}
	}
	
	return false
}

// GetPluginInfo 获取插件信息
func (l *WASMPluginLoader) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	for _, module := range l.loadedModules {
		if module.ID == pluginID {
			if wrapper, ok := module.PluginInstance.(*WASMPluginWrapper); ok {
				return wrapper.GetInfo(), nil
			}
		}
	}
	
	return nil, fmt.Errorf("WASM plugin not found: %s", pluginID)
}

// ReloadPlugin 重新加载插件
func (l *WASMPluginLoader) ReloadPlugin(ctx context.Context, pluginID string) error {
	// 获取插件路径
	l.mutex.RLock()
	var pluginPath string
	for path, module := range l.loadedModules {
		if module.ID == pluginID {
			pluginPath = path
			break
		}
	}
	l.mutex.RUnlock()
	
	if pluginPath == "" {
		return fmt.Errorf("WASM plugin not found: %s", pluginID)
	}
	
	// 卸载插件
	if err := l.UnloadPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to unload plugin for reload: %w", err)
	}
	
	// 重新加载插件
	_, err := l.LoadPlugin(ctx, pluginPath)
	if err != nil {
		return fmt.Errorf("failed to reload plugin: %w", err)
	}
	
	return nil
}

// ValidatePlugin 验证插件
func (l *WASMPluginLoader) ValidatePlugin(pluginPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin file does not exist: %s", pluginPath)
	}
	
	// 检查文件扩展名
	if filepath.Ext(pluginPath) != ".wasm" {
		return fmt.Errorf("invalid WASM plugin file extension: %s", pluginPath)
	}
	
	// 验证路径权限
	if err := l.validatePluginPath(pluginPath); err != nil {
		return err
	}
	
	// 安全验证
	if err := l.securityManager.ValidatePlugin(pluginPath); err != nil {
		return err
	}
	
	// 尝试加载模块进行验证
	wasmBytes, err := os.ReadFile(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to read WASM file: %w", err)
	}
	
	// 验证 WASM 模块格式
	_, err = wasmtime.NewModule(l.engine, wasmBytes)
	if err != nil {
		return fmt.Errorf("invalid WASM module: %w", err)
	}
	
	return nil
}

// GetLoaderType 获取加载器类型（实现PluginLoader接口）
func (l *WASMPluginLoader) GetLoaderType() PluginType {
	return PluginTypeWebAssembly
}

// GetLoaderInfo 获取加载器信息（实现PluginLoader接口）
func (l *WASMPluginLoader) GetLoaderInfo() map[string]interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	return map[string]interface{}{
		"type":               "wasm",
		"loaded_count":       len(l.loadedModules),
		"max_modules":        l.config.MaxModules,
		"load_timeout":       l.config.LoadTimeout,
		"execution_timeout":  l.config.ExecutionTimeout,
		"max_memory_size":    l.config.MaxMemorySize,
		"max_stack_size":     l.config.MaxStackSize,
		"enable_sandbox":     l.config.EnableSandbox,
	}
}

// Shutdown 关闭加载器（实现PluginLoader接口）
func (l *WASMPluginLoader) Shutdown(ctx context.Context) error {
	return l.Cleanup()
}

// Cleanup 清理资源
func (l *WASMPluginLoader) Cleanup() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// 停止资源监控
	if l.resourceMonitor != nil {
		l.resourceMonitor.Stop()
	}
	
	// 卸载所有插件
	for pluginID := range l.loadedModules {
		if err := l.UnloadPlugin(l.ctx, pluginID); err != nil {
			l.logger.Warn("failed to unload plugin during cleanup", "id", pluginID, "error", err)
		}
	}
	
	// 取消上下文
	l.cancel()
	
	l.logger.Info("WASM plugin loader cleanup completed")
	return nil
}

// validatePluginPath 验证插件路径
func (l *WASMPluginLoader) validatePluginPath(pluginPath string) error {
	// 检查路径是否在允许的路径列表中
	if len(l.config.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range l.config.AllowedPaths {
			if filepath.HasPrefix(pluginPath, allowedPath) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("plugin path not allowed: %s", pluginPath)
		}
	}
	
	return nil
}

// NewResourceMonitor 创建资源监控器
func NewResourceMonitor(logger *slog.Logger) *ResourceMonitor {
	return &ResourceMonitor{
		modules: make(map[string]*WASMModule),
		ticker:  time.NewTicker(1 * time.Second),
		stopCh:  make(chan struct{}),
		logger:  logger,
	}
}

// AddModule 添加模块到监控
func (rm *ResourceMonitor) AddModule(module *WASMModule) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	rm.modules[module.ID] = module
}

// RemoveModule 从监控中移除模块
func (rm *ResourceMonitor) RemoveModule(moduleID string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	delete(rm.modules, moduleID)
}

// Stop 停止资源监控
func (rm *ResourceMonitor) Stop() {
	rm.ticker.Stop()
	close(rm.stopCh)
}

// loadWASMModule 加载 WASM 模块
func (l *WASMPluginLoader) loadWASMModule(ctx context.Context, pluginPath string) (*WASMModule, error) {
	// 读取 WASM 文件
	wasmBytes, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read WASM file: %w", err)
	}

	// 创建 WASM 模块
	module, err := wasmtime.NewModule(l.engine, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create WASM module: %w", err)
	}

	// 创建新的存储实例
	store := wasmtime.NewStore(l.engine)
	store.AddFuel(1000000) // 设置燃料限制

	// 创建 WASM 实例
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
	if err != nil {
		return nil, fmt.Errorf("failed to create WASM instance: %w", err)
	}

	// 获取导出函数
	exports := make(map[string]*wasmtime.Func)
	for _, export := range instance.Exports(store) {
		if fn := export.Func(); fn != nil {
			// 使用反射或其他方式获取函数名称
			// 这里简化处理，使用索引作为名称
			funcName := fmt.Sprintf("func_%p", fn)
			exports[funcName] = fn
		}
	}
	
	// 手动添加常见的导出函数
	if pluginInit := instance.GetFunc(store, "plugin_init"); pluginInit != nil {
		exports["plugin_init"] = pluginInit
	}
	if pluginInfo := instance.GetFunc(store, "plugin_info"); pluginInfo != nil {
		exports["plugin_info"] = pluginInfo
	}
	if pluginExecute := instance.GetFunc(store, "plugin_execute"); pluginExecute != nil {
		exports["plugin_execute"] = pluginExecute
	}
	if pluginCleanup := instance.GetFunc(store, "plugin_cleanup"); pluginCleanup != nil {
		exports["plugin_cleanup"] = pluginCleanup
	}

	// 获取内存
	var memory *wasmtime.Memory
	if memExport := instance.GetExport(store, "memory"); memExport != nil {
		memory = memExport.Memory()
	}

	// 生成插件 ID
	pluginID := fmt.Sprintf("wasm_%d_%s", time.Now().UnixNano(), filepath.Base(pluginPath))

	wasmModule := &WASMModule{
		ID:       pluginID,
		Path:     pluginPath,
		Module:   module,
		Instance: instance,
		Store:    store,
		Exports:  exports,
		Memory:   memory,
		State:    PluginStateLoaded,
		Metadata: make(map[string]interface{}),
		ResourceUsage: &ResourceUsage{
			MemoryUsed:     0,
			CPUTime:        0,
			ExecutionCount: 0,
			ErrorCount:     0,
		},
		LoadedAt: time.Now(),
	}

	return wasmModule, nil
}

// initializePlugin 初始化插件
func (l *WASMPluginLoader) initializePlugin(module *WASMModule) error {
	// 检查必需的导出函数
	requiredFunctions := []string{"plugin_init", "plugin_info", "plugin_execute"}
	for _, funcName := range requiredFunctions {
		if _, exists := module.Exports[funcName]; !exists {
			return fmt.Errorf("required function not found: %s", funcName)
		}
	}

	// 调用插件初始化函数
	initFunc := module.Exports["plugin_init"]
	if initFunc != nil {
		_, err := initFunc.Call(module.Store)
		if err != nil {
			return fmt.Errorf("plugin initialization failed: %w", err)
		}
	}

	return nil
}

// stopPlugin 停止插件
func (l *WASMPluginLoader) stopPlugin(module *WASMModule) error {
	module.State = PluginStateStopped

	// 调用插件清理函数
	if cleanupFunc, exists := module.Exports["plugin_cleanup"]; exists {
		_, err := cleanupFunc.Call(module.Store)
		if err != nil {
			l.logger.Warn("plugin cleanup function failed", "id", module.ID, "error", err)
		}
	}

	module.State = PluginStateStopped
	return nil
}

// cleanupModule 清理模块资源
func (l *WASMPluginLoader) cleanupModule(module *WASMModule) error {
	// 销毁沙箱
	if module.sandbox != nil {
		// 简化沙箱销毁，实际应该调用安全管理器
		module.sandbox.active = false
		module.sandbox = nil
	}

	// 清理 WASM 资源
	if module.Store != nil {
		// Store 会在 GC 时自动清理
		module.Store = nil
	}

	module.Instance = nil
	module.Module = nil
	module.Memory = nil
	module.Exports = nil

	return nil
}

// createWASMPluginWrapper 创建 WASM 插件包装器
func (l *WASMPluginLoader) createWASMPluginWrapper(module *WASMModule) *WASMPluginWrapper {
	return &WASMPluginWrapper{
		module: module,
		loader: l,
		logger: l.logger,
	}
}

// WASMPluginWrapper WASM 插件包装器
// 实现 Plugin 接口，为 WASM 插件提供统一的接口
type WASMPluginWrapper struct {
	// module 关联的 WASM 模块
	module *WASMModule

	// loader 加载器引用
	loader *WASMPluginLoader

	// logger 日志记录器
	logger *slog.Logger

	// mutex 状态锁
	mutex sync.RWMutex
}
// GetInfo 获取插件信息
func (w *WASMPluginWrapper) GetInfo() *PluginInfo {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.getPluginInfo()
}

// getPluginInfo 内部获取插件信息的方法
func (w *WASMPluginWrapper) getPluginInfo() *PluginInfo {
	// 调用插件的 plugin_info 函数获取信息
	infoFunc, exists := w.module.Exports["plugin_info"]
	if !exists {
		return &PluginInfo{
			Name:        filepath.Base(w.module.Path),
			Version:     "1.0.0",
			Description: "WebAssembly Plugin",
			Author:      "Unknown",
		}
	}

	// 调用函数获取插件信息
	rawResult, err := infoFunc.Call(w.module.Store)
	if err != nil {
		w.logger.Warn("failed to get plugin info", "id", w.module.ID, "error", err)
		return &PluginInfo{
			Name:        filepath.Base(w.module.Path),
			Version:     "1.0.0",
			Description: "WebAssembly Plugin",
			Author:      "Unknown",
		}
	}

	// 解析返回结果（这里简化处理，实际应该根据具体协议解析）
	info := &PluginInfo{
		Name:        filepath.Base(w.module.Path),
		Version:     "1.0.0",
		Description: "WebAssembly Plugin",
		Author:      "Unknown",
	}

	// 类型断言转换为正确类型
	result, ok := rawResult.([]wasmtime.Val)
	if !ok {
		w.logger.Warn("invalid result type from plugin_info", "id", w.module.ID)
		result = []wasmtime.Val{}
	}

	// 如果有返回值，尝试解析
	if len(result) > 0 {
		if result[0].Kind() == wasmtime.KindI32 {
			addr := result[0].I32()
			// 从内存中读取插件信息（简化实现）
			if w.module.Memory != nil {
				// 这里应该实现具体的内存读取和解析逻辑
				_ = addr // 避免未使用变量警告
			}
		}
	}

	return info
}

// GetCapabilities 获取插件能力
func (w *WASMPluginWrapper) GetCapabilities() []string {
	return []string{"wasm", "sandbox", "resource_limit"}
}

// GetDependencies 获取插件依赖
func (w *WASMPluginWrapper) GetDependencies() []string {
	return []string{}
}

// GetMetrics 获取插件指标
func (w *WASMPluginWrapper) GetMetrics() (*PluginMetrics, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	return &PluginMetrics{
		PluginID:        w.module.ID,
		Uptime:          time.Since(w.module.LoadedAt),
		MemoryUsage:     int64(w.module.ResourceUsage.MemoryUsed),
		CPUUsage:        float64(w.module.ResourceUsage.CPUTime.Nanoseconds()) / 1e9,
		RequestCount:    w.module.ResourceUsage.ExecutionCount,
		ErrorCount:      w.module.ResourceUsage.ErrorCount,
		SuccessRate:     w.calculateSuccessRate(),
		CustomMetrics:   make(map[string]interface{}),
		Timestamp:       time.Now(),
	}, nil
}

// calculateSuccessRate 计算成功率
func (w *WASMPluginWrapper) calculateSuccessRate() float64 {
	if w.module.ResourceUsage.ExecutionCount == 0 {
		return 0.0
	}
	successCount := w.module.ResourceUsage.ExecutionCount - w.module.ResourceUsage.ErrorCount
	return float64(successCount) / float64(w.module.ResourceUsage.ExecutionCount) * 100.0
}

// ValidateConfig 验证配置
func (w *WASMPluginWrapper) ValidateConfig(config map[string]interface{}) error {
	// 简化实现，实际应该调用插件的验证函数
	return nil
}

// UpdateConfig 更新配置
func (w *WASMPluginWrapper) UpdateConfig(config map[string]interface{}) error {
	// 简化实现，实际应该调用插件的配置更新函数
	return nil
}

// HealthCheck 健康检查
func (w *WASMPluginWrapper) HealthCheck() error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	if w.module.State != PluginStateRunning {
		return fmt.Errorf("plugin is not running")
	}
	// 简化实现，实际应该调用插件的事件处理函数
	return nil
}

// Initialize 初始化插件
func (w *WASMPluginWrapper) Initialize(ctx PluginContext) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.module.State != PluginStateLoaded {
		return fmt.Errorf("plugin is not in loaded state")
	}

	// 调用插件初始化函数
	initFunc, exists := w.module.Exports["plugin_initialize"]
	if exists {
		_, err := initFunc.Call(w.module.Store)
		if err != nil {
			return fmt.Errorf("plugin initialization failed: %w", err)
		}
	}

	w.module.State = PluginStateRunning
	return nil
}

// Start 启动插件
func (w *WASMPluginWrapper) Start() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.module.State != PluginStateRunning {
		return fmt.Errorf("plugin is not in running state")
	}

	// 调用插件启动函数
	startFunc, exists := w.module.Exports["plugin_start"]
	if exists {
		_, err := startFunc.Call(w.module.Store)
		if err != nil {
			return fmt.Errorf("plugin start failed: %w", err)
		}
	}

	return nil
}

// Stop 停止插件
func (w *WASMPluginWrapper) Stop() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.module.State = PluginStateStopped

	// 调用插件停止函数
	stopFunc, exists := w.module.Exports["plugin_stop"]
	if exists {
		_, err := stopFunc.Call(w.module.Store)
		if err != nil {
			w.logger.Warn("plugin stop function failed", "id", w.module.ID, "error", err)
		}
	}

	w.module.State = PluginStateStopped
	return nil
}

// Cleanup 清理插件资源
func (w *WASMPluginWrapper) Cleanup() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 清理 WASM 实例
	if w.module.Instance != nil {
		w.module.Instance = nil
	}

	// 清理存储
	if w.module.Store != nil {
		w.module.Store = nil
	}

	// 清理模块
	if w.module.Module != nil {
		w.module.Module = nil
	}

	w.module.State = PluginStateUnloaded
	w.logger.Info("WASM plugin cleaned up", "plugin_id", w.module.ID)
	return nil
}

// Execute 执行插件功能
func (w *WASMPluginWrapper) Execute(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if w.module.State != PluginStateRunning {
		return nil, fmt.Errorf("plugin is not in running state")
	}

	// 更新资源使用统计
	w.module.ResourceUsage.ExecutionCount++
	w.module.ResourceUsage.LastExecution = time.Now()
	w.module.LastAccess = time.Now()

	// 创建执行上下文
	execCtx, execCancel := context.WithTimeout(ctx, w.loader.config.ExecutionTimeout)
	defer execCancel()

	// 执行插件函数
	execFunc, exists := w.module.Exports["plugin_execute"]
	if !exists {
		return nil, fmt.Errorf("plugin_execute function not found")
	}

	// 在沙箱中执行
	var result []wasmtime.Val
	var err error

	if w.module.sandbox != nil && w.module.sandbox.active {
		// 在沙箱环境中执行
		result, err = w.executeInSandbox(execCtx, execFunc, request)
	} else {
		// 直接执行
		rawResult, callErr := execFunc.Call(w.module.Store)
		if callErr != nil {
			err = callErr
		} else if vals, ok := rawResult.([]wasmtime.Val); ok {
			result = vals
		} else {
			err = fmt.Errorf("unexpected return type from plugin function")
		}
	}

	if err != nil {
		w.module.ResourceUsage.ErrorCount++
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	// 解析执行结果
	response := make(map[string]interface{})
	response["success"] = true
	response["data"] = make(map[string]interface{})

	// 处理返回值（简化实现）
	if len(result) > 0 {
		addr := result[0].I32()
		// 从内存中读取结果数据
		if w.module.Memory != nil {
			// 这里应该实现具体的内存读取和解析逻辑
			// 设置结果地址
			if data, ok := response["data"].(map[string]interface{}); ok {
				data["result_addr"] = int32(addr)
			} else {
				// 如果data不存在，创建它
				response["data"] = map[string]interface{}{
					"result_addr": int32(addr),
				}
			}
		}
	}

	return response, nil
}

// executeInSandbox 在沙箱中执行函数
func (w *WASMPluginWrapper) executeInSandbox(ctx context.Context, fn *wasmtime.Func, request map[string]interface{}) ([]wasmtime.Val, error) {
	// 检查沙箱限制
	if err := w.checkSandboxLimits(); err != nil {
		return nil, err
	}

	// 设置执行超时
	done := make(chan struct{})
	var result []wasmtime.Val
	var execErr error

	go func() {
		defer close(done)
		rawResult, callErr := fn.Call(w.module.Store)
		if callErr != nil {
			execErr = callErr
			return
		}

		if vals, ok := rawResult.([]wasmtime.Val); ok {
			result = vals
		} else {
			execErr = fmt.Errorf("unexpected return type from plugin function")
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return result, execErr
	}
}

// checkSandboxLimits 检查沙箱限制
func (w *WASMPluginWrapper) checkSandboxLimits() error {
	if w.module.sandbox == nil {
		return nil
	}

	// 检查内存使用
	if w.module.Memory != nil {
		memorySize := w.module.Memory.DataSize(w.module.Store)
		if uint64(memorySize) > w.module.sandbox.limits.MaxMemory {
			return fmt.Errorf("memory limit exceeded: %d > %d", memorySize, w.module.sandbox.limits.MaxMemory)
		}
	}

	// 检查执行时间
	if w.module.sandbox.monitor != nil {
		if w.module.sandbox.monitor.executionTime > w.module.sandbox.limits.MaxExecutionTime {
			return fmt.Errorf("execution time limit exceeded")
		}
	}

	return nil
}

// GetState 获取插件状态
func (w *WASMPluginWrapper) GetState() PluginState {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.module.State
}



// IsHealthy 检查插件健康状态
func (w *WASMPluginWrapper) IsHealthy() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// 检查基本状态
	if w.module.State != PluginStateRunning {
		return false
	}

	// 检查资源使用是否正常
	if w.module.ResourceUsage.ErrorCount > 10 {
		return false
	}

	// 调用插件健康检查函数
	healthFunc, exists := w.module.Exports["plugin_health_check"]
	if exists {
		rawResult, err := healthFunc.Call(w.module.Store)
		if err != nil {
			return false
		}
		// 检查返回值
		if result, ok := rawResult.([]wasmtime.Val); ok && len(result) > 0 {
			healthy := result[0].I32()
			return healthy != 0
		}
	}

	return true
}

// GetConfig 获取插件配置
func (w *WASMPluginWrapper) GetConfig() map[string]interface{} {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	config := make(map[string]interface{})
	config["id"] = w.module.ID
	config["path"] = w.module.Path
	config["type"] = "webassembly"
	config["state"] = w.module.State.String()
	config["load_time"] = w.module.LoadTime

	return config
}

// SetConfig 设置插件配置
func (w *WASMPluginWrapper) SetConfig(config map[string]interface{}) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 调用插件配置设置函数
	configFunc, exists := w.module.Exports["plugin_set_config"]
	if exists {
		// 这里应该将配置序列化并传递给 WASM 函数
		_, err := configFunc.Call(w.module.Store)
		if err != nil {
			return fmt.Errorf("failed to set plugin config: %w", err)
		}
	}

	return nil
}

// HandleEvent 处理事件
func (w *WASMPluginWrapper) HandleEvent(event interface{}) error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if w.module.State != PluginStateRunning {
		return fmt.Errorf("plugin is not in running state")
	}

	// 调用插件事件处理函数
	eventFunc, exists := w.module.Exports["plugin_handle_event"]
	if exists {
		// 这里应该将事件序列化并传递给 WASM 函数
		_, err := eventFunc.Call(w.module.Store)
		if err != nil {
			return fmt.Errorf("failed to handle event: %w", err)
		}
	}

	return nil
}