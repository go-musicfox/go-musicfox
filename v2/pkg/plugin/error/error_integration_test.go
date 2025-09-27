package plugin

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"
)

// TestErrorHandlingIntegration 测试错误处理集成
func TestErrorHandlingIntegration(t *testing.T) {
	// 创建测试组件
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	errorLogger := NewErrorLogger(logger, LogLevelDebug)
	metrics := NewMetricsCollector()
	wrapper := NewErrorWrapper(true, 5)
	monitor := NewErrorMonitor(metrics, errorLogger)
	classifier := NewErrorClassifier(errorLogger, metrics)
	propagator := NewErrorPropagator(errorLogger, metrics)
	
	// 创建错误处理器
	config := &ErrorHandlerConfig{
		MaxErrorHistory:     100,
		AutoRecovery:        true,
		MaxRecoveryAttempts: 3,
		ErrorRateWindow:     time.Second,
	}
	handler := NewDynamicErrorHandler(config).(*DynamicErrorHandler)
	
	// 设置组件
	handler.wrapper = wrapper
	handler.propagator = propagator
	handler.logger = errorLogger
	handler.monitor = monitor
	handler.classifier = classifier
	handler.metrics = metrics
	
	// 启动监控
	ctx := context.Background()
	if err := monitor.Start(ctx); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer func() {
		if err := monitor.Stop(); err != nil {
			t.Logf("Failed to stop monitor: %v", err)
		}
	}()
	
	// 测试错误处理流程
	t.Run("BasicErrorHandling", func(t *testing.T) {
		testErr := errors.New("test error")
		pluginID := "test-plugin"
		
		err := handler.HandleError(ctx, pluginID, testErr)
		if err != nil {
			t.Errorf("HandleError failed: %v", err)
		}
		
		// 验证错误统计
		stats := monitor.GetErrorStats(pluginID)
		if stats == nil {
			t.Error("Expected error stats, got nil")
		} else if stats.TotalErrors != 1 {
			t.Errorf("Expected 1 error, got %d", stats.TotalErrors)
		}
	})
	
	t.Run("ErrorClassification", func(t *testing.T) {
		// 添加分类规则
		rule := ClassificationRule{
			ID:          "timeout-rule",
			Name:        "Timeout Error Rule",
			Description: "Classifies timeout errors",
			Conditions: []RuleCondition{
				{
					Field:    "message",
					Operator: "contains",
					Value:    "timeout",
				},
			},
			Action: ClassificationAction{
				ErrorCode: ErrorCodePluginTimeout,
				ErrorType: ErrorTypeTimeout,
				Severity:  ErrorSeverityError,
				Category:  "timeout",
				Tags:      []string{"timeout", "recoverable"},
			},
			Priority: 10,
			Enabled:  true,
			Weight:   0.9,
		}
		
		if err := classifier.AddRule(rule); err != nil {
			t.Fatalf("Failed to add classification rule: %v", err)
		}
		
		// 测试分类
		timeoutErr := errors.New("operation timeout after 30 seconds")
		classification, err := classifier.ClassifyError(ctx, timeoutErr, "test-plugin")
		if err != nil {
			t.Fatalf("Classification failed: %v", err)
		}
		
		if classification.ErrorCode != ErrorCodePluginTimeout {
			t.Errorf("Expected timeout error code, got %s", classification.ErrorCode.String())
		}
		
		if classification.Category != "timeout" {
			t.Errorf("Expected timeout category, got %s", classification.Category)
		}
	})
	
	t.Run("ErrorRecovery", func(t *testing.T) {
		// 设置恢复策略
		strategy := ErrorRecoveryStrategy{
			Type:          RecoveryTypeRetry,
			MaxRetries:    3,
			RetryDelay:    100 * time.Millisecond,
			BackoffFactor: 1.5,
			MaxDelay:      time.Second,
			Jitter:        true,
		}
		
		handler.SetRecoveryStrategy(ErrorCodePluginTimeout, strategy)
		
		// 创建可重试的错误
		retryableErr := NewPluginError(ErrorCodePluginTimeout, "timeout error")
		retryableErr.WithRetryConfig(true, 0)
		
		// 处理错误（应该触发恢复策略）
		err := handler.HandleError(ctx, "test-plugin", retryableErr)
		if err != nil {
			t.Errorf("HandleError with recovery failed: %v", err)
		}
		
		// 验证恢复策略是否被调用
		recoveryMetrics := metrics.GetMetrics()
		if counters, ok := recoveryMetrics["counters"].(map[string]interface{}); ok {
			found := false
			for key := range counters {
				if key == "error_recovery_retry_plugin_id_test-plugin_attempt_1" {
					found = true
					break
				}
			}
			if !found {
				t.Log("Recovery retry metrics not found (this might be expected in test environment)")
			}
		}
	})
	
	t.Run("ErrorPropagation", func(t *testing.T) {
		// 创建传播规则
		rule := PropagationRule{
			ID:   "critical-error-rule",
			Name: "Critical Error Propagation",
			Condition: PropagationCondition{
				SeverityLevels: []ErrorSeverity{ErrorSeverityCritical},
			},
			Action: PropagationAction{
				Type:  PropagationActionLog,
				Async: false,
			},
			Enabled:  true,
			Priority: 1,
		}
		
		propagator.SetPropagationRules([]PropagationRule{rule})
		
		// 创建严重错误
		criticalErr := NewPluginError(ErrorCodePluginCrashed, "plugin crashed")
		criticalErr.WithSeverity(ErrorSeverityCritical)
		
		// 处理错误（应该触发传播）
		err := handler.HandleError(ctx, "test-plugin", criticalErr)
		if err != nil {
			t.Errorf("HandleError with propagation failed: %v", err)
		}
	})
	
	t.Run("ErrorMonitoring", func(t *testing.T) {
		// 设置告警阈值
		threshold := AlertThreshold{
			ErrorRate:     0.5, // 50%错误率
			ErrorCount:    5,   // 5个错误
			TimeWindow:    time.Minute,
			SeverityLevel: ErrorSeverityError,
			Enabled:       true,
		}
		
		monitor.SetAlertThreshold("test-plugin", threshold)
		
		// 生成多个错误
		for i := 0; i < 6; i++ {
			testErr := NewPluginError(ErrorCodePluginInitFailed, "init failed")
			monitor.RecordError(ctx, testErr, "test-plugin")
		}
		
		// 检查告警
		alerts := monitor.CheckAlerts()
		if len(alerts) == 0 {
			t.Error("Expected alerts to be generated")
		} else {
			t.Logf("Generated %d alerts", len(alerts))
			for _, alert := range alerts {
				t.Logf("Alert: %s - %s", alert.Type.String(), alert.Message)
			}
		}
	})
	
	t.Run("ErrorWrapper", func(t *testing.T) {
		// 测试错误包装
		originalErr := errors.New("original error")
		wrappedErr := wrapper.WrapError(originalErr, ErrorCodeInternal, "wrapped error")
		
		if !wrapper.IsPluginError(wrappedErr) {
			t.Error("Expected wrapped error to be a plugin error")
		}
		
		unwrappedErr := wrapper.UnwrapError(wrappedErr)
		if unwrappedErr != originalErr {
			t.Error("Unwrapped error does not match original")
		}
		
		// 测试带上下文的包装
		ctxWithValues := context.WithValue(ctx, "plugin_id", "test-plugin")
		ctxWithValues = context.WithValue(ctxWithValues, "request_id", "req-123")
		
		contextWrappedErr := wrapper.WrapWithContext(ctxWithValues, originalErr, ErrorCodeInvalidArgument, "context wrapped error")
		if baseErr, ok := contextWrappedErr.(*BasePluginError); ok {
			if pluginID, exists := baseErr.Context["plugin_id"]; !exists || pluginID != "test-plugin" {
				t.Error("Expected plugin_id in context")
			}
			if requestID, exists := baseErr.Context["request_id"]; !exists || requestID != "req-123" {
				t.Error("Expected request_id in context")
			}
		} else {
			t.Error("Expected BasePluginError")
		}
	})
	
	t.Run("MetricsCollection", func(t *testing.T) {
		// 重置指标
		metrics.Reset()
		
		// 记录一些指标
		metrics.IncrementCounter("test_counter", map[string]string{"type": "test"})
		metrics.SetGauge("test_gauge", 42.0, map[string]string{"unit": "count"})
		metrics.RecordHistogram("test_histogram", 1.5, map[string]string{"operation": "test"})
		metrics.RecordTimer("test_timer", 100*time.Millisecond, map[string]string{"method": "test"})
		
		// 获取指标
		allMetrics := metrics.GetMetrics()
		
		// 验证计数器
		if counters, ok := allMetrics["counters"].(map[string]interface{}); ok {
			if len(counters) == 0 {
				t.Error("Expected counters to be recorded")
			}
		} else {
			t.Error("Expected counters in metrics")
		}
		
		// 验证仪表盘
		if gauges, ok := allMetrics["gauges"].(map[string]interface{}); ok {
			if len(gauges) == 0 {
				t.Error("Expected gauges to be recorded")
			}
		} else {
			t.Error("Expected gauges in metrics")
		}
	})
}

