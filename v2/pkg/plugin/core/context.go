package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/plugin/loader"
)





// PluginContext 插件上下文接口
// 为插件提供访问核心服务的能力
type PluginContext interface {
	// 获取上下文
	GetContext() context.Context

	// 获取依赖注入容器
	GetContainer() ServiceRegistry

	// 获取事件总线
	GetEventBus() EventBus

	// 获取服务注册表
	GetServiceRegistry() ServiceRegistry

	// 获取日志记录器
	GetLogger() Logger

	// 获取插件配置
	GetPluginConfig() PluginConfig
	UpdateConfig(config PluginConfig) error

	// 获取插件数据目录
	GetDataDir() string
	GetTempDir() string

	// 插件间通信
	SendMessage(topic string, data interface{}) error
	Subscribe(topic string, handler EventHandler) error
	Unsubscribe(topic string, handler EventHandler) error
	BroadcastMessage(message interface{}) error

	// 资源管理
	GetResourceMonitor() *ResourceMonitor
	GetSecurityManager() *SecurityManager
	GetIsolationGroup() *IsolationGroup

	// 生命周期管理
	Shutdown() error
}

// EventBus 事件总线接口
type EventBus interface {
	// 发布事件
	Publish(eventType string, data interface{}) error

	// 订阅事件
	Subscribe(eventType string, handler EventHandler) error

	// 取消订阅
	Unsubscribe(eventType string, handler EventHandler) error

	// 获取订阅者数量
	GetSubscriberCount(eventType string) int
}

// EventHandler 事件处理器类型
type EventHandler func(context.Context, interface{}) error

// EventBusAdapter 事件总线适配器，将内部EventBus适配为loader.EventBus
type EventBusAdapter struct {
	eventBus EventBus
}

// Publish 发布事件
func (e *EventBusAdapter) Publish(event string, data interface{}) error {
	return e.eventBus.Publish(event, data)
}

// Subscribe 订阅事件
func (e *EventBusAdapter) Subscribe(event string, handler func(interface{})) error {
	// 将loader的handler适配为内部的EventHandler
	eventHandler := func(ctx context.Context, data interface{}) error {
		handler(data)
		return nil
	}
	return e.eventBus.Subscribe(event, eventHandler)
}

// Unsubscribe 取消订阅事件
func (e *EventBusAdapter) Unsubscribe(event string, handler func(interface{})) error {
	// 简化实现，实际应该维护handler映射
	return nil
}

// ServiceRegistryAdapter 服务注册表适配器
type ServiceRegistryAdapter struct {
	registry ServiceRegistry
}

// RegisterService 注册服务
func (s *ServiceRegistryAdapter) RegisterService(name string, service interface{}) error {
	return s.registry.RegisterService(name, service)
}

// GetService 获取服务
func (s *ServiceRegistryAdapter) GetService(name string) (interface{}, error) {
	return s.registry.GetService(name)
}

// UnregisterService 注销服务
func (s *ServiceRegistryAdapter) UnregisterService(name string) error {
	return s.registry.UnregisterService(name)
}

// ListServices 列出所有服务
func (s *ServiceRegistryAdapter) ListServices() []string {
	return s.registry.ListServices()
}

// LoggerAdapter 日志适配器
type LoggerAdapter struct {
	logger Logger
}

// PluginContextAdapter 插件上下文适配器，将PluginContextImpl适配为loader.PluginContext
type PluginContextAdapter struct {
	ctx *PluginContextImpl
}

// GetContext 获取上下文
func (p *PluginContextAdapter) GetContext() context.Context {
	return p.ctx.GetContext()
}

// GetEventBus 获取事件总线
func (p *PluginContextAdapter) GetEventBus() loader.EventBus {
	return p.ctx.GetLoaderEventBus()
}

// GetServiceRegistry 获取服务注册表
func (p *PluginContextAdapter) GetServiceRegistry() loader.ServiceRegistry {
	return p.ctx.GetLoaderServiceRegistry()
}

// GetLogger 获取日志器
func (p *PluginContextAdapter) GetLogger() loader.Logger {
	return p.ctx.GetLoaderLogger()
}

