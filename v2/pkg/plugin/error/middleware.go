package plugin

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

// ErrorMiddleware 错误处理中间件接口
type ErrorMiddleware interface {
	// Handle 处理错误
	Handle(ctx context.Context, next ErrorHandler, err error) error
	// GetName 获取中间件名称
	GetName() string
	// GetPriority 获取优先级
	GetPriority() int
}

// ErrorHandler 错误处理器函数类型
type ErrorHandler func(ctx context.Context, err error) error

// MiddlewareChain 中间件链
type MiddlewareChain struct {
	middlewares []ErrorMiddleware
	mutex       sync.RWMutex
	logger      Logger
}

// NewMiddlewareChain 创建中间件链
func NewMiddlewareChain(logger Logger) *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]ErrorMiddleware, 0),
		logger:      logger,
	}
}

// Add 添加中间件
func (mc *MiddlewareChain) Add(middleware ErrorMiddleware) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	// 按优先级插入中间件
	inserted := false
	for i, existing := range mc.middlewares {
		if middleware.GetPriority() < existing.GetPriority() {
			mc.middlewares = append(mc.middlewares[:i], append([]ErrorMiddleware{middleware}, mc.middlewares[i:]...)...)
			inserted = true
			break
		}
	}
	
	if !inserted {
		mc.middlewares = append(mc.middlewares, middleware)
	}
	
	if mc.logger != nil {
		mc.logger.Info("Error middleware added", map[string]interface{}{
			"name": middleware.GetName(),
			"priority": middleware.GetPriority(),
		})
	}
}

// Remove 移除中间件
func (mc *MiddlewareChain) Remove(name string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	for i, middleware := range mc.middlewares {
		if middleware.GetName() == name {
			mc.middlewares = append(mc.middlewares[:i], mc.middlewares[i+1:]...)
			if mc.logger != nil {
				mc.logger.Info("Error middleware removed", map[string]interface{}{
					"name": name,
				})
			}
			break
		}
	}
}

// Execute 执行中间件链
func (mc *MiddlewareChain) Execute(ctx context.Context, err error, finalHandler ErrorHandler) error {
	mc.mutex.RLock()
	middlewares := make([]ErrorMiddleware, len(mc.middlewares))
	copy(middlewares, mc.middlewares)
	mc.mutex.RUnlock()
	
	// 构建中间件链
	handler := finalHandler
	for i := len(middlewares) - 1; i >= 0; i-- {
		middleware := middlewares[i]
		currentHandler := handler
		handler = func(ctx context.Context, err error) error {
			return middleware.Handle(ctx, currentHandler, err)
		}
	}
	
	return handler(ctx, err)
}

// GetMiddlewares 获取所有中间件
func (mc *MiddlewareChain) GetMiddlewares() []ErrorMiddleware {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	middlewares := make([]ErrorMiddleware, len(mc.middlewares))
	copy(middlewares, mc.middlewares)
	return middlewares
}

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	logger   Logger
	priority int
}

// NewLoggingMiddleware 创建日志中间件
func NewLoggingMiddleware(logger Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger:   logger,
		priority: 1000, // 高优先级，最后执行
	}
}

// Handle 处理错误
func (lm *LoggingMiddleware) Handle(ctx context.Context, next ErrorHandler, err error) error {
	start := time.Now()
	
	// 记录错误信息
	if lm.logger != nil {
		if pluginErr, ok := err.(*BasePluginError); ok {
			lm.logger.Error("Plugin error occurred", map[string]interface{}{
				"error_code": pluginErr.GetCode().String(),
				"error_type": pluginErr.GetType().String(),
				"severity": pluginErr.GetSeverity().String(),
				"message": pluginErr.Error(),
				"context": pluginErr.GetContext(),
				"timestamp": pluginErr.GetTimestamp(),
			})
		} else {
			lm.logger.Error("Error occurred", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}
	
	// 调用下一个处理器
	handleErr := next(ctx, err)
	
	// 记录处理时间
	duration := time.Since(start)
	if lm.logger != nil {
		lm.logger.Debug("Error handling completed", map[string]interface{}{
			"duration": duration,
			"success": handleErr == nil,
		})
	}
	
	return handleErr
}

// GetName 获取中间件名称
func (lm *LoggingMiddleware) GetName() string {
	return "logging"
}

// GetPriority 获取优先级
func (lm *LoggingMiddleware) GetPriority() int {
	return lm.priority
}

// MetricsMiddleware 指标中间件
type MetricsMiddleware struct {
	metrics  MetricsCollector
	priority int
}

// NewMetricsMiddleware 创建指标中间件
func NewMetricsMiddleware(metrics MetricsCollector) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics:  metrics,
		priority: 900, // 高优先级
	}
}

