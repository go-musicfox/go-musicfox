package plugin

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// MockLogger 模拟日志记录器
type MockLogger struct {
	logs []string
}

func (ml *MockLogger) Debug(msg string, fields ...map[string]interface{}) {
	var allFields map[string]interface{}
	if len(fields) > 0 {
		allFields = fields[0]
	}
	ml.logs = append(ml.logs, fmt.Sprintf("DEBUG: %s %v", msg, allFields))
}

func (ml *MockLogger) Info(msg string, fields map[string]interface{}) {
	ml.logs = append(ml.logs, fmt.Sprintf("INFO: %s %v", msg, fields))
}

func (ml *MockLogger) Warn(msg string, fields map[string]interface{}) {
	ml.logs = append(ml.logs, fmt.Sprintf("WARN: %s %v", msg, fields))
}

func (ml *MockLogger) Error(msg string, fields map[string]interface{}) {
	ml.logs = append(ml.logs, fmt.Sprintf("ERROR: %s %v", msg, fields))
}

func (ml *MockLogger) Trace(msg string, fields map[string]interface{}) {
	ml.logs = append(ml.logs, fmt.Sprintf("TRACE: %s %v", msg, fields))
}

func (ml *MockLogger) Fatal(msg string, fields map[string]interface{}) {
	ml.logs = append(ml.logs, fmt.Sprintf("FATAL: %s %v", msg, fields))
}

func (ml *MockLogger) LogError(ctx context.Context, err PluginError, pluginID string) {
	ml.logs = append(ml.logs, fmt.Sprintf("LOG_ERROR: %s %s %v", pluginID, err.Error(), err))
}

func (ml *MockLogger) LogErrorWithContext(ctx context.Context, err PluginError, pluginID string, additionalFields map[string]interface{}) {
	ml.logs = append(ml.logs, fmt.Sprintf("LOG_ERROR_WITH_CONTEXT: %s %s %v %v", pluginID, err.Error(), err, additionalFields))
}

func (ml *MockLogger) SetLevel(level LogLevel) {
	// Mock implementation
}

func (ml *MockLogger) GetLevel() LogLevel {
	return LogLevelDebug
}

func (ml *MockLogger) GetLogs() []string {
	return ml.logs
}

func (ml *MockLogger) Clear() {
	ml.logs = nil
}

// MockEventBus 模拟事件总线
type MockEventBus struct {
	events []MockEvent
}

type MockEvent struct {
	Type string
	Data interface{}
}

func (meb *MockEventBus) Publish(eventType string, data interface{}) error {
	meb.events = append(meb.events, MockEvent{Type: eventType, Data: data})
	return nil
}

func (meb *MockEventBus) Subscribe(eventType string, handler interface{}) error {
	return nil
}

func (meb *MockEventBus) Unsubscribe(eventType string, handler interface{}) error {
	return nil
}

func (meb *MockEventBus) GetEvents() []MockEvent {
	return meb.events
}

func (meb *MockEventBus) Clear() {
	meb.events = nil
}

// MockPluginLoader 模拟插件加载器
type MockPluginLoader struct {
	reloadCalls  []string
	restartCalls []string
	failReload   map[string]bool
	failRestart  map[string]bool
}

func NewMockPluginLoader() *MockPluginLoader {
	return &MockPluginLoader{
		failReload:  make(map[string]bool),
		failRestart: make(map[string]bool),
	}
}

func (mpl *MockPluginLoader) ReloadPlugin(pluginID string) error {
	mpl.reloadCalls = append(mpl.reloadCalls, pluginID)
	if mpl.failReload[pluginID] {
		return fmt.Errorf("failed to reload plugin %s", pluginID)
	}
	return nil
}

func (mpl *MockPluginLoader) RestartPlugin(pluginID string) error {
	mpl.restartCalls = append(mpl.restartCalls, pluginID)
	if mpl.failRestart[pluginID] {
		return fmt.Errorf("failed to restart plugin %s", pluginID)
	}
	return nil
}

func (mpl *MockPluginLoader) SetReloadFailure(pluginID string, fail bool) {
	mpl.failReload[pluginID] = fail
}

func (mpl *MockPluginLoader) SetRestartFailure(pluginID string, fail bool) {
	mpl.failRestart[pluginID] = fail
}