// GetConfig 获取配置
func (p *PluginContextAdapter) GetConfig() map[string]interface{} {
	return p.ctx.GetConfig()
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

// ServiceRegistry 服务注册表接口
type ServiceRegistry interface {
	// 注册服务
	RegisterService(name string, service interface{}) error

	// 获取服务
	GetService(name string) (interface{}, error)

	// 注销服务
	UnregisterService(name string) error

	// 列出所有服务
	ListServices() []string

	// 检查服务是否存在
	HasService(name string) bool
}

// PluginContextImpl PluginContext接口的默认实现
type PluginContextImpl struct {
	ctx           context.Context
	cancel        context.CancelFunc
	container     ServiceRegistry
	eventBus      EventBus
	logger        Logger
	config        PluginConfig
	resourceMonitor *ResourceMonitor
	securityManager *SecurityManager
	configWatcher   *ConfigWatcher
	isolationGroup  *IsolationGroup
	mu            sync.RWMutex
}

// NewPluginContext 创建新的插件上下文
func NewPluginContext(ctx context.Context, config PluginConfig) *PluginContextImpl {
	ctx, cancel := context.WithCancel(ctx)
	pluginCtx := &PluginContextImpl{
		ctx:       ctx,
		cancel:    cancel,
		container: NewServiceRegistry(),
		eventBus:  NewEventBus(),
		logger:    NewLogger(config.GetID()),
		config:    config,
	}
	
	// 初始化资源监控
	if config.GetResourceLimits() != nil && config.GetResourceLimits().Enabled {
		pluginCtx.resourceMonitor = NewResourceMonitor(config.GetResourceLimits())
	}
	
	// 初始化安全管理器
	if config.GetSecurityConfig() != nil {
		pluginCtx.securityManager = NewSecurityManager(config.GetSecurityConfig())
	}
	
	// 初始化配置监听器
	pluginCtx.configWatcher = NewConfigWatcher(config.GetID())
	
	// 添加配置变更处理器
	pluginCtx.configWatcher.AddHandler(func(newConfig PluginConfig) {
		pluginCtx.UpdateConfig(newConfig)
	})
	
	// 启动配置监听
	pluginCtx.configWatcher.Start()
	
	// 初始化隔离组
	pluginCtx.isolationGroup = NewIsolationGroup(config.GetID())
	
	return pluginCtx
}

// GetContext 获取上下文
func (p *PluginContextImpl) GetContext() context.Context {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ctx
}

// GetContainer 获取依赖注入容器
func (p *PluginContextImpl) GetContainer() ServiceRegistry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.container
}

// GetEventBus 获取事件总线（实现PluginContext接口）
func (p *PluginContextImpl) GetEventBus() EventBus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.eventBus
}

// GetLoaderEventBus 获取loader事件总线（实现loader.PluginContext接口）
func (p *PluginContextImpl) GetLoaderEventBus() loader.EventBus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return &EventBusAdapter{eventBus: p.eventBus}
}

// GetInternalEventBus 获取内部事件总线
func (p *PluginContextImpl) GetInternalEventBus() EventBus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.eventBus
}

// GetServiceRegistry 获取服务注册表（实现PluginContext接口）
func (p *PluginContextImpl) GetServiceRegistry() ServiceRegistry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.container
}

// GetLoaderServiceRegistry 获取loader服务注册表（实现loader.PluginContext接口）
func (p *PluginContextImpl) GetLoaderServiceRegistry() loader.ServiceRegistry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return &ServiceRegistryAdapter{registry: p.container}
}

// GetLogger 获取日志记录器（实现PluginContext接口）
func (p *PluginContextImpl) GetLogger() Logger {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.logger
}

// GetLoaderLogger 获取loader日志记录器（实现loader.PluginContext接口）
func (p *PluginContextImpl) GetLoaderLogger() loader.Logger {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return &LoggerAdapter{logger: p.logger}
}

// GetPluginConfig 获取插件配置
func (p *PluginContextImpl) GetPluginConfig() PluginConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// GetConfig 获取配置（实现loader.PluginContext接口）
func (p *PluginContextImpl) GetConfig() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.config != nil {
		return p.config.GetCustomConfig()
	}
	return make(map[string]interface{})
}

// UpdateConfig 更新配置
func (p *PluginContextImpl) UpdateConfig(config PluginConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// 验证配置
	if err := config.Validate(); err != nil {
		return err
	}
	
	p.config = config
	
	// 更新资源监控配置
	if p.resourceMonitor != nil {
		p.resourceMonitor.UpdateLimits(config.GetResourceLimits())
	}
	
	// 更新安全管理器配置
	if p.securityManager != nil {
		p.securityManager.UpdateConfig(config.GetSecurityConfig())
	}
	
	return nil
}

