// Package plugin 实现插件错误处理和资源管理
package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 导入必要的依赖
type EventBus interface {
	Publish(eventType string, data interface{}) error
	Subscribe(eventType string, handler interface{}) error
	Unsubscribe(eventType string, handler interface{}) error
}

// PluginLoader 插件加载器接口
type PluginLoader interface {
	ReloadPlugin(pluginID string) error
	RestartPlugin(pluginID string) error
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	MaxMemory     int64         `json:"max_memory"`     // 最大内存使用量（字节）
	MaxCPU        float64       `json:"max_cpu"`        // 最大CPU使用率（0-1）
	MaxGoroutines int           `json:"max_goroutines"` // 最大协程数
	MaxDuration   time.Duration `json:"max_duration"`   // 最大执行时间
	MaxFileSize   int64         `json:"max_file_size"`  // 最大文件大小
	MaxNetworkIO  int64         `json:"max_network_io"` // 最大网络IO
}

// DegradationConfig 降级配置
type DegradationConfig struct {
	DisableFeatures []string `json:"disable_features"` // 禁用的功能列表
	ReduceQuality   bool     `json:"reduce_quality"`   // 是否降低质量
	FallbackMode    bool     `json:"fallback_mode"`    // 是否启用回退模式
}

// PluginErrorHandler 错误处理器接口
type PluginErrorHandler interface {
	// HandleError 处理错误
	HandleError(ctx context.Context, pluginID string, err error) error
	
	// RegisterErrorCallback 注册错误回调
	RegisterErrorCallback(pluginID string, callback ErrorCallback) error
	
	// UnregisterErrorCallback 注销错误回调
	UnregisterErrorCallback(pluginID string, callback ErrorCallback) error
	
	// GetErrorHistory 获取错误历史
	GetErrorHistory(pluginID string) ([]ErrorRecord, error)
	
	// ClearErrorHistory 清理错误历史
	ClearErrorHistory(pluginID string) error
	
	// GetErrorStats 获取错误统计
	GetErrorStats(pluginID string) (*ErrorStats, error)
}

// ResourceManager 资源管理器接口
type ResourceManager interface {
	// MonitorResources 监控资源使用
	MonitorResources(ctx context.Context, pluginID string) error
	
	// GetResourceUsage 获取资源使用情况
	GetResourceUsage(pluginID string) (*ResourceUsage, error)
	
	// SetResourceLimits 设置资源限制
	SetResourceLimits(pluginID string, limits *ResourceLimits) error
	
	// CheckResourceLimits 检查资源限制
	CheckResourceLimits(pluginID string) error
	
	// CleanupResources 清理资源
	CleanupResources(ctx context.Context, pluginID string) error
	
	// ForceCleanup 强制清理资源
	ForceCleanup(pluginID string) error
}

// DynamicErrorHandler 动态错误处理器实现
type DynamicErrorHandler struct {
	// errorHistory 错误历史记录
	errorHistory map[string][]ErrorRecord
	
	// errorCallbacks 错误回调函数
	errorCallbacks map[string][]ErrorCallback
	
	// errorStats 错误统计
	errorStats map[string]*ErrorStats
	
	// mutex 读写锁
	mutex sync.RWMutex
	
	// config 错误处理配置
	config *ErrorHandlerConfig
	
	// 新增组件
	wrapper       ErrorWrapper
	propagator    ErrorPropagator
	logger        Logger
	monitor       ErrorMonitor
	classifier    ErrorClassifier
	metrics       MetricsCollector
	eventBus      EventBus
	recoveryStrategies map[ErrorCode]ErrorRecoveryStrategy
	pluginLoader  PluginLoader
}

// DynamicResourceManager 动态资源管理器实现
type DynamicResourceManager struct {
	// resourceUsage 资源使用记录
	resourceUsage map[string]*ResourceUsage
	
	// resourceLimits 资源限制
	resourceLimits map[string]*ResourceLimits
	
	// monitoringContexts 监控上下文
	monitoringContexts map[string]context.CancelFunc
	
	// mutex 读写锁
	mutex sync.RWMutex
	
	// config 资源管理配置
	config *ResourceManagerConfig
}