// Handle 处理错误
func (mm *MetricsMiddleware) Handle(ctx context.Context, next ErrorHandler, err error) error {
	start := time.Now()
	
	// 记录错误指标
	if mm.metrics != nil {
		labels := map[string]string{
			"error_type": "unknown",
		}
		
		if pluginErr, ok := err.(*BasePluginError); ok {
			labels["error_type"] = pluginErr.GetType().String()
			labels["error_code"] = pluginErr.GetCode().String()
			labels["severity"] = pluginErr.GetSeverity().String()
		}
		
		mm.metrics.IncrementCounter("error_handling_total", labels)
	}
	
	// 调用下一个处理器
	handleErr := next(ctx, err)
	
	// 记录处理时间
	duration := time.Since(start)
	if mm.metrics != nil {
		labels := map[string]string{
			"success": fmt.Sprintf("%t", handleErr == nil),
		}
		mm.metrics.RecordTimer("error_handling_duration", duration, labels)
		
		if handleErr == nil {
			mm.metrics.IncrementCounter("error_handling_success_total", labels)
		} else {
			mm.metrics.IncrementCounter("error_handling_failure_total", labels)
		}
	}
	
	return handleErr
}

// GetName 获取中间件名称
func (mm *MetricsMiddleware) GetName() string {
	return "metrics"
}

// GetPriority 获取优先级
func (mm *MetricsMiddleware) GetPriority() int {
	return mm.priority
}

// ContextMiddleware 上下文中间件
type ContextMiddleware struct {
	priority int
}

// NewContextMiddleware 创建上下文中间件
func NewContextMiddleware() *ContextMiddleware {
	return &ContextMiddleware{
		priority: 100, // 低优先级，最先执行
	}
}

// Handle 处理错误
func (cm *ContextMiddleware) Handle(ctx context.Context, next ErrorHandler, err error) error {
	// 增强错误上下文信息
	if pluginErr, ok := err.(*BasePluginError); ok {
		// 添加请求上下文信息
		if requestID := ctx.Value("request_id"); requestID != nil {
			pluginErr.WithContext("request_id", requestID)
		}
		
		if userID := ctx.Value("user_id"); userID != nil {
			pluginErr.WithContext("user_id", userID)
		}
		
		if pluginID := ctx.Value("plugin_id"); pluginID != nil {
			pluginErr.WithContext("plugin_id", pluginID)
		}
		
		// 添加调用堆栈信息
		if pluginErr.GetStackTrace() == "" {
			pluginErr.WithStackTrace(string(debug.Stack()))
		}
		
		// 添加时间戳
		pluginErr.WithContext("handled_at", time.Now())
	}
	
	return next(ctx, err)
}

// GetName 获取中间件名称
func (cm *ContextMiddleware) GetName() string {
	return "context"
}

// GetPriority 获取优先级
func (cm *ContextMiddleware) GetPriority() int {
	return cm.priority
}

// RetryMiddleware 重试中间件
type RetryMiddleware struct {
	retryExecutor *RetryExecutor
	priority      int
}

// NewRetryMiddleware 创建重试中间件
func NewRetryMiddleware(retryExecutor *RetryExecutor) *RetryMiddleware {
	return &RetryMiddleware{
		retryExecutor: retryExecutor,
		priority:      200, // 中等优先级
	}
}

// Handle 处理错误
func (rm *RetryMiddleware) Handle(ctx context.Context, next ErrorHandler, err error) error {
	// 检查错误是否可重试
	if pluginErr, ok := err.(*BasePluginError); ok && pluginErr.IsRetryable() {
		// 使用重试执行器处理
		return rm.retryExecutor.Execute(ctx, func() error {
			return next(ctx, err)
		})
	}
	
	// 不可重试的错误直接传递给下一个处理器
	return next(ctx, err)
}

// GetName 获取中间件名称
func (rm *RetryMiddleware) GetName() string {
	return "retry"
}

// GetPriority 获取优先级
func (rm *RetryMiddleware) GetPriority() int {
	return rm.priority
}

// CircuitBreakerMiddleware 熔断器中间件
type CircuitBreakerMiddleware struct {
	circuitBreaker CircuitBreaker
	priority       int
}

// NewCircuitBreakerMiddleware 创建熔断器中间件
func NewCircuitBreakerMiddleware(circuitBreaker CircuitBreaker) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		circuitBreaker: circuitBreaker,
		priority:       300, // 中等优先级
	}
}

// Handle 处理错误
func (cbm *CircuitBreakerMiddleware) Handle(ctx context.Context, next ErrorHandler, err error) error {
	// 通过熔断器执行错误处理
	return cbm.circuitBreaker.Execute(ctx, func() error {
		return next(ctx, err)
	})
}

// GetName 获取中间件名称
func (cbm *CircuitBreakerMiddleware) GetName() string {
	return "circuit_breaker"
}

// GetPriority 获取优先级
func (cbm *CircuitBreakerMiddleware) GetPriority() int {
	return cbm.priority
}

// AlertMiddleware 告警中间件
type AlertMiddleware struct {
	alertManager AlertManager
	priority     int
}

// NewAlertMiddleware 创建告警中间件
func NewAlertMiddleware(alertManager AlertManager) *AlertMiddleware {
	return &AlertMiddleware{
		alertManager: alertManager,
		priority:     800, // 高优先级
	}
}