// GetDataDir 获取数据目录
func (p *PluginContextImpl) GetDataDir() string {
	return filepath.Join("data", "plugins", p.config.GetID())
}

// GetTempDir 获取临时目录
func (p *PluginContextImpl) GetTempDir() string {
	return filepath.Join(os.TempDir(), "musicfox", "plugins", p.config.GetID())
}

// SendMessage 发送消息
func (p *PluginContextImpl) SendMessage(topic string, data interface{}) error {
	// 检查权限
	if p.securityManager != nil {
		if !p.securityManager.CheckPermission(PermissionEventAccess) {
			return NewPluginError(ErrorCodePermissionDenied, "event access permission denied")
		}
	}
	
	return p.eventBus.Publish(topic, data)
}

// Subscribe 订阅消息
func (p *PluginContextImpl) Subscribe(topic string, handler EventHandler) error {
	// 检查权限
	if p.securityManager != nil {
		if !p.securityManager.CheckPermission(PermissionEventAccess) {
			return NewPluginError(ErrorCodePermissionDenied, "event access permission denied")
		}
	}
	
	return p.eventBus.Subscribe(topic, handler)
}

// Unsubscribe 取消订阅
func (p *PluginContextImpl) Unsubscribe(topic string, handler EventHandler) error {
	return p.eventBus.Unsubscribe(topic, handler)
}

// BroadcastMessage 广播消息给所有插件
func (p *PluginContextImpl) BroadcastMessage(message interface{}) error {
	// 通过事件总线广播消息
	return p.eventBus.Publish("plugin.broadcast", message)
}

// GetResourceMonitor 获取资源监控器
func (p *PluginContextImpl) GetResourceMonitor() *ResourceMonitor {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.resourceMonitor
}

// GetSecurityManager 获取安全管理器
func (p *PluginContextImpl) GetSecurityManager() *SecurityManager {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.securityManager
}

// GetIsolationGroup 获取隔离组
func (p *PluginContextImpl) GetIsolationGroup() *IsolationGroup {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isolationGroup
}

// ReloadConfig 重新加载配置（手动触发配置热更新）
func (p *PluginContextImpl) ReloadConfig(newConfig PluginConfig) error {
	// 验证新配置
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	
	// 通知配置监听器
	if p.configWatcher != nil {
		p.configWatcher.NotifyConfigChange(newConfig)
	}
	
	return nil
}

// GetConfigWatcher 获取配置监听器
func (p *PluginContextImpl) GetConfigWatcher() *ConfigWatcher {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.configWatcher
}

// Shutdown 关闭插件上下文
func (p *PluginContextImpl) Shutdown() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// 停止配置监听器
	if p.configWatcher != nil {
		p.configWatcher.Stop()
	}
	
	// 停止资源监控
	if p.resourceMonitor != nil {
		p.resourceMonitor.Stop()
	}
	
	// 清理隔离组
	if p.isolationGroup != nil {
		p.isolationGroup.Cleanup()
	}
	
	// 取消上下文
	if p.cancel != nil {
		p.cancel()
	}
	
	return nil
}

// Cleanup 清理插件上下文资源
func (p *PluginContextImpl) Cleanup() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	var errors []error
	
	// 停止配置监听器
	if p.configWatcher != nil {
		if err := p.configWatcher.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop config watcher: %w", err))
		}
		p.configWatcher = nil
	}
	
	// 停止资源监控
	if p.resourceMonitor != nil {
		p.resourceMonitor.Stop()
		p.resourceMonitor = nil
	}
	
	// 清理隔离组
	if p.isolationGroup != nil {
		if err := p.isolationGroup.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("failed to cleanup isolation group: %w", err))
		}
		p.isolationGroup = nil
	}
	
	// 清理安全管理器
	if p.securityManager != nil {
		if err := p.securityManager.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("failed to cleanup security manager: %w", err))
		}
		p.securityManager = nil
	}
	
	// 取消上下文
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	
	// 合并错误
	if len(errors) > 0 {
		errorMsgs := make([]string, len(errors))
		for i, err := range errors {
			errorMsgs[i] = err.Error()
		}
		return fmt.Errorf("cleanup errors: %s", strings.Join(errorMsgs, "; "))
	}
	
	return nil
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	limits    *ResourceLimits
	metrics   *ResourceMetrics
	ticker    *time.Ticker
	stopCh    chan struct{}
	mu        sync.RWMutex
}