// ErrorRecord 错误记录
type ErrorRecord struct {
	// Timestamp 错误时间
	Timestamp time.Time
	
	// Error 错误信息
	Error error
	
	// ErrorType 错误类型
	ErrorType ErrorType
	
	// Context 错误上下文
	Context map[string]interface{}
	
	// StackTrace 堆栈跟踪
	StackTrace string
	
	// Handled 是否已处理
	Handled bool
	
	// HandledBy 处理者
	HandledBy string
}

// ErrorCallback 错误回调函数类型
type ErrorCallback func(pluginID string, record ErrorRecord) error

// ErrorStats 错误统计信息
type ErrorStats struct {
	// TotalErrors 总错误数
	TotalErrors int
	
	// ErrorsByType 按类型分组的错误数
	ErrorsByType map[ErrorType]int
	
	// LastError 最后一个错误
	LastError *ErrorRecord
	
	// FirstError 第一个错误
	FirstError *ErrorRecord
	
	// ErrorRate 错误率（每分钟）
	ErrorRate float64
	
	// RecoveryCount 恢复次数
	RecoveryCount int
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	// MemoryUsage 内存使用量（字节）
	MemoryUsage int64
	
	// CPUUsage CPU使用率（百分比）
	CPUUsage float64
	
	// GoroutineCount 协程数量
	GoroutineCount int
	
	// FileDescriptors 文件描述符数量
	FileDescriptors int
	
	// NetworkConnections 网络连接数
	NetworkConnections int
	
	// LastUpdated 最后更新时间
	LastUpdated time.Time
	
	// History 历史记录
	History []ResourceSnapshot
}

// ResourceSnapshot 资源快照
type ResourceSnapshot struct {
	// Timestamp 时间戳
	Timestamp time.Time
	
	// MemoryUsage 内存使用量
	MemoryUsage int64
	
	// CPUUsage CPU使用率
	CPUUsage float64
	
	// GoroutineCount 协程数量
	GoroutineCount int
}

// ErrorHandlerConfig 错误处理器配置
type ErrorHandlerConfig struct {
	// MaxErrorHistory 最大错误历史记录数
	MaxErrorHistory int
	
	// EnableStackTrace 是否启用堆栈跟踪
	EnableStackTrace bool
	
	// ErrorRateWindow 错误率计算窗口（分钟）
	ErrorRateWindow time.Duration
	
	// AutoRecovery 是否启用自动恢复
	AutoRecovery bool
	
	// MaxRecoveryAttempts 最大恢复尝试次数
	MaxRecoveryAttempts int
}

// ResourceManagerConfig 资源管理器配置
type ResourceManagerConfig struct {
	// MonitorInterval 监控间隔
	MonitorInterval time.Duration
	
	// MaxHistorySize 最大历史记录数
	MaxHistorySize int
	
	// EnableDetailedMonitoring 是否启用详细监控
	EnableDetailedMonitoring bool
	
	// ResourceCheckInterval 资源检查间隔
	ResourceCheckInterval time.Duration
	
	// CleanupTimeout 清理超时时间
	CleanupTimeout time.Duration
}

// NewDynamicErrorHandler 创建新的动态错误处理器
func NewDynamicErrorHandler(config *ErrorHandlerConfig) PluginErrorHandler {
	if config == nil {
		config = DefaultErrorHandlerConfig()
	}
	
	// 创建组件
	// 创建一个默认的slog.Logger
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger := NewErrorLogger(slogLogger, LogLevelInfo)
	metrics := NewMetricsCollector()
	wrapper := NewErrorWrapper(true, 10)
	monitor := NewErrorMonitor(metrics, logger)
	classifier := NewErrorClassifier(logger, metrics)
	propagator := NewErrorPropagator(logger, metrics)
	
	return &DynamicErrorHandler{
		errorHistory:   make(map[string][]ErrorRecord),
		errorCallbacks: make(map[string][]ErrorCallback),
		errorStats:     make(map[string]*ErrorStats),
		config:         config,
		wrapper:        wrapper,
		propagator:     propagator,
		logger:         logger,
		monitor:        monitor,
		classifier:     classifier,
		metrics:        metrics,
		recoveryStrategies: make(map[ErrorCode]ErrorRecoveryStrategy),
		pluginLoader:   nil, // 需要外部设置
	}
}