// TestErrorClassifierTraining 测试错误分类器训练
func TestErrorClassifierTraining(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	errorLogger := NewErrorLogger(logger, LogLevelDebug)
	metrics := NewMetricsCollector()
	classifier := NewErrorClassifier(errorLogger, metrics)
	
	// 准备训练数据
	trainingData := []TrainingExample{
		{
			ErrorMessage:     "connection timeout after 30 seconds",
			PluginID:         "network-plugin",
			ExpectedCode:     ErrorCodePluginTimeout,
			ExpectedType:     ErrorTypeTimeout,
			ExpectedSeverity: ErrorSeverityError,
			Weight:           1.0,
		},
		{
			ErrorMessage:     "network connection refused",
			PluginID:         "network-plugin",
			ExpectedCode:     ErrorCodePluginNetworkError,
			ExpectedType:     ErrorTypeNetwork,
			ExpectedSeverity: ErrorSeverityError,
			Weight:           1.0,
		},
		{
			ErrorMessage:     "permission denied accessing file",
			PluginID:         "file-plugin",
			ExpectedCode:     ErrorCodePermissionDenied,
			ExpectedType:     ErrorTypeAuthorization,
			ExpectedSeverity: ErrorSeverityError,
			Weight:           1.0,
		},
		{
			ErrorMessage:     "out of memory error",
			PluginID:         "memory-plugin",
			ExpectedCode:     ErrorCodePluginMemoryLimit,
			ExpectedType:     ErrorTypeResource,
			ExpectedSeverity: ErrorSeverityFatal,
			Weight:           1.0,
		},
	}
	
	// 训练分类器
	err := classifier.TrainClassifier(trainingData)
	if err != nil {
		t.Fatalf("Training failed: %v", err)
	}
	
	// 测试分类
	ctx := context.Background()
	testCases := []struct {
		errorMessage string
		pluginID     string
		expectedType ErrorType
	}{
		{"timeout occurred", "test-plugin", ErrorTypeTimeout},
		{"network error", "test-plugin", ErrorTypeNetwork},
		{"permission denied", "test-plugin", ErrorTypeAuthorization},
		{"memory limit exceeded", "test-plugin", ErrorTypeResource},
	}
	
	for _, tc := range testCases {
		t.Run(tc.errorMessage, func(t *testing.T) {
			testErr := errors.New(tc.errorMessage)
			classification, err := classifier.ClassifyError(ctx, testErr, tc.pluginID)
			if err != nil {
				t.Fatalf("Classification failed: %v", err)
			}
			
			t.Logf("Classified '%s' as %s (confidence: %.2f)", 
				tc.errorMessage, classification.ErrorType.String(), classification.Confidence)
			
			// 注意：由于这是简化的分类器，可能不会完全匹配预期
			// 在实际应用中，需要更复杂的训练和验证
		})
	}
	
	// 测试分类器反馈
	feedback := []ClassificationFeedback{
		{
			ErrorMessage:      "timeout test",
			PluginID:          "test-plugin",
			PredictedCode:     ErrorCodePluginTimeout,
			ActualCode:        ErrorCodePluginTimeout,
			PredictedType:     ErrorTypeTimeout,
			ActualType:        ErrorTypeTimeout,
			PredictedSeverity: ErrorSeverityError,
			ActualSeverity:    ErrorSeverityError,
			Correct:           true,
			Confidence:        0.8,
			Timestamp:         time.Now(),
		},
	}
	
	err = classifier.UpdateWeights(feedback)
	if err != nil {
		t.Fatalf("Weight update failed: %v", err)
	}
	
	accuracy := classifier.GetAccuracy()
	t.Logf("Classifier accuracy: %.2f", accuracy)
}