// ResourceMetrics 资源使用指标
type ResourceMetrics struct {
	MemoryUsageMB    int64     `json:"memory_usage_mb"`
	CPUUsagePercent  float64   `json:"cpu_usage_percent"`
	GoroutineCount   int       `json:"goroutine_count"`
	FileHandleCount  int       `json:"file_handle_count"`
	NetworkConnCount int       `json:"network_conn_count"`
	DiskUsageMB      int64     `json:"disk_usage_mb"`
	LastUpdated      time.Time `json:"last_updated"`
}

// NewResourceMonitor 创建资源监控器
func NewResourceMonitor(limits *ResourceLimits) *ResourceMonitor {
	rm := &ResourceMonitor{
		limits:  limits,
		metrics: &ResourceMetrics{},
		stopCh:  make(chan struct{}),
	}
	
	// 启动监控
	rm.Start()
	return rm
}

// Start 启动资源监控
func (rm *ResourceMonitor) Start() {
	rm.ticker = time.NewTicker(5 * time.Second) // 每5秒监控一次
	go rm.monitor()
}

// Stop 停止资源监控
func (rm *ResourceMonitor) Stop() {
	if rm.ticker != nil {
		rm.ticker.Stop()
	}
	close(rm.stopCh)
}

// monitor 监控资源使用情况
func (rm *ResourceMonitor) monitor() {
	for {
		select {
		case <-rm.ticker.C:
			rm.updateMetrics()
			rm.checkLimits()
		case <-rm.stopCh:
			return
		}
	}
}

// updateMetrics 更新资源使用指标
func (rm *ResourceMonitor) updateMetrics() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	// 获取内存统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	rm.metrics.MemoryUsageMB = int64(m.Alloc / 1024 / 1024)
	
	// 获取协程数
	rm.metrics.GoroutineCount = runtime.NumGoroutine()
	
	// 获取文件句柄数（简化实现）
	rm.metrics.FileHandleCount = rm.getFileHandleCount()
	
	// 更新时间戳
	rm.metrics.LastUpdated = time.Now()
}

// checkLimits 检查资源限制
func (rm *ResourceMonitor) checkLimits() {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// 检查内存限制
	if rm.limits.MaxMemoryMB > 0 && rm.metrics.MemoryUsageMB > rm.limits.MaxMemoryMB {
		rm.handleLimitExceeded("memory", rm.metrics.MemoryUsageMB, rm.limits.MaxMemoryMB)
	}
	
	// 检查CPU限制
	if rm.limits.MaxCPUPercent > 0 && rm.metrics.CPUUsagePercent > rm.limits.MaxCPUPercent {
		rm.handleLimitExceeded("cpu", rm.metrics.CPUUsagePercent, rm.limits.MaxCPUPercent)
	}
	
	// 检查协程数限制
	if rm.limits.MaxGoroutines > 0 && rm.metrics.GoroutineCount > rm.limits.MaxGoroutines {
		rm.handleLimitExceeded("goroutines", rm.metrics.GoroutineCount, rm.limits.MaxGoroutines)
	}
}

// handleLimitExceeded 处理资源限制超出
func (rm *ResourceMonitor) handleLimitExceeded(resource string, current, limit interface{}) {
	switch rm.limits.EnforceMode {
	case EnforceModeWarn:
		// 记录警告日志
		fmt.Printf("[WARN] Resource limit exceeded: %s current=%v limit=%v\n", resource, current, limit)
	case EnforceModeLimit:
		// 限制资源使用
		fmt.Printf("[LIMIT] Resource limit exceeded: %s current=%v limit=%v\n", resource, current, limit)
	case EnforceModeKill:
		// 终止插件
		fmt.Printf("[KILL] Resource limit exceeded: %s current=%v limit=%v\n", resource, current, limit)
	}
}

// UpdateLimits 更新资源限制
func (rm *ResourceMonitor) UpdateLimits(limits *ResourceLimits) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.limits = limits
}

// GetMetrics 获取资源使用指标
func (rm *ResourceMonitor) GetMetrics() *ResourceMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	// 返回副本以避免并发问题
	metrics := *rm.metrics
	return &metrics
}

// getFileHandleCount 获取文件句柄数（简化实现）
func (rm *ResourceMonitor) getFileHandleCount() int {
	// 在实际实现中，可以通过系统调用获取准确的文件句柄数
	// 这里返回一个模拟值
	return 10
}