// NewDynamicResourceManager 创建新的动态资源管理器
func NewDynamicResourceManager(config *ResourceManagerConfig) *DynamicResourceManager {
	if config == nil {
		config = DefaultResourceManagerConfig()
	}
	
	return &DynamicResourceManager{
		resourceUsage:      make(map[string]*ResourceUsage),
		resourceLimits:     make(map[string]*ResourceLimits),
		monitoringContexts: make(map[string]context.CancelFunc),
		config:             config,
	}
}

// DefaultErrorHandlerConfig 返回默认的错误处理器配置
func DefaultErrorHandlerConfig() *ErrorHandlerConfig {
	return &ErrorHandlerConfig{
		MaxErrorHistory:     1000,
		EnableStackTrace:    true,
		ErrorRateWindow:     5 * time.Minute,
		AutoRecovery:        true,
		MaxRecoveryAttempts: 3,
	}
}

// DefaultResourceManagerConfig 返回默认的资源管理器配置
func DefaultResourceManagerConfig() *ResourceManagerConfig {
	return &ResourceManagerConfig{
		MonitorInterval:          30 * time.Second,
		MaxHistorySize:           100,
		EnableDetailedMonitoring: true,
		ResourceCheckInterval:    10 * time.Second,
		CleanupTimeout:           30 * time.Second,
	}
}

// HandleError 处理错误
func (deh *DynamicErrorHandler) HandleError(ctx context.Context, pluginID string, err error) error {
	if err == nil {
		return nil
	}
	
	// 1. 错误包装
	var pluginErr PluginError
	if deh.wrapper.IsPluginError(err) {
		pluginErr = err.(PluginError)
	} else {
		pluginErr = deh.wrapper.WrapWithContext(ctx, err, ErrorCodeUnknown, "Wrapped error")
	}
	
	// 2. 错误分类
	if deh.classifier != nil {
		if classification, classErr := deh.classifier.ClassifyError(ctx, err, pluginID); classErr == nil {
			// 更新分类结果
			if baseErr, ok := pluginErr.(*BasePluginError); ok {
				baseErr.Code = classification.ErrorCode
				baseErr.Type = classification.ErrorType
				baseErr.Severity = classification.Severity
				baseErr.WithContext("classification_confidence", classification.Confidence)
				baseErr.WithContext("classification_reason", classification.Reason)
			}
		}
	}
	
	// 3. 错误监控
	if deh.monitor != nil {
		deh.monitor.RecordError(ctx, pluginErr, pluginID)
	}
	
	// 4. 错误传播
	if deh.propagator != nil {
		if propErr := deh.propagator.PropagateError(ctx, pluginErr, pluginID, []string{}); propErr != nil {
			deh.logger.Error("Error propagation failed", map[string]interface{}{
				"plugin_id": pluginID,
				"error": propErr.Error(),
			})
		}
	}
	
	// 5. 创建错误记录
	record := ErrorRecord{
		Timestamp: time.Now(),
		Error:     pluginErr,
		ErrorType: deh.classifyError(pluginErr),
		Context:   deh.extractErrorContext(ctx),
		Handled:   false,
	}
	
	// 添加堆栈跟踪
	if deh.config.EnableStackTrace {
		record.StackTrace = deh.captureStackTrace()
	}
	
	// 6. 记录错误
	deh.recordError(pluginID, record)
	
	// 7. 更新错误统计
	deh.updateErrorStats(pluginID, record)
	
	// 8. 调用错误回调
	if err := deh.callErrorCallbacks(pluginID, record); err != nil {
		deh.logger.Error("Error callback failed", map[string]interface{}{
			"plugin_id": pluginID,
			"error": err.Error(),
		})
	}
	
	// 9. 尝试错误恢复
	if strategy, exists := deh.recoveryStrategies[pluginErr.GetCode()]; exists {
		if recoveryErr := deh.executeRecoveryStrategy(ctx, strategy, pluginErr, pluginID); recoveryErr != nil {
			deh.logger.Error("Error recovery failed", map[string]interface{}{
				"plugin_id": pluginID,
				"strategy": strategy.Type.String(),
				"error": recoveryErr.Error(),
			})
		}
	} else if deh.config.AutoRecovery {
		if err := deh.attemptRecovery(ctx, pluginID, record); err != nil {
			deh.logger.Error("Auto recovery failed", map[string]interface{}{
				"plugin_id": pluginID,
				"error": err.Error(),
			})
		}
	}
	
	// 10. 发布错误事件
	if deh.eventBus != nil {
		errorEvent := map[string]interface{}{
			"plugin_id": pluginID,
			"error_code": pluginErr.GetCode().String(),
			"error_type": pluginErr.GetType().String(),
			"severity": pluginErr.GetSeverity().String(),
			"message": pluginErr.Error(),
			"timestamp": time.Now(),
		}
		deh.eventBus.Publish("plugin.error", errorEvent)
	}
	
	return nil
}

