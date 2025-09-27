package plugin

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// TestDynamicErrorHandler 测试动态错误处理器
func TestDynamicErrorHandler(t *testing.T) {
	config := &ErrorHandlerConfig{
		MaxErrorHistory:     10,
		AutoRecovery:        true,
		MaxRecoveryAttempts: 3,
		ErrorRateWindow:     100 * time.Millisecond,
	}
	
	handler := NewDynamicErrorHandler(config)
	if handler == nil {
		t.Fatal("Expected non-nil error handler")
	}
	
	dynamicHandler, ok := handler.(*DynamicErrorHandler)
	if !ok {
		t.Fatal("Expected DynamicErrorHandler type")
	}
	
	// 测试基本错误处理
	ctx := context.Background()
	testErr := errors.New("test error")
	pluginID := "test-plugin"
	
	err := handler.HandleError(ctx, pluginID, testErr)
	if err != nil {
		t.Errorf("HandleError failed: %v", err)
	}
	
	// 验证错误历史
	history, err := handler.GetErrorHistory(pluginID)
	if err != nil {
		t.Errorf("GetErrorHistory failed: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("Expected 1 error in history, got %d", len(history))
	}
	
	// 测试错误统计
	stats, err := handler.GetErrorStats(pluginID)
	if err != nil {
		t.Errorf("GetErrorStats failed: %v", err)
	}
	if stats == nil {
		t.Error("Expected error stats, got nil")
	} else {
		if stats.TotalErrors != 1 {
			t.Errorf("Expected 1 total error, got %d", stats.TotalErrors)
		}
	}
	
	// 测试恢复策略设置
	strategy := ErrorRecoveryStrategy{
		Type:       RecoveryTypeRetry,
		MaxRetries: 5,
		RetryDelay: 200 * time.Millisecond,
	}
	
	dynamicHandler.SetRecoveryStrategy(ErrorCodePluginTimeout, strategy)
	retrievedStrategy, exists := dynamicHandler.GetRecoveryStrategy(ErrorCodePluginTimeout)
	
	if !exists {
		t.Error("Expected recovery strategy to exist")
	} else {
		if retrievedStrategy.Type != RecoveryTypeRetry {
			t.Errorf("Expected retry recovery type, got %s", retrievedStrategy.Type.String())
		}
		if retrievedStrategy.MaxRetries != 5 {
			t.Errorf("Expected 5 max retries, got %d", retrievedStrategy.MaxRetries)
		}
	}
	
	// 测试清除错误历史
	handler.ClearErrorHistory(pluginID)
	historyAfterClear, err := handler.GetErrorHistory(pluginID)
	if err != nil {
		t.Errorf("GetErrorHistory after clear failed: %v", err)
	}
	if len(historyAfterClear) != 0 {
		t.Errorf("Expected 0 errors in history after clear, got %d", len(historyAfterClear))
	}
}

// TestErrorRecoveryStrategies 测试错误恢复策略
func TestErrorRecoveryStrategies(t *testing.T) {
	config := &ErrorHandlerConfig{
		MaxErrorHistory:     10,
		AutoRecovery:        true,
		MaxRecoveryAttempts: 3,
		ErrorRateWindow:     50 * time.Millisecond,
	}
	
	handler := NewDynamicErrorHandler(config).(*DynamicErrorHandler)
	ctx := context.Background()
	pluginID := "test-plugin"
	
	t.Run("RetryStrategy", func(t *testing.T) {
		// 设置重试策略
		strategy := ErrorRecoveryStrategy{
			Type:          RecoveryTypeRetry,
			MaxRetries:    3,
			RetryDelay:    50 * time.Millisecond,
			BackoffFactor: 1.5,
			MaxDelay:      500 * time.Millisecond,
			Jitter:        true,
		}
		
		handler.SetRecoveryStrategy(ErrorCodePluginTimeout, strategy)
		
		// 创建可重试的错误
		retryableErr := NewPluginError(ErrorCodePluginTimeout, "timeout error")
		retryableErr.WithRetryConfig(true, 0)
		
		// 执行重试策略
		err := handler.executeRecoveryStrategy(ctx, strategy, retryableErr, pluginID)
		if err != nil {
			t.Logf("Retry strategy completed with error: %v", err)
		}
	})
	
	t.Run("RestartStrategy", func(t *testing.T) {
		// 设置重启策略
		strategy := ErrorRecoveryStrategy{
			Type:       RecoveryTypeRestart,
			MaxRetries: 1,
			RetryDelay: 100 * time.Millisecond,
		}
		
		handler.SetRecoveryStrategy(ErrorCodePluginCrashed, strategy)
		
		// 创建需要重启的错误
		crashedErr := NewPluginError(ErrorCodePluginCrashed, "plugin crashed")
		
		// 执行重启策略
		err := handler.executeRecoveryStrategy(ctx, strategy, crashedErr, pluginID)
		if err != nil {
			t.Logf("Restart strategy completed with error: %v", err)
		}
	})
	
	t.Run("FallbackStrategy", func(t *testing.T) {
		// 设置回退策略
		strategy := ErrorRecoveryStrategy{
			Type: RecoveryTypeFallback,
		}
		
		handler.SetRecoveryStrategy(ErrorCodePluginInitFailed, strategy)
		
		// 创建需要回退的错误
		initErr := NewPluginError(ErrorCodePluginInitFailed, "init failed")
		
		// 执行回退策略
		err := handler.executeRecoveryStrategy(ctx, strategy, initErr, pluginID)
		if err != nil {
			t.Errorf("Fallback strategy failed: %v", err)
		}
	})
	
	t.Run("GracefulDegradationStrategy", func(t *testing.T) {
		// 设置优雅降级策略
		strategy := ErrorRecoveryStrategy{
			Type: RecoveryTypeGracefulDegradation,
			DegradationConfig: &DegradationConfig{
				DisableFeatures: []string{"advanced_feature"},
				ReduceQuality:   true,
				FallbackMode:    true,
			},
		}
		
		handler.SetRecoveryStrategy(ErrorCodePluginResourceLimit, strategy)
		
		// 创建资源限制错误
		resourceErr := NewPluginError(ErrorCodePluginResourceLimit, "resource limit exceeded")
		
		// 执行优雅降级策略
		err := handler.executeRecoveryStrategy(ctx, strategy, resourceErr, pluginID)
		if err != nil {
			t.Errorf("Graceful degradation strategy failed: %v", err)
		}
	})
}

// TestErrorHandlerConcurrency 测试错误处理器并发安全性
func TestErrorHandlerConcurrency(t *testing.T) {
	config := &ErrorHandlerConfig{
		MaxErrorHistory:     100,
		AutoRecovery:        true,
		MaxRecoveryAttempts: 3,
		ErrorRateWindow:     10 * time.Millisecond,
	}
	
	handler := NewDynamicErrorHandler(config)
	ctx := context.Background()
	
	// 并发处理错误
	const numGoroutines = 10
	const errorsPerGoroutine = 10
	
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()
			
			for j := 0; j < errorsPerGoroutine; j++ {
				pluginID := fmt.Sprintf("plugin-%d", goroutineID)
				testErr := NewPluginError(ErrorCodePluginTimeout, fmt.Sprintf("error %d from goroutine %d", j, goroutineID))
				
				err := handler.HandleError(ctx, pluginID, testErr)
				if err != nil {
					t.Errorf("HandleError failed in goroutine %d: %v", goroutineID, err)
				}
			}
		}(i)
	}
	
	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// 验证结果
	for i := 0; i < numGoroutines; i++ {
		pluginID := fmt.Sprintf("plugin-%d", i)
		stats, err := handler.GetErrorStats(pluginID)
		if err != nil {
			t.Errorf("GetErrorStats failed for plugin %s: %v", pluginID, err)
			continue
		}
		if stats == nil {
			t.Errorf("Expected error stats for plugin %s, got nil", pluginID)
			continue
		}
		
		if stats.TotalErrors != errorsPerGoroutine {
			t.Errorf("Expected %d errors for plugin %s, got %d", errorsPerGoroutine, pluginID, stats.TotalErrors)
		}
		
		history, err := handler.GetErrorHistory(pluginID)
		if err != nil {
			t.Errorf("GetErrorHistory failed for plugin %s: %v", pluginID, err)
			continue
		}
		if len(history) != errorsPerGoroutine {
			t.Errorf("Expected %d errors in history for plugin %s, got %d", errorsPerGoroutine, pluginID, len(history))
		}
	}
}