// TestEnhancedErrorTypes 测试增强的错误类型
func TestEnhancedErrorTypes(t *testing.T) {
	// 测试基础错误创建
	err := NewPluginErrorWithCause(ErrorCodePluginCrashed, "Plugin crashed unexpectedly", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if err.GetCode() != ErrorCodePluginCrashed {
		t.Errorf("Expected error code %v, got %v", ErrorCodePluginCrashed, err.GetCode())
	}
	
	if err.GetType() != ErrorTypeSystem {
		t.Errorf("Expected error type %v, got %v", ErrorTypeSystem, err.GetType())
	}
	
	if err.GetSeverity() != ErrorSeverityCritical {
		t.Errorf("Expected severity %v, got %v", ErrorSeverityCritical, err.GetSeverity())
	}
	
	// 测试错误上下文
	err.WithContext("plugin_id", "test-plugin")
	err.WithContext("user_id", "user123")
	
	context := err.GetContext()
	if context["plugin_id"] != "test-plugin" {
		t.Errorf("Expected plugin_id 'test-plugin', got %v", context["plugin_id"])
	}
	
	// 测试重试配置
	err.WithRetryConfig(true, 5*time.Second)
	if !err.IsRetryable() {
		t.Error("Expected error to be retryable")
	}
	
	if err.GetRetryAfter() != 5*time.Second {
		t.Errorf("Expected retry after 5s, got %v", err.GetRetryAfter())
	}
	
	// 测试堆栈跟踪
	if err.GetStackTrace() == "" {
		t.Error("Expected stack trace to be set")
	}
}

// TestMusicFoxErrorCompatibility 测试MusicFox错误兼容性
func TestMusicFoxErrorCompatibility(t *testing.T) {
	err := NewMusicFoxError(ErrorCodeAudioDecodeFailed, "Audio playback failed", "Device not available", "AudioPlugin")
	
	if err.GetCode() != ErrorCodeAudioDecodeFailed {
		t.Errorf("Expected error code %v, got %v", ErrorCodeAudioDecodeFailed, err.GetCode())
	}
	
	context := err.GetContext()
	if context["details"] != "Device not available" {
		t.Errorf("Expected details 'Device not available', got %v", context["details"])
	}
	
	if context["source"] != "AudioPlugin" {
		t.Errorf("Expected source 'AudioPlugin', got %v", context["source"])
	}
}

// TestErrorWrapping 测试错误包装
func TestErrorWrapping(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, ErrorCodePluginIOError, "IO operation failed")
	
	if wrappedErr == nil {
		t.Fatal("Expected wrapped error, got nil")
	}
	
	if wrappedErr.GetCause() != originalErr {
		t.Errorf("Expected cause to be original error, got %v", wrappedErr.GetCause())
	}
	
	// 测试errors.Unwrap
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Expected wrapped error to match original error with errors.Is")
	}
	
	// 测试errors.As
	var pluginErr *BasePluginError
	if !errors.As(wrappedErr, &pluginErr) {
		t.Error("Expected wrapped error to be assignable to BasePluginError")
	}
}

// TestSmartRetryStrategy 测试智能重试策略
func TestSmartRetryStrategy(t *testing.T) {
	config := DefaultRetryConfig()
	config.MaxAttempts = 3
	config.BaseDelay = 100 * time.Millisecond
	
	strategy := NewSmartRetryStrategy(config)
	
	// 测试可重试错误
	retryableErr := NewPluginError(ErrorCodeUnavailable, "Service unavailable")
	if !strategy.ShouldRetry(0, retryableErr) {
		t.Error("Expected retryable error to be retryable")
	}
	
	// 测试不可重试错误
	nonRetryableErr := NewPluginError(ErrorCodeInvalidArgument, "Invalid argument")
	if strategy.ShouldRetry(0, nonRetryableErr) {
		t.Error("Expected non-retryable error to not be retryable")
	}
	
	// 测试最大重试次数
	if strategy.ShouldRetry(3, retryableErr) {
		t.Error("Expected retry to be rejected after max attempts")
	}
	
	// 测试延迟计算
	delay := strategy.GetDelay(0)
	// 由于抖动，延迟可能略小于基础延迟，允许一定的容差
	minDelay := time.Duration(float64(config.BaseDelay) * 0.8) // 允许20%的容差
	if delay < minDelay {
		t.Errorf("Expected delay >= %v, got %v", minDelay, delay)
	}
}