// RegisterErrorCallback 注册错误回调
func (deh *DynamicErrorHandler) RegisterErrorCallback(pluginID string, callback ErrorCallback) error {
	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}
	
	deh.mutex.Lock()
	defer deh.mutex.Unlock()
	
	if deh.errorCallbacks[pluginID] == nil {
		deh.errorCallbacks[pluginID] = make([]ErrorCallback, 0)
	}
	
	deh.errorCallbacks[pluginID] = append(deh.errorCallbacks[pluginID], callback)
	return nil
}

// UnregisterErrorCallback 注销错误回调
func (deh *DynamicErrorHandler) UnregisterErrorCallback(pluginID string, callback ErrorCallback) error {
	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}
	
	deh.mutex.Lock()
	defer deh.mutex.Unlock()
	
	callbacks, exists := deh.errorCallbacks[pluginID]
	if !exists {
		return fmt.Errorf("no callbacks found for plugin: %s", pluginID)
	}
	
	// 由于函数比较困难，这里简单地清空所有回调
	// 在实际实现中可能需要更复杂的标识机制
	deh.errorCallbacks[pluginID] = make([]ErrorCallback, 0)
	_ = callbacks // 避免未使用变量警告
	
	return nil
}

// GetErrorHistory 获取错误历史
func (deh *DynamicErrorHandler) GetErrorHistory(pluginID string) ([]ErrorRecord, error) {
	deh.mutex.RLock()
	defer deh.mutex.RUnlock()
	
	history, exists := deh.errorHistory[pluginID]
	if !exists {
		// 返回空列表而不是错误
		return []ErrorRecord{}, nil
	}
	
	// 返回历史记录的副本
	result := make([]ErrorRecord, len(history))
	copy(result, history)
	return result, nil
}

// ClearErrorHistory 清理错误历史
func (deh *DynamicErrorHandler) ClearErrorHistory(pluginID string) error {
	deh.mutex.Lock()
	defer deh.mutex.Unlock()
	
	delete(deh.errorHistory, pluginID)
	delete(deh.errorStats, pluginID)
	return nil
}

// GetErrorStats 获取错误统计
func (deh *DynamicErrorHandler) GetErrorStats(pluginID string) (*ErrorStats, error) {
	deh.mutex.RLock()
	defer deh.mutex.RUnlock()
	
	stats, exists := deh.errorStats[pluginID]
	if !exists {
		// 返回空的统计信息而不是错误
		return &ErrorStats{
			TotalErrors:  0,
			ErrorsByType: make(map[ErrorType]int),
			ErrorRate:    0.0,
		}, nil
	}
	
	// 返回统计信息的副本
	statsCopy := *stats
	return &statsCopy, nil
}

// recordError 记录错误
func (deh *DynamicErrorHandler) recordError(pluginID string, record ErrorRecord) {
	deh.mutex.Lock()
	defer deh.mutex.Unlock()
	
	// 初始化错误历史
	if deh.errorHistory[pluginID] == nil {
		deh.errorHistory[pluginID] = make([]ErrorRecord, 0)
	}
	
	// 添加错误记录
	deh.errorHistory[pluginID] = append(deh.errorHistory[pluginID], record)
	
	// 限制历史记录大小
	if len(deh.errorHistory[pluginID]) > deh.config.MaxErrorHistory {
		deh.errorHistory[pluginID] = deh.errorHistory[pluginID][1:]
	}
}