// Handle 处理错误
func (am *AlertMiddleware) Handle(ctx context.Context, next ErrorHandler, err error) error {
	// 检查是否需要触发告警
	if pluginErr, ok := err.(*BasePluginError); ok {
		// 根据错误严重程度决定是否触发告警
		if pluginErr.GetSeverity() >= ErrorSeverityError {
			// 构建告警数据
			alertData := map[string]interface{}{
				"error_code":    pluginErr.GetCode().String(),
				"error_type":    pluginErr.GetType().String(),
				"severity":      pluginErr.GetSeverity().String(),
				"message":       pluginErr.Error(),
				"timestamp":     pluginErr.GetTimestamp(),
				"context":       pluginErr.GetContext(),
				"stack_trace":   pluginErr.GetStackTrace(),
			}
			
			// 添加上下文信息
			if pluginID := ctx.Value("plugin_id"); pluginID != nil {
				alertData["plugin_id"] = pluginID
			}
			
			// 触发告警（异步）
			go func() {
				// 根据错误类型选择告警规则
				ruleID := am.getAlertRuleID(pluginErr)
				if ruleID != "" {
					am.alertManager.TriggerAlert(ruleID, alertData)
				}
			}()
		}
	}
	
	return next(ctx, err)
}

// getAlertRuleID 获取告警规则ID
func (am *AlertMiddleware) getAlertRuleID(err *BasePluginError) string {
	// 根据错误类型和严重程度选择告警规则
	switch err.GetSeverity() {
	case ErrorSeverityCritical:
		return "critical_error_alert"
	case ErrorSeverityFatal:
		return "fatal_error_alert"
	case ErrorSeverityError:
		return "error_alert"
	default:
		return ""
	}
}

// GetName 获取中间件名称
func (am *AlertMiddleware) GetName() string {
	return "alert"
}

// GetPriority 获取优先级
func (am *AlertMiddleware) GetPriority() int {
	return am.priority
}

// RecoveryMiddleware 恢复中间件
type RecoveryMiddleware struct {
	recoveryManager *RecoveryManager
	priority        int
}

// NewRecoveryMiddleware 创建恢复中间件
func NewRecoveryMiddleware(recoveryManager *RecoveryManager) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		recoveryManager: recoveryManager,
		priority:        400, // 中等优先级
	}
}

// Handle 处理错误
func (rm *RecoveryMiddleware) Handle(ctx context.Context, next ErrorHandler, err error) error {
	// 尝试恢复
	if pluginID := ctx.Value("plugin_id"); pluginID != nil {
		if pluginIDStr, ok := pluginID.(string); ok {
			// 异步执行恢复操作
			go func() {
				recoveryCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				
				rm.recoveryManager.Recover(recoveryCtx, pluginIDStr, err)
			}()
		}
	}
	
	return next(ctx, err)
}

// GetName 获取中间件名称
func (rm *RecoveryMiddleware) GetName() string {
	return "recovery"
}

// GetPriority 获取优先级
func (rm *RecoveryMiddleware) GetPriority() int {
	return rm.priority
}

// ErrorMiddlewareManager 错误中间件管理器
type ErrorMiddlewareManager struct {
	chain  *MiddlewareChain
	logger Logger
}

// NewErrorMiddlewareManager 创建错误中间件管理器
func NewErrorMiddlewareManager(logger Logger) *ErrorMiddlewareManager {
	return &ErrorMiddlewareManager{
		chain:  NewMiddlewareChain(logger),
		logger: logger,
	}
}

// RegisterDefaultMiddlewares 注册默认中间件
func (emm *ErrorMiddlewareManager) RegisterDefaultMiddlewares(metrics MetricsCollector, alertManager AlertManager, recoveryManager *RecoveryManager, retryExecutor *RetryExecutor, circuitBreaker CircuitBreaker) {
	// 按优先级顺序添加中间件
	emm.chain.Add(NewContextMiddleware())
	emm.chain.Add(NewRetryMiddleware(retryExecutor))
	emm.chain.Add(NewCircuitBreakerMiddleware(circuitBreaker))
	emm.chain.Add(NewRecoveryMiddleware(recoveryManager))
	emm.chain.Add(NewAlertMiddleware(alertManager))
	emm.chain.Add(NewMetricsMiddleware(metrics))
	emm.chain.Add(NewLoggingMiddleware(emm.logger))
}

// AddMiddleware 添加中间件
func (emm *ErrorMiddlewareManager) AddMiddleware(middleware ErrorMiddleware) {
	emm.chain.Add(middleware)
}

// RemoveMiddleware 移除中间件
func (emm *ErrorMiddlewareManager) RemoveMiddleware(name string) {
	emm.chain.Remove(name)
}

// HandleError 处理错误
func (emm *ErrorMiddlewareManager) HandleError(ctx context.Context, err error, finalHandler ErrorHandler) error {
	return emm.chain.Execute(ctx, err, finalHandler)
}

// GetMiddlewares 获取所有中间件
func (emm *ErrorMiddlewareManager) GetMiddlewares() []ErrorMiddleware {
	return emm.chain.GetMiddlewares()
}