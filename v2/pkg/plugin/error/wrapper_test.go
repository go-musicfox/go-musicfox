package plugin

import (
	"context"
	"errors"
	"testing"
)

// TestErrorWrapper 测试错误包装器
func TestErrorWrapper(t *testing.T) {
	wrapper := NewErrorWrapper(true, 5)
	
	t.Run("WrapError", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := wrapper.WrapError(originalErr, ErrorCodeInternal, "wrapped error")
		
		if !wrapper.IsPluginError(wrappedErr) {
			t.Error("Expected wrapped error to be a plugin error")
		}
		
		if pluginErr, ok := wrappedErr.(PluginError); ok {
			if pluginErr.GetCode() != ErrorCodeInternal {
				t.Errorf("Expected error code %s, got %s", ErrorCodeInternal.String(), pluginErr.GetCode().String())
			}
			// 检查错误消息格式："INTERNAL: wrapped error (caused by: original error)"
			expectedMsg := "INTERNAL: wrapped error (caused by: original error)"
			if pluginErr.Error() != expectedMsg {
				t.Errorf("Expected error message '%s', got '%s'", expectedMsg, pluginErr.Error())
			}
		} else {
			t.Error("Expected PluginError interface")
		}
	})
	
	t.Run("WrapWithContext", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "plugin_id", "test-plugin")
		ctx = context.WithValue(ctx, "request_id", "req-123")
		
		originalErr := errors.New("original error")
		wrappedErr := wrapper.WrapWithContext(ctx, originalErr, ErrorCodeValidation, "context wrapped error")
		
		if baseErr, ok := wrappedErr.(*BasePluginError); ok {
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
	
	t.Run("UnwrapError", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := wrapper.WrapError(originalErr, ErrorCodeInternal, "wrapped error")
		
		unwrappedErr := wrapper.UnwrapError(wrappedErr)
		if unwrappedErr != originalErr {
			t.Error("Unwrapped error does not match original")
		}
	})
	
	t.Run("IsPluginError", func(t *testing.T) {
		// 测试插件错误
		pluginErr := NewPluginError(ErrorCodeInternal, "plugin error")
		if !wrapper.IsPluginError(pluginErr) {
			t.Error("Expected plugin error to be identified as plugin error")
		}
		
		// 测试普通错误
		normalErr := errors.New("normal error")
		if wrapper.IsPluginError(normalErr) {
			t.Error("Expected normal error not to be identified as plugin error")
		}
	})
	
	t.Run("GetErrorChain", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr1 := wrapper.WrapError(originalErr, ErrorCodeInternal, "wrapped error 1")
		wrappedErr2 := wrapper.WrapError(wrappedErr1, ErrorCodeValidation, "wrapped error 2")
		
		chain := wrapper.GetErrorChain(wrappedErr2)
		if len(chain) != 3 {
			t.Errorf("Expected error chain length 3, got %d", len(chain))
		}
		
		// 验证链的顺序（从最外层到最内层）
		if chain[0] != wrappedErr2 {
			t.Error("Expected first error in chain to be wrappedErr2")
		}
		if chain[2] != originalErr {
			t.Error("Expected last error in chain to be originalErr")
		}
	})
	
	t.Run("GetRootCause", func(t *testing.T) {
		originalErr := errors.New("root cause")
		wrappedErr1 := wrapper.WrapError(originalErr, ErrorCodeInternal, "wrapped error 1")
		wrappedErr2 := wrapper.WrapError(wrappedErr1, ErrorCodeValidation, "wrapped error 2")
		
		rootCause := wrapper.GetRootCause(wrappedErr2)
		if rootCause != originalErr {
			t.Error("Root cause does not match original error")
		}
		
		// 测试单个错误的根因
		singleErr := errors.New("single error")
		singleRootCause := wrapper.GetRootCause(singleErr)
		if singleRootCause != singleErr {
			t.Error("Single error root cause should be itself")
		}
	})
	
	t.Run("AddErrorContext", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := wrapper.WrapError(originalErr, ErrorCodeInternal, "wrapped error")
		
		contextualErr := wrapper.AddErrorContext(wrappedErr, map[string]interface{}{
			"operation": "test_operation",
			"timestamp": "2023-01-01T00:00:00Z",
		})
		
		if baseErr, ok := contextualErr.(*BasePluginError); ok {
			if operation, exists := baseErr.Context["operation"]; !exists || operation != "test_operation" {
				t.Error("Expected operation in context")
			}
			if timestamp, exists := baseErr.Context["timestamp"]; !exists || timestamp != "2023-01-01T00:00:00Z" {
				t.Error("Expected timestamp in context")
			}
		} else {
			t.Error("Expected BasePluginError")
		}
	})
}

// TestErrorChain 测试错误链
func TestErrorChain(t *testing.T) {
	chain := NewErrorChain()
	
	// 添加错误
	err1 := NewPluginError(ErrorCodeInternal, "error 1")
	err2 := NewPluginError(ErrorCodeInternal, "error 2")
	err3 := NewPluginError(ErrorCodeInternal, "error 3")
	
	chain.AddError(err1)
	chain.AddError(err2)
	chain.AddError(err3)
	
	// 测试获取错误
	allErrors := chain.GetErrors()
	if len(allErrors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(allErrors))
	}
	
	// 测试获取根因
	rootCause := chain.GetRootCause()
	if rootCause != err1 {
		t.Error("Root cause should be the first error added")
	}
	
	// 测试获取最新错误
	latestError := chain.GetLatestError()
	if latestError != err3 {
		t.Error("Latest error should be the last error added")
	}
	
	// 测试错误数量
	if chain.Length() != 3 {
		t.Errorf("Expected chain length 3, got %d", chain.Length())
	}
	
	// 测试清空链
	chain.Clear()
	if chain.Length() != 0 {
		t.Errorf("Expected empty chain after clear, got length %d", chain.Length())
	}
	
	if chain.GetRootCause() != nil {
		t.Error("Expected nil root cause after clear")
	}
}