// IsHealthy 检查资源使用是否健康
func (rm *ResourceMonitor) IsHealthy() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// 检查各项资源是否超限
	if rm.limits.MaxMemoryMB > 0 && rm.metrics.MemoryUsageMB > rm.limits.MaxMemoryMB {
		return false
	}
	
	if rm.limits.MaxCPUPercent > 0 && rm.metrics.CPUUsagePercent > rm.limits.MaxCPUPercent {
		return false
	}
	
	if rm.limits.MaxGoroutines > 0 && rm.metrics.GoroutineCount > rm.limits.MaxGoroutines {
		return false
	}
	
	return true
}

// SecurityManager 安全管理器
type SecurityManager struct {
	config      *SecurityConfig
	permissions map[Permission]bool
	mu          sync.RWMutex
}

// NewSecurityManager 创建安全管理器
func NewSecurityManager(config *SecurityConfig) *SecurityManager {
	sm := &SecurityManager{
		config:      config,
		permissions: make(map[Permission]bool),
	}
	
	// 初始化权限
	for _, perm := range config.Permissions {
		sm.permissions[perm] = true
	}
	
	return sm
}

// CheckPermission 检查权限
func (sm *SecurityManager) CheckPermission(permission Permission) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	// 检查是否有全部权限
	if sm.permissions[PermissionAll] {
		return true
	}
	
	return sm.permissions[permission]
}

// UpdateConfig 更新安全配置
func (sm *SecurityManager) UpdateConfig(config *SecurityConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sm.config = config
	sm.permissions = make(map[Permission]bool)
	
	// 重新初始化权限
	for _, perm := range config.Permissions {
		sm.permissions[perm] = true
	}
	
	return nil
}

// ValidateAccess 验证访问权限
func (sm *SecurityManager) ValidateAccess(resource string, action string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	// TODO: 实现具体的访问验证逻辑
	return nil
}

// Cleanup 清理安全管理器资源
func (sm *SecurityManager) Cleanup() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// 清理权限映射
	sm.permissions = make(map[Permission]bool)
	sm.config = nil
	
	return nil
}

// ConfigWatcher 配置监听器
type ConfigWatcher struct {
	pluginID string
	handlers []func(PluginConfig)
	stopCh   chan struct{}
	mu       sync.RWMutex
}

// NewConfigWatcher 创建配置监听器
func NewConfigWatcher(pluginID string) *ConfigWatcher {
	return &ConfigWatcher{
		pluginID: pluginID,
		handlers: make([]func(PluginConfig), 0),
		stopCh:   make(chan struct{}),
	}
}

// AddHandler 添加配置变更处理器
func (cw *ConfigWatcher) AddHandler(handler func(PluginConfig)) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.handlers = append(cw.handlers, handler)
}

// RemoveHandler 移除配置变更处理器
func (cw *ConfigWatcher) RemoveHandler(handler func(PluginConfig)) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	
	// 简化实现：清空所有处理器
	// 在实际实现中，应该精确匹配并移除特定处理器
	cw.handlers = make([]func(PluginConfig), 0)
}

// NotifyConfigChange 通知配置变更
func (cw *ConfigWatcher) NotifyConfigChange(config PluginConfig) {
	cw.mu.RLock()
	handlers := make([]func(PluginConfig), len(cw.handlers))
	copy(handlers, cw.handlers)
	cw.mu.RUnlock()
	
	// 异步通知所有处理器
	for _, handler := range handlers {
		go func(h func(PluginConfig)) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("[ERROR] Config change handler panic: %v\n", r)
				}
			}()
			h(config)
		}(handler)
	}
}

// Start 启动配置监听
func (cw *ConfigWatcher) Start() {
	// TODO: 实现文件系统监听或其他配置变更检测机制
	// 这里提供一个基础框架
	go cw.watchLoop()
}

// Stop 停止配置监听
func (cw *ConfigWatcher) Stop() error {
	close(cw.stopCh)
	return nil
}

// watchLoop 配置监听循环
func (cw *ConfigWatcher) watchLoop() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次配置变更
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// TODO: 检查配置文件是否有变更
			// 如果有变更，加载新配置并通知处理器
		case <-cw.stopCh:
			return
		}
	}
}


// IsolationGroup 隔离组
type IsolationGroup struct {
	pluginID  string
	namespace string
	resources map[string]interface{}
	mu        sync.RWMutex
}