// updateErrorStats 更新错误统计
func (deh *DynamicErrorHandler) updateErrorStats(pluginID string, record ErrorRecord) {
	deh.mutex.Lock()
	defer deh.mutex.Unlock()
	
	if deh.errorStats[pluginID] == nil {
		deh.errorStats[pluginID] = &ErrorStats{
			ErrorsByType: make(map[ErrorType]int),
		}
	}
	
	stats := deh.errorStats[pluginID]
	stats.TotalErrors++
	stats.ErrorsByType[record.ErrorType]++
	stats.LastError = &record
	
	if stats.FirstError == nil {
		stats.FirstError = &record
	}
	
	// 计算错误率（内部调用，已持有锁）
	stats.ErrorRate = deh.calculateErrorRateInternal(pluginID)
}

// classifyError 分类错误
func (deh *DynamicErrorHandler) classifyError(err error) ErrorType {
	// 简单的错误分类逻辑
	errorMsg := err.Error()
	
	switch {
	case contains(errorMsg, "initialization"):
		return ErrorTypePlugin
	case contains(errorMsg, "runtime"):
		return ErrorTypeSystem
	case contains(errorMsg, "resource"):
		return ErrorTypeResource
	case contains(errorMsg, "communication"):
		return ErrorTypeNetwork
	case contains(errorMsg, "security"):
		return ErrorTypeAuthentication
	default:
		return ErrorTypeSystem
	}
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// executeRecoveryStrategy 执行恢复策略
func (deh *DynamicErrorHandler) executeRecoveryStrategy(ctx context.Context, strategy ErrorRecoveryStrategy, err PluginError, pluginID string) error {
	switch strategy.Type {
	case RecoveryTypeRetry:
		return deh.executeRetryStrategy(ctx, strategy, err, pluginID)
	case RecoveryTypeRestart:
		return deh.executeRestartStrategy(ctx, strategy, err, pluginID)
	case RecoveryTypeFallback:
		return deh.executeFallbackStrategy(ctx, strategy, err, pluginID)
	case RecoveryTypeGracefulDegradation:
		return deh.executeGracefulDegradationStrategy(ctx, strategy, err, pluginID)
	default:
		return fmt.Errorf("unsupported recovery strategy: %s", strategy.Type.String())
	}
}

// executeRetryStrategy 执行重试策略
func (deh *DynamicErrorHandler) executeRetryStrategy(ctx context.Context, strategy ErrorRecoveryStrategy, err PluginError, pluginID string) error {
	if !err.IsRetryable() {
		return fmt.Errorf("error is not retryable")
	}
	
	for i := 0; i < strategy.MaxRetries; i++ {
		if i > 0 {
			// 计算退避延迟
			delay := strategy.RetryDelay
			if strategy.BackoffFactor > 1.0 {
				delay = time.Duration(float64(delay) * math.Pow(strategy.BackoffFactor, float64(i)))
			}
			if delay > strategy.MaxDelay {
				delay = strategy.MaxDelay
			}
			
			// 添加抖动
			if strategy.Jitter {
				jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
				delay += jitter
			}
			
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
		
		deh.logger.Info("Retrying operation", map[string]interface{}{
			"plugin_id": pluginID,
			"attempt": i + 1,
			"max_retries": strategy.MaxRetries,
		})
		
		// 这里应该重新执行失败的操作
		// 由于我们没有具体的操作上下文，这里只是记录重试
		if deh.metrics != nil {
			deh.metrics.IncrementCounter("error_recovery_retry", map[string]string{
				"plugin_id": pluginID,
				"attempt": fmt.Sprintf("%d", i+1),
			})
		}
	}
	
	return fmt.Errorf("retry strategy failed after %d attempts", strategy.MaxRetries)
}

// executeRestartStrategy 执行重启策略
func (deh *DynamicErrorHandler) executeRestartStrategy(ctx context.Context, strategy ErrorRecoveryStrategy, err PluginError, pluginID string) error {
	deh.logger.Info("Executing restart strategy", map[string]interface{}{
		"plugin_id": pluginID,
	})
	
	// 尝试重启插件
	if deh.pluginLoader != nil {
		if restartErr := deh.pluginLoader.RestartPlugin(pluginID); restartErr != nil {
			return fmt.Errorf("failed to restart plugin %s: %w", pluginID, restartErr)
		}
	}
	
	// 发布重启事件
	if deh.eventBus != nil {
		restartEvent := map[string]interface{}{
			"plugin_id": pluginID,
			"reason": "error_recovery",
			"error_code": err.GetCode().String(),
			"timestamp": time.Now(),
		}
		deh.eventBus.Publish("plugin.restart_requested", restartEvent)
	}
	
	if deh.metrics != nil {
		deh.metrics.IncrementCounter("error_recovery_restart", map[string]string{
			"plugin_id": pluginID,
		})
	}
	
	return nil
}

// executeFallbackStrategy 执行降级策略
func (deh *DynamicErrorHandler) executeFallbackStrategy(ctx context.Context, strategy ErrorRecoveryStrategy, err PluginError, pluginID string) error {
	deh.logger.Info("Executing fallback strategy", map[string]interface{}{
		"plugin_id": pluginID,
		"fallback": strategy.Fallback,
	})
	
	// 发布降级事件
	if deh.eventBus != nil {
		fallbackEvent := map[string]interface{}{
			"plugin_id": pluginID,
			"fallback_target": strategy.Fallback,
			"reason": "error_recovery",
			"timestamp": time.Now(),
		}
		deh.eventBus.Publish("plugin.fallback_activated", fallbackEvent)
	}
	
	if deh.metrics != nil {
		deh.metrics.IncrementCounter("error_recovery_fallback", map[string]string{
			"plugin_id": pluginID,
			"fallback": strategy.Fallback,
		})
	}
	
	return nil
}

// executeGracefulDegradationStrategy 执行优雅降级策略
func (deh *DynamicErrorHandler) executeGracefulDegradationStrategy(ctx context.Context, strategy ErrorRecoveryStrategy, err PluginError, pluginID string) error {
	deh.logger.Info("Executing graceful degradation strategy", map[string]interface{}{
		"plugin_id": pluginID,
	})
	
	// 发布优雅降级事件
	if deh.eventBus != nil {
		degradationEvent := map[string]interface{}{
			"plugin_id": pluginID,
			"degradation_level": "partial",
			"reason": "error_recovery",
			"timestamp": time.Now(),
		}
		deh.eventBus.Publish("plugin.degradation_activated", degradationEvent)
	}
	
	if deh.metrics != nil {
		deh.metrics.IncrementCounter("error_recovery_degradation", map[string]string{
			"plugin_id": pluginID,
		})
	}
	
	return nil
}

// SetRecoveryStrategy 设置恢复策略
func (deh *DynamicErrorHandler) SetRecoveryStrategy(errorCode ErrorCode, strategy ErrorRecoveryStrategy) {
	deh.mutex.Lock()
	defer deh.mutex.Unlock()
	
	deh.recoveryStrategies[errorCode] = strategy
	
	if deh.logger != nil {
		deh.logger.Info("Recovery strategy set", map[string]interface{}{
			"error_code": errorCode.String(),
			"strategy_type": strategy.Type.String(),
		})
	}
}

// GetRecoveryStrategy 获取恢复策略
func (deh *DynamicErrorHandler) GetRecoveryStrategy(errorCode ErrorCode) (ErrorRecoveryStrategy, bool) {
	deh.mutex.RLock()
	defer deh.mutex.RUnlock()
	
	strategy, exists := deh.recoveryStrategies[errorCode]
	return strategy, exists
}

// SetEventBus 设置事件总线
func (deh *DynamicErrorHandler) SetEventBus(eventBus EventBus) {
	deh.mutex.Lock()
	defer deh.mutex.Unlock()
	
	deh.eventBus = eventBus
}

// GetMetrics 获取指标收集器
func (deh *DynamicErrorHandler) GetMetrics() MetricsCollector {
	return deh.metrics
}

// GetClassifier 获取错误分类器
func (deh *DynamicErrorHandler) GetClassifier() ErrorClassifier {
	return deh.classifier
}

// GetMonitor 获取错误监控器
func (deh *DynamicErrorHandler) GetMonitor() ErrorMonitor {
	return deh.monitor
}

// SetPluginLoader 设置插件加载器
func (deh *DynamicErrorHandler) SetPluginLoader(pluginLoader PluginLoader) {
	deh.mutex.Lock()
	defer deh.mutex.Unlock()
	deh.pluginLoader = pluginLoader
}

// extractErrorContext 提取错误上下文
func (deh *DynamicErrorHandler) extractErrorContext(ctx context.Context) map[string]interface{} {
	context := make(map[string]interface{})
	
	if ctx != nil {
		if deadline, ok := ctx.Deadline(); ok {
			context["deadline"] = deadline
		}
		
		if ctx.Err() != nil {
			context["context_error"] = ctx.Err().Error()
		}
	}
	
	return context
}

// captureStackTrace 捕获堆栈跟踪
func (deh *DynamicErrorHandler) captureStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// callErrorCallbacks 调用错误回调
func (deh *DynamicErrorHandler) callErrorCallbacks(pluginID string, record ErrorRecord) error {
	callbacks, exists := deh.errorCallbacks[pluginID]
	if !exists || len(callbacks) == 0 {
		return nil
	}
	
	for _, callback := range callbacks {
		if err := callback(pluginID, record); err != nil {
			return err
		}
	}
	
	return nil
}

// attemptRecovery 尝试恢复
func (deh *DynamicErrorHandler) attemptRecovery(ctx context.Context, pluginID string, record ErrorRecord) error {
	deh.mutex.Lock()
	defer deh.mutex.Unlock()
	
	stats := deh.errorStats[pluginID]
	if stats == nil {
		return fmt.Errorf("no stats found for plugin: %s", pluginID)
	}
	
	if stats.RecoveryCount >= deh.config.MaxRecoveryAttempts {
		return fmt.Errorf("max recovery attempts reached for plugin: %s", pluginID)
	}
	
	// 这里应该实现具体的恢复逻辑
	// 例如重启插件、重置状态等
	stats.RecoveryCount++
	
	return nil
}

// calculateErrorRate 计算错误率
func (deh *DynamicErrorHandler) calculateErrorRate(pluginID string) float64 {
	deh.mutex.RLock()
	defer deh.mutex.RUnlock()
	return deh.calculateErrorRateInternal(pluginID)
}

// calculateErrorRateInternal 计算错误率（内部方法，不获取锁）
func (deh *DynamicErrorHandler) calculateErrorRateInternal(pluginID string) float64 {
	history, exists := deh.errorHistory[pluginID]
	if !exists || len(history) == 0 {
		return 0.0
	}
	
	now := time.Now()
	windowStart := now.Add(-deh.config.ErrorRateWindow)
	
	count := 0
	for _, record := range history {
		if record.Timestamp.After(windowStart) {
			count++
		}
	}
	
	// 返回每分钟的错误数
	minutes := deh.config.ErrorRateWindow.Minutes()
	return float64(count) / minutes
}

// DynamicResourceManager 方法实现

// MonitorResources 监控资源使用
func (drm *DynamicResourceManager) MonitorResources(ctx context.Context, pluginID string) error {
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}
	
	drm.mutex.Lock()
	defer drm.mutex.Unlock()
	
	// 如果已经在监控，先停止
	if cancel, exists := drm.monitoringContexts[pluginID]; exists {
		cancel()
	}
	
	// 创建新的监控上下文
	monitorCtx, cancel := context.WithCancel(ctx)
	drm.monitoringContexts[pluginID] = cancel
	
	// 启动监控协程
	go drm.monitorResourcesLoop(monitorCtx, pluginID)
	
	return nil
}

// GetResourceUsage 获取资源使用情况
func (drm *DynamicResourceManager) GetResourceUsage(pluginID string) (*ResourceUsage, error) {
	if pluginID == "" {
		return nil, fmt.Errorf("plugin ID cannot be empty")
	}
	
	drm.mutex.RLock()
	defer drm.mutex.RUnlock()
	
	usage, exists := drm.resourceUsage[pluginID]
	if !exists {
		return nil, fmt.Errorf("no resource usage found for plugin: %s", pluginID)
	}
	
	// 返回使用情况的副本
	usageCopy := *usage
	return &usageCopy, nil
}

// SetResourceLimits 设置资源限制
func (drm *DynamicResourceManager) SetResourceLimits(pluginID string, limits *ResourceLimits) error {
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}
	
	if limits == nil {
		return fmt.Errorf("limits cannot be nil")
	}
	
	drm.mutex.Lock()
	defer drm.mutex.Unlock()
	
	drm.resourceLimits[pluginID] = limits
	return nil
}