// TestErrorHandlerConfiguration 测试错误处理器配置
func TestErrorHandlerConfiguration(t *testing.T) {
	t.Run("DefaultConfiguration", func(t *testing.T) {
		handler := NewDynamicErrorHandler(nil)
		if handler == nil {
			t.Fatal("Expected non-nil error handler with default config")
		}
		
		dynamicHandler := handler.(*DynamicErrorHandler)
		if dynamicHandler.config.MaxErrorHistory != 1000 {
			t.Errorf("Expected default MaxErrorHistory 1000, got %d", dynamicHandler.config.MaxErrorHistory)
		}
		
		if !dynamicHandler.config.AutoRecovery {
			t.Error("Expected default AutoRecovery to be true")
		}
	})
	
	t.Run("CustomConfiguration", func(t *testing.T) {
		config := &ErrorHandlerConfig{
			MaxErrorHistory:     50,
			AutoRecovery:        false,
			MaxRecoveryAttempts: 5,
			ErrorRateWindow:     200 * time.Millisecond,
		}
		
		handler := NewDynamicErrorHandler(config)
		dynamicHandler := handler.(*DynamicErrorHandler)
		
		if dynamicHandler.config.MaxErrorHistory != 50 {
			t.Errorf("Expected MaxErrorHistory 50, got %d", dynamicHandler.config.MaxErrorHistory)
		}
		
		if dynamicHandler.config.AutoRecovery {
			t.Error("Expected AutoRecovery to be false")
		}
		
		if dynamicHandler.config.MaxRecoveryAttempts != 5 {
			t.Errorf("Expected MaxRecoveryAttempts 5, got %d", dynamicHandler.config.MaxRecoveryAttempts)
		}
		
		if dynamicHandler.config.ErrorRateWindow != 200*time.Millisecond {
			t.Errorf("Expected ErrorRateWindow 200ms, got %v", dynamicHandler.config.ErrorRateWindow)
		}
	})
}