// TestRetryExecutor 测试重试执行器
func TestRetryExecutor(t *testing.T) {
	logger := &MockLogger{}
	metrics := NewMetricsCollector()
	strategy := NewSmartRetryStrategy(DefaultRetryConfig())
	executor := NewRetryExecutor(strategy, logger, metrics)
	
	// 测试成功执行
	ctx := context.Background()
	executeCount := 0
	err := executor.Execute(ctx, func() error {
		executeCount++
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if executeCount != 1 {
		t.Errorf("Expected 1 execution, got %d", executeCount)
	}
	
	// 测试重试执行
	executeCount = 0
	failCount := 2
	err = executor.Execute(ctx, func() error {
		executeCount++
		if executeCount <= failCount {
			return NewPluginError(ErrorCodeUnavailable, "Service unavailable")
		}
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}
	
	if executeCount != failCount+1 {
		t.Errorf("Expected %d executions, got %d", failCount+1, executeCount)
	}
}

// TestSmartCircuitBreaker 测试智能熔断器
func TestSmartCircuitBreaker(t *testing.T) {
	logger := &MockLogger{}
	metrics := NewMetricsCollector()
	eventBus := &MockEventBus{}
	
	config := DefaultCircuitBreakerConfig()
	config.FailureThreshold = 3
	config.SuccessThreshold = 2
	
	cb := NewSmartCircuitBreaker(config, logger, metrics, eventBus)
	
	// 测试初始状态
	if cb.GetState() != CircuitBreakerStateClosed {
		t.Errorf("Expected initial state to be closed, got %v", cb.GetState())
	}
	
	// 测试失败计数
	ctx := context.Background()
	for i := 0; i < config.FailureThreshold; i++ {
		cb.Execute(ctx, func() error {
			return errors.New("test error")
		})
	}
	
	// 应该打开熔断器
	if cb.GetState() != CircuitBreakerStateOpen {
		t.Errorf("Expected state to be open after failures, got %v", cb.GetState())
	}
	
	// 测试熔断器打开时的行为
	err := cb.Execute(ctx, func() error {
		return nil
	})
	
	if err == nil {
		t.Error("Expected error when circuit breaker is open")
	}
	
	// 测试统计信息
	stats := cb.GetStats()
	if stats.State != CircuitBreakerStateOpen {
		t.Errorf("Expected stats state to be open, got %v", stats.State)
	}
	
	if stats.FailedRequests < int64(config.FailureThreshold) {
		t.Errorf("Expected at least %d failed requests, got %d", config.FailureThreshold, stats.FailedRequests)
	}
}

// TestRecoveryStrategies 测试恢复策略
func TestRecoveryStrategies(t *testing.T) {
	logger := &MockLogger{}
	metrics := NewMetricsCollector()
	eventBus := &MockEventBus{}
	pluginLoader := NewMockPluginLoader()
	
	// 测试重启恢复策略
	restartStrategy := NewRestartRecoveryStrategy(3, 100*time.Millisecond, logger, metrics, eventBus, pluginLoader)
	
	crashedErr := NewPluginError(ErrorCodePluginCrashed, "Plugin crashed")
	if !restartStrategy.CanRecover(crashedErr) {
		t.Error("Expected restart strategy to handle crashed plugin")
	}
	
	ctx := context.Background()
	err := restartStrategy.Recover(ctx, "test-plugin", crashedErr)
	if err != nil {
		t.Errorf("Expected successful recovery, got %v", err)
	}
	
	if len(pluginLoader.restartCalls) != 1 {
		t.Errorf("Expected 1 restart call, got %d", len(pluginLoader.restartCalls))
	}
	
	// 测试重载恢复策略
	reloadStrategy := NewReloadRecoveryStrategy(3, 100*time.Millisecond, logger, metrics, eventBus, pluginLoader)
	
	loadErr := NewPluginError(ErrorCodePluginInitFailed, "Plugin load failed")
	if !reloadStrategy.CanRecover(loadErr) {
		t.Error("Expected reload strategy to handle load failure")
	}
	
	err = reloadStrategy.Recover(ctx, "test-plugin", loadErr)
	if err != nil {
		t.Errorf("Expected successful recovery, got %v", err)
	}
	
	if len(pluginLoader.reloadCalls) != 1 {
		t.Errorf("Expected 1 reload call, got %d", len(pluginLoader.reloadCalls))
	}
}

// TestRecoveryManager 测试恢复管理器
func TestRecoveryManager(t *testing.T) {
	logger := &MockLogger{}
	metrics := NewMetricsCollector()
	eventBus := &MockEventBus{}
	
	manager := NewRecoveryManager(logger, metrics, eventBus)
	
	// 创建插件加载器
	pluginLoader := NewMockPluginLoader()
	
	// 添加恢复策略
	restartStrategy := NewRestartRecoveryStrategy(3, 100*time.Millisecond, logger, metrics, eventBus, pluginLoader)
	reloadStrategy := NewReloadRecoveryStrategy(3, 100*time.Millisecond, logger, metrics, eventBus, pluginLoader)
	fallbackStrategy := NewFallbackRecoveryStrategy(map[string]string{"test-plugin": "fallback-plugin"}, nil, logger, metrics, eventBus, pluginLoader)
	
	manager.AddStrategy(restartStrategy)
	manager.AddStrategy(reloadStrategy)
	manager.AddStrategy(fallbackStrategy)
	
	// 测试策略排序（按优先级）
	strategies := manager.GetStrategies()
	if len(strategies) != 3 {
		t.Errorf("Expected 3 strategies, got %d", len(strategies))
	}
	
	// 测试恢复执行
	ctx := context.Background()
	crashedErr := NewPluginError(ErrorCodePluginCrashed, "Plugin crashed")
	err := manager.Recover(ctx, "test-plugin", crashedErr)
	if err != nil {
		t.Errorf("Expected successful recovery, got %v", err)
	}
}

// TestAlertManager 测试告警管理器
func TestAlertManager(t *testing.T) {
	logger := &MockLogger{}
	metrics := NewMetricsCollector()
	eventBus := &MockEventBus{}
	
	alertManager := NewSmartAlertManager(logger, metrics, eventBus)
	
	// 注册告警规则
	rule := &AlertRule{
		ID:          "test-rule",
		Name:        "Test Alert Rule",
		Description: "Test alert for errors",
		Condition:   "error_rate",
		Threshold:   0.1,
		Duration:    5 * time.Minute,
		Severity:    AlertSeverityHigh,
		Enabled:     true,
		Actions: []AlertAction{
			{
				Type:    AlertActionTypeLog,
				Enabled: true,
			},
		},
	}
	
	err := alertManager.RegisterAlert(rule)
	if err != nil {
		t.Errorf("Expected no error registering alert, got %v", err)
	}
	
	// 触发告警
	alertData := map[string]interface{}{
		"plugin_id": "test-plugin",
		"error_rate": 0.15,
	}
	
	err = alertManager.TriggerAlert("test-rule", alertData)
	if err != nil {
		t.Errorf("Expected no error triggering alert, got %v", err)
	}
	
	// 检查活跃告警
	activeAlerts := alertManager.GetActiveAlerts()
	if len(activeAlerts) != 1 {
		t.Errorf("Expected 1 active alert, got %d", len(activeAlerts))
	}
	
	// 解决告警
	alert := activeAlerts[0]
	err = alertManager.ResolveAlert(alert.ID, "test-user")
	if err != nil {
		t.Errorf("Expected no error resolving alert, got %v", err)
	}
	
	// 检查告警状态
	resolvedAlert, err := alertManager.GetAlert(alert.ID)
	if err != nil {
		t.Errorf("Expected no error getting alert, got %v", err)
	}
	
	if !resolvedAlert.Resolved {
		t.Error("Expected alert to be resolved")
	}
}

// TestErrorMiddleware 测试错误中间件
func TestErrorMiddleware(t *testing.T) {
	logger := &MockLogger{}
	metrics := NewMetricsCollector()
	
	// 创建中间件链
	chain := NewMiddlewareChain(logger)
	
	// 添加中间件
	chain.Add(NewContextMiddleware())
	chain.Add(NewLoggingMiddleware(logger))
	chain.Add(NewMetricsMiddleware(metrics))
	
	// 测试中间件执行
	ctx := context.WithValue(context.Background(), "plugin_id", "test-plugin")
	testErr := NewPluginError(ErrorCodePluginCrashed, "Test error")
	
	finalHandlerCalled := false
	finalHandler := func(ctx context.Context, err error) error {
		finalHandlerCalled = true
		return nil
	}
	
	err := chain.Execute(ctx, testErr, finalHandler)
	if err != nil {
		t.Errorf("Expected no error from middleware chain, got %v", err)
	}
	
	if !finalHandlerCalled {
		t.Error("Expected final handler to be called")
	}
	
	// 检查日志记录
	logs := logger.GetLogs()
	if len(logs) == 0 {
		t.Error("Expected logs to be recorded")
	}
	
	// 检查上下文增强
	context := testErr.GetContext()
	if context["plugin_id"] != "test-plugin" {
		t.Errorf("Expected plugin_id to be set in context, got %v", context["plugin_id"])
	}
}

// TestErrorMiddlewareManager 测试错误中间件管理器
func TestErrorMiddlewareManager(t *testing.T) {
	logger := &MockLogger{}
	metrics := NewMetricsCollector()
	eventBus := &MockEventBus{}
	
	// 创建依赖组件
	alertManager := NewSmartAlertManager(logger, metrics, eventBus)
	recoveryManager := NewRecoveryManager(logger, metrics, eventBus)
	retryExecutor := NewRetryExecutor(NewSmartRetryStrategy(DefaultRetryConfig()), logger, metrics)
	circuitBreaker := NewSmartCircuitBreaker(DefaultCircuitBreakerConfig(), logger, metrics, eventBus)
	
	// 创建中间件管理器
	manager := NewErrorMiddlewareManager(logger)
	manager.RegisterDefaultMiddlewares(metrics, alertManager, recoveryManager, retryExecutor, circuitBreaker)
	
	// 测试中间件数量
	middlewares := manager.GetMiddlewares()
	if len(middlewares) != 7 {
		t.Errorf("Expected 7 default middlewares, got %d", len(middlewares))
	}
	
	// 测试错误处理
	ctx := context.WithValue(context.Background(), "plugin_id", "test-plugin")
	testErr := NewPluginError(ErrorCodePluginCrashed, "Test error")
	
	finalHandlerCalled := false
	finalHandler := func(ctx context.Context, err error) error {
		finalHandlerCalled = true
		return nil
	}
	
	err := manager.HandleError(ctx, testErr, finalHandler)
	if err != nil {
		t.Errorf("Expected no error from middleware manager, got %v", err)
	}
	
	if !finalHandlerCalled {
		t.Error("Expected final handler to be called")
	}
}

// TestIntegration 集成测试
func TestIntegration(t *testing.T) {
	logger := &MockLogger{}
	metrics := NewMetricsCollector()
	eventBus := &MockEventBus{}
	pluginLoader := NewMockPluginLoader()
	
	// 创建完整的错误处理系统
	alertManager := NewSmartAlertManager(logger, metrics, eventBus)
	recoveryManager := NewRecoveryManager(logger, metrics, eventBus)
	retryExecutor := NewRetryExecutor(NewSmartRetryStrategy(DefaultRetryConfig()), logger, metrics)
	circuitBreaker := NewSmartCircuitBreaker(DefaultCircuitBreakerConfig(), logger, metrics, eventBus)
	middlewareManager := NewErrorMiddlewareManager(logger)
	
	// 注册组件
	middlewareManager.RegisterDefaultMiddlewares(metrics, alertManager, recoveryManager, retryExecutor, circuitBreaker)
	
	// 添加恢复策略
	restartStrategy := NewRestartRecoveryStrategy(3, 100*time.Millisecond, logger, metrics, eventBus, pluginLoader)
	recoveryManager.AddStrategy(restartStrategy)
	
	// 注册告警规则
	rule := &AlertRule{
		ID:          "critical-error-rule",
		Name:        "Critical Error Alert",
		Condition:   "error_rate",
		Threshold:   0.1,
		Severity:    AlertSeverityCritical,
		Enabled:     true,
		Actions: []AlertAction{
			{
				Type:    AlertActionTypeLog,
				Enabled: true,
			},
		},
	}
	alertManager.RegisterAlert(rule)
	
	// 模拟插件错误
	ctx := context.WithValue(context.Background(), "plugin_id", "test-plugin")
	criticalErr := NewPluginError(ErrorCodePluginCrashed, "Critical plugin error")
	
	// 处理错误
	finalHandler := func(ctx context.Context, err error) error {
		return nil // 模拟成功处理
	}
	
	err := middlewareManager.HandleError(ctx, criticalErr, finalHandler)
	if err != nil {
		t.Errorf("Expected successful error handling, got %v", err)
	}
	
	// 验证日志记录
	logs := logger.GetLogs()
	if len(logs) == 0 {
		t.Error("Expected logs to be recorded")
	}
	
	// 验证事件发布
	events := eventBus.GetEvents()
	if len(events) == 0 {
		t.Error("Expected events to be published")
	}
	
	// 验证恢复尝试
	time.Sleep(200 * time.Millisecond) // 等待异步恢复操作
	if len(pluginLoader.restartCalls) == 0 {
		t.Error("Expected plugin restart to be attempted")
	}
	
	// 验证指标收集
	metricsData := metrics.GetMetrics()
	if len(metricsData) == 0 {
		t.Error("Expected metrics to be collected")
	}
}

// BenchmarkErrorHandling 错误处理性能基准测试
func BenchmarkErrorHandling(b *testing.B) {
	logger := &MockLogger{}
	metrics := NewMetricsCollector()
	eventBus := &MockEventBus{}
	
	// 创建中间件管理器
	alertManager := NewSmartAlertManager(logger, metrics, eventBus)
	recoveryManager := NewRecoveryManager(logger, metrics, eventBus)
	retryExecutor := NewRetryExecutor(NewSmartRetryStrategy(DefaultRetryConfig()), logger, metrics)
	circuitBreaker := NewSmartCircuitBreaker(DefaultCircuitBreakerConfig(), logger, metrics, eventBus)
	middlewareManager := NewErrorMiddlewareManager(logger)
	middlewareManager.RegisterDefaultMiddlewares(metrics, alertManager, recoveryManager, retryExecutor, circuitBreaker)
	
	ctx := context.WithValue(context.Background(), "plugin_id", "test-plugin")
	testErr := NewPluginError(ErrorCodePluginTimeout, "Timeout error")
	finalHandler := func(ctx context.Context, err error) error {
		return nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middlewareManager.HandleError(ctx, testErr, finalHandler)
	}
}

// TestErrorClassification 测试错误分类
func TestErrorClassification(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expectedType ErrorType
		expectedSeverity ErrorSeverity
	}{
		{ErrorCodePluginCrashed, ErrorTypeSystem, ErrorSeverityCritical},
		{ErrorCodeAudioDecodeFailed, ErrorTypeAudio, ErrorSeverityError},
		{ErrorCodeMusicSourceUnavailable, ErrorTypeMusicSource, ErrorSeverityError},
		{ErrorCodeUIRenderFailed, ErrorTypeUI, ErrorSeverityError},
		{ErrorCodeInvalidArgument, ErrorTypeSystem, ErrorSeverityWarning},
	}
	
	for _, test := range tests {
		err := NewPluginError(test.code, "Test error")
		
		if err.GetType() != test.expectedType {
			t.Errorf("Error code %v: expected type %v, got %v", test.code, test.expectedType, err.GetType())
		}
		
		if err.GetSeverity() != test.expectedSeverity {
			t.Errorf("Error code %v: expected severity %v, got %v", test.code, test.expectedSeverity, err.GetSeverity())
		}
	}
}

// TestTemporaryAndTimeoutErrors 测试临时性和超时错误检测
func TestTemporaryAndTimeoutErrors(t *testing.T) {
	// 测试临时性错误
	tempErr := NewPluginError(ErrorCodeUnavailable, "Service unavailable")
	if !IsTemporary(tempErr) {
		t.Error("Expected unavailable error to be temporary")
	}
	
	permanentErr := NewPluginError(ErrorCodeInvalidArgument, "Invalid argument")
	if IsTemporary(permanentErr) {
		t.Error("Expected invalid argument error to not be temporary")
	}
	
	// 测试超时错误
	timeoutErr := NewPluginError(ErrorCodePluginTimeout, "Plugin timeout")
	if !IsTimeout(timeoutErr) {
		t.Error("Expected timeout error to be detected as timeout")
	}
	
	nonTimeoutErr := NewPluginError(ErrorCodePluginCrashed, "Plugin crashed")
	if IsTimeout(nonTimeoutErr) {
		t.Error("Expected non-timeout error to not be detected as timeout")
	}
}