// CheckResourceLimits 检查资源限制
func (drm *DynamicResourceManager) CheckResourceLimits(pluginID string) error {
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}
	
	drm.mutex.RLock()
	usage, hasUsage := drm.resourceUsage[pluginID]
	limits, hasLimits := drm.resourceLimits[pluginID]
	drm.mutex.RUnlock()
	
	if !hasLimits {
		return nil // 没有限制，不检查
	}
	
	if !hasUsage {
		return fmt.Errorf("no resource usage data for plugin: %s", pluginID)
	}
	
	// 检查内存限制
	if limits.MaxMemory > 0 {
		if usage.MemoryUsage > limits.MaxMemory {
			return fmt.Errorf("memory usage exceeded: %d > %d", usage.MemoryUsage, limits.MaxMemory)
		}
	}
	
	// 检查CPU限制
	if limits.MaxCPU > 0 {
		if usage.CPUUsage > limits.MaxCPU {
			return fmt.Errorf("CPU usage exceeded: %.2f > %.2f", usage.CPUUsage, limits.MaxCPU)
		}
	}
	
	// 检查协程限制
	if limits.MaxGoroutines > 0 {
		if usage.GoroutineCount > limits.MaxGoroutines {
			return fmt.Errorf("goroutine count exceeded: %d > %d", usage.GoroutineCount, limits.MaxGoroutines)
		}
	}
	
	return nil
}