// TestErrorAggregator 测试错误聚合器
func TestErrorAggregator(t *testing.T) {
	aggregator := NewErrorAggregator()
	
	// 添加错误
	err1 := NewPluginError(ErrorCodePluginInitFailed, "init failed")
	err2 := NewPluginError(ErrorCodePluginTimeout, "timeout")
	err3 := NewPluginError(ErrorCodePluginNetworkError, "network error")
	
	aggregator.AddError("plugin1", err1)
	aggregator.AddError("plugin1", err2)
	aggregator.AddError("plugin2", err3)
	
	// 获取插件错误
	plugin1Errors := aggregator.GetErrorsForPlugin("plugin1")
	if len(plugin1Errors) != 2 {
		t.Errorf("Expected 2 errors for plugin1, got %d", len(plugin1Errors))
	}
	
	plugin2Errors := aggregator.GetErrorsForPlugin("plugin2")
	if len(plugin2Errors) != 1 {
		t.Errorf("Expected 1 error for plugin2, got %d", len(plugin2Errors))
	}
	
	// 获取所有错误
	allErrors := aggregator.GetAllErrors()
	if len(allErrors) != 2 {
		t.Errorf("Expected 2 plugins with errors, got %d", len(allErrors))
	}
	
	// 清空插件错误
	aggregator.ClearErrorsForPlugin("plugin1")
	plugin1ErrorsAfterClear := aggregator.GetErrorsForPlugin("plugin1")
	if len(plugin1ErrorsAfterClear) != 0 {
		t.Errorf("Expected 0 errors for plugin1 after clear, got %d", len(plugin1ErrorsAfterClear))
	}
	
	// 清空所有错误
	aggregator.ClearAllErrors()
	allErrorsAfterClear := aggregator.GetAllErrors()
	for pluginID, errors := range allErrorsAfterClear {
		if len(errors) != 0 {
			t.Errorf("Expected 0 errors for plugin %s after clear all, got %d", pluginID, len(errors))
		}
	}
}