// NewIsolationGroup 创建隔离组
func NewIsolationGroup(pluginID string) *IsolationGroup {
	return &IsolationGroup{
		pluginID:  pluginID,
		namespace: fmt.Sprintf("plugin_%s", pluginID),
		resources: make(map[string]interface{}),
	}
}

// AllocateResource 分配资源
func (ig *IsolationGroup) AllocateResource(name string, resource interface{}) {
	ig.mu.Lock()
	defer ig.mu.Unlock()
	ig.resources[name] = resource
}

// GetResource 获取资源
func (ig *IsolationGroup) GetResource(name string) (interface{}, bool) {
	ig.mu.RLock()
	defer ig.mu.RUnlock()
	resource, exists := ig.resources[name]
	return resource, exists
}

// Cleanup 清理资源
func (ig *IsolationGroup) Cleanup() error {
	ig.mu.Lock()
	defer ig.mu.Unlock()
	
	// 清理所有资源
	for name := range ig.resources {
		delete(ig.resources, name)
	}
	
	return nil
}

// Logger 日志记录器接口
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// DefaultLogger 默认日志记录器
type DefaultLogger struct {
	pluginID string
}

// NewLogger 创建日志记录器
func NewLogger(pluginID string) Logger {
	return &DefaultLogger{pluginID: pluginID}
}

// Debug 记录调试日志
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[DEBUG][%s] %s\n", l.pluginID, fmt.Sprintf(msg, args...))
}

// Info 记录信息日志
func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO][%s] %s\n", l.pluginID, fmt.Sprintf(msg, args...))
}

// Warn 记录警告日志
func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN][%s] %s\n", l.pluginID, fmt.Sprintf(msg, args...))
}

// Error 记录错误日志
func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR][%s] %s\n", l.pluginID, fmt.Sprintf(msg, args...))
}

// NewServiceRegistry 创建服务注册表
func NewServiceRegistry() ServiceRegistry {
	return &DefaultServiceRegistry{
		services: make(map[string]interface{}),
	}
}

// DefaultServiceRegistry 默认服务注册表实现
type DefaultServiceRegistry struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

// RegisterService 注册服务
func (r *DefaultServiceRegistry) RegisterService(name string, service interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[name] = service
	return nil
}

// GetService 获取服务
func (r *DefaultServiceRegistry) GetService(name string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	service, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	return service, nil
}

// UnregisterService 注销服务
func (r *DefaultServiceRegistry) UnregisterService(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.services, name)
	return nil
}

// ListServices 列出所有服务
func (r *DefaultServiceRegistry) ListServices() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	services := make([]string, 0, len(r.services))
	for name := range r.services {
		services = append(services, name)
	}
	return services
}

// HasService 检查服务是否存在
func (r *DefaultServiceRegistry) HasService(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.services[name]
	return exists
}

// NewEventBus 创建事件总线
func NewEventBus() EventBus {
	return &DefaultEventBus{
		subscribers: make(map[string][]EventHandler),
	}
}

// DefaultEventBus 默认事件总线实现
type DefaultEventBus struct {
	subscribers map[string][]EventHandler
	mu          sync.RWMutex
}

// Publish 发布事件
func (eb *DefaultEventBus) Publish(eventType string, data interface{}) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	
	handlers, exists := eb.subscribers[eventType]
	if !exists {
		return nil
	}
	
	for _, handler := range handlers {
		go func(h EventHandler) {
			_ = h(context.Background(), data)
		}(handler)
	}
	
	return nil
}

// Subscribe 订阅事件
func (eb *DefaultEventBus) Subscribe(eventType string, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
	return nil
}

// Unsubscribe 取消订阅
func (eb *DefaultEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	handlers, exists := eb.subscribers[eventType]
	if !exists {
		return nil
	}
	
	// 移除指定的处理器
	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			eb.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
	
	return nil
}

// GetSubscriberCount 获取订阅者数量
func (eb *DefaultEventBus) GetSubscriberCount(eventType string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	if handlers, exists := eb.subscribers[eventType]; exists {
		return len(handlers)
	}
	return 0
}

// UnsubscribeAll 取消所有匹配前缀的订阅
func (eb *DefaultEventBus) UnsubscribeAll(eventPrefix string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	// 找到所有匹配前缀的事件类型
	var toDelete []string
	for eventType := range eb.subscribers {
		if strings.HasPrefix(eventType, eventPrefix) {
			toDelete = append(toDelete, eventType)
		}
	}
	
	// 删除匹配的订阅
	for _, eventType := range toDelete {
		delete(eb.subscribers, eventType)
	}
}