// TestErrorHandlerMetrics 测试错误处理器指标
func TestErrorHandlerMetrics(t *testing.T) {
	config := &ErrorHandlerConfig{
		MaxErrorHistory:     10,
		AutoRecovery:        true,
		MaxRecoveryAttempts: 3,
		ErrorRateWindow:     50 * time.Millisecond,
	}
	
	handler := NewDynamicErrorHandler(config).(*DynamicErrorHandler)
	ctx := context.Background()
	
	// 处理不同类型的错误
	errorTypes := []struct {
		code     ErrorCode
		message  string
		pluginID string
	}{
		{ErrorCodePluginTimeout, "timeout error", "plugin1"},
		{ErrorCodePluginInitFailed, "init failed", "plugin1"},
		{ErrorCodePluginNetworkError, "network error", "plugin2"},
		{ErrorCodePluginCrashed, "plugin crashed", "plugin2"},
		{ErrorCodePluginMemoryLimit, "memory limit", "plugin3"},
	}
	
	for _, et := range errorTypes {
		testErr := NewPluginError(et.code, et.message)
		err := handler.HandleError(ctx, et.pluginID, testErr)
		if err != nil {
			t.Errorf("HandleError failed for %s: %v", et.code.String(), err)
		}
	}
	
	// 获取指标
	metrics := handler.GetMetrics()
	if metrics == nil {
		t.Fatal("Expected metrics, got nil")
	}
	
	// 验证指标数据
	allMetrics := metrics.GetMetrics()
	if len(allMetrics) == 0 {
		t.Error("Expected metrics data, got empty map")
	}
	
	// 验证各插件的错误统计
		for _, pluginID := range []string{"plugin1", "plugin2", "plugin3"} {
			stats, err := handler.GetErrorStats(pluginID)
			if err != nil {
				t.Errorf("GetErrorStats failed for %s: %v", pluginID, err)
				continue
			}
			if stats == nil {
				t.Errorf("Expected error stats for %s, got nil", pluginID)
				continue
			}
		
		if stats.TotalErrors == 0 {
			t.Errorf("Expected errors for %s, got 0", pluginID)
		}
		
		t.Logf("Plugin %s: %d total errors, last error: %v", 
			pluginID, stats.TotalErrors, stats.LastError)
	}
}

// TestErrorHandlerEventBus 测试错误处理器事件总线
func TestErrorHandlerEventBus(t *testing.T) {
	config := &ErrorHandlerConfig{
		MaxErrorHistory:     10,
		AutoRecovery:        true,
		MaxRecoveryAttempts: 3,
		ErrorRateWindow:     50 * time.Millisecond,
	}
	
	handler := NewDynamicErrorHandler(config).(*DynamicErrorHandler)
	
	// 创建模拟事件总线
	eventBus := &MockEventBus{
		events: make([]MockEvent, 0),
	}
	
	handler.SetEventBus(eventBus)
	
	// 处理错误
	ctx := context.Background()
	testErr := NewPluginError(ErrorCodePluginTimeout, "timeout error")
	err := handler.HandleError(ctx, "test-plugin", testErr)
	if err != nil {
		t.Errorf("HandleError failed: %v", err)
	}
	
	// 验证事件是否被发布
	if len(eventBus.events) == 0 {
		t.Error("Expected events to be published, got none")
	} else {
		t.Logf("Published %d events", len(eventBus.events))
		for i, event := range eventBus.events {
			t.Logf("Event %d: %+v", i, event)
		}
	}
}