// CleanupResources 清理资源
func (drm *DynamicResourceManager) CleanupResources(ctx context.Context, pluginID string) error {
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}
	
	drm.mutex.Lock()
	defer drm.mutex.Unlock()
	
	// 停止监控
	if cancel, exists := drm.monitoringContexts[pluginID]; exists {
		cancel()
		delete(drm.monitoringContexts, pluginID)
	}
	
	// 清理资源使用记录
	delete(drm.resourceUsage, pluginID)
	delete(drm.resourceLimits, pluginID)
	
	return nil
}

// ForceCleanup 强制清理资源
func (drm *DynamicResourceManager) ForceCleanup(pluginID string) error {
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}
	
	drm.mutex.Lock()
	defer drm.mutex.Unlock()
	
	// 强制停止监控
	if cancel, exists := drm.monitoringContexts[pluginID]; exists {
		cancel()
		delete(drm.monitoringContexts, pluginID)
	}
	
	// 强制清理所有相关数据
	delete(drm.resourceUsage, pluginID)
	delete(drm.resourceLimits, pluginID)
	
	// 触发垃圾回收
	runtime.GC()
	
	return nil
}

// monitorResourcesLoop 资源监控循环
func (drm *DynamicResourceManager) monitorResourcesLoop(ctx context.Context, pluginID string) {
	ticker := time.NewTicker(drm.config.MonitorInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			drm.collectResourceUsage(pluginID)
		}
	}
}

// collectResourceUsage 收集资源使用情况
func (drm *DynamicResourceManager) collectResourceUsage(pluginID string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	usage := &ResourceUsage{
		MemoryUsage:        int64(m.Alloc),
		CPUUsage:          0.0, // 简化实现，实际应该计算CPU使用率
		GoroutineCount:    runtime.NumGoroutine(),
		FileDescriptors:   0, // 简化实现
		NetworkConnections: 0, // 简化实现
		LastUpdated:       time.Now(),
		History:           make([]ResourceSnapshot, 0),
	}
	
	drm.mutex.Lock()
	defer drm.mutex.Unlock()
	
	// 更新资源使用记录
	drm.resourceUsage[pluginID] = usage
	
	// 添加历史快照
	if existing, exists := drm.resourceUsage[pluginID]; exists {
		snapshot := ResourceSnapshot{
			Timestamp:      time.Now(),
			MemoryUsage:    usage.MemoryUsage,
			CPUUsage:       usage.CPUUsage,
			GoroutineCount: usage.GoroutineCount,
		}
		
		existing.History = append(existing.History, snapshot)
		
		// 限制历史记录大小
		if len(existing.History) > drm.config.MaxHistorySize {
			existing.History = existing.History[1:]
		}
	}
}