// TestErrorAggregatorWrapper 测试错误聚合器包装器功能
func TestErrorAggregatorWrapper(t *testing.T) {
	aggregator := NewErrorAggregator()
	
	// 添加错误
	err1 := NewPluginError(ErrorCodePluginInitFailed, "init failed")
	err2 := NewPluginError(ErrorCodePluginTimeout, "timeout")
	err3 := NewPluginError(ErrorCodePluginNetworkError, "network error")
	
	aggregator.AddError("plugin1", err1)
	aggregator.AddError("plugin1", err2)
	aggregator.AddError("plugin2", err3)
	
	// 测试获取插件错误
	plugin1Errors := aggregator.GetErrorsForPlugin("plugin1")
	if len(plugin1Errors) != 2 {
		t.Errorf("Expected 2 errors for plugin1, got %d", len(plugin1Errors))
	}
	
	plugin2Errors := aggregator.GetErrorsForPlugin("plugin2")
	if len(plugin2Errors) != 1 {
		t.Errorf("Expected 1 error for plugin2, got %d", len(plugin2Errors))
	}
	
	// 测试获取所有错误
	allErrors := aggregator.GetAllErrors()
	if len(allErrors) != 2 {
		t.Errorf("Expected 2 plugins with errors, got %d", len(allErrors))
	}
	
	// 测试错误统计
	stats := aggregator.GetErrorStats("plugin1")
	if stats.TotalErrors != 2 {
		t.Errorf("Expected 2 total errors for plugin1, got %d", stats.TotalErrors)
	}
	// 注意：ErrorStats结构体中没有PluginID字段，跳过验证
	
	// 测试清空插件错误
	aggregator.ClearErrorsForPlugin("plugin1")
	plugin1ErrorsAfterClear := aggregator.GetErrorsForPlugin("plugin1")
	if len(plugin1ErrorsAfterClear) != 0 {
		t.Errorf("Expected 0 errors for plugin1 after clear, got %d", len(plugin1ErrorsAfterClear))
	}
	
	// 测试清空所有错误
	aggregator.ClearAllErrors()
	allErrorsAfterClear := aggregator.GetAllErrors()
	for pluginID, errors := range allErrorsAfterClear {
		if len(errors) != 0 {
			t.Errorf("Expected 0 errors for plugin %s after clear all, got %d", pluginID, len(errors))
		}
	}
}

// TestErrorWrapperConfiguration 测试错误包装器配置
func TestErrorWrapperConfiguration(t *testing.T) {
	t.Run("WithStackTrace", func(t *testing.T) {
		wrapper := NewErrorWrapper(true, 5)
		originalErr := errors.New("test error")
		wrappedErr := wrapper.WrapError(originalErr, ErrorCodeInternal, "wrapped error")
		
		if baseErr, ok := wrappedErr.(*BasePluginError); ok {
			if baseErr.StackTrace == "" {
				t.Error("Expected stack trace to be captured")
			}
		} else {
			t.Error("Expected BasePluginError")
		}
	})
	
	t.Run("WithoutStackTrace", func(t *testing.T) {
		wrapper := NewErrorWrapper(false, 5)
		originalErr := errors.New("test error")
		wrappedErr := wrapper.WrapError(originalErr, ErrorCodeInternal, "wrapped error")
		
		if baseErr, ok := wrappedErr.(*BasePluginError); ok {
			if baseErr.StackTrace != "" {
				t.Error("Expected no stack trace to be captured")
			}
		} else {
			t.Error("Expected BasePluginError")
		}
	})
	
	t.Run("MaxDepthLimit", func(t *testing.T) {
		wrapper := NewErrorWrapper(true, 2)
		
		originalErr := errors.New("original")
		wrappedErr1 := wrapper.WrapError(originalErr, ErrorCodeInternal, "wrap 1")
		wrappedErr2 := wrapper.WrapError(wrappedErr1, ErrorCodeValidation, "wrap 2")
		wrappedErr3 := wrapper.WrapError(wrappedErr2, ErrorCodePluginTimeout, "wrap 3")
		
		chain := wrapper.GetErrorChain(wrappedErr3)
		// 由于最大深度限制，链的长度可能会受到限制
		// 这里主要测试不会导致无限递归
		if len(chain) == 0 {
			t.Error("Expected non-empty error chain")
		}
	})
}

// TestErrorWrapperConcurrency 测试错误包装器并发安全性
func TestErrorWrapperConcurrency(t *testing.T) {
	wrapper := NewErrorWrapper(true, 10)
	
	const numGoroutines = 10
	const operationsPerGoroutine = 100
	
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()
			
			for j := 0; j < operationsPerGoroutine; j++ {
				originalErr := errors.New("concurrent error")
				wrappedErr := wrapper.WrapError(originalErr, ErrorCodeInternal, "wrapped error")
				
				// 验证包装是否成功
				if !wrapper.IsPluginError(wrappedErr) {
					t.Errorf("Goroutine %d: Expected plugin error", goroutineID)
				}
				
				// 验证解包是否成功
				unwrappedErr := wrapper.UnwrapError(wrappedErr)
				if unwrappedErr != originalErr {
					t.Errorf("Goroutine %d: Unwrap failed", goroutineID)
				}
			}
		}(i)
	}
	
	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}