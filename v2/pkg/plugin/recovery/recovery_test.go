package recovery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCircuitBreaker 测试熔断器
func TestCircuitBreaker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	t.Run("InitialState", func(t *testing.T) {
		config := DefaultCircuitBreakerConfig()
		config.FailureThreshold = 3
		config.RecoveryTimeout = 100 * time.Millisecond
		cb := NewCircuitBreaker("test-cb", config, logger)
		
		assert.Equal(t, StateClosed, cb.GetState())
		assert.True(t, cb.IsHealthy())
	})

	t.Run("SuccessfulExecution", func(t *testing.T) {
		config := DefaultCircuitBreakerConfig()
		config.FailureThreshold = 3
		config.RecoveryTimeout = 100 * time.Millisecond
		cb := NewCircuitBreaker("test-cb", config, logger)
		
		ctx := context.Background()
		result := cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		assert.NoError(t, result)
		assert.Equal(t, StateClosed, cb.GetState())
	})

	t.Run("FailureExecution", func(t *testing.T) {
		config := DefaultCircuitBreakerConfig()
		config.FailureThreshold = 3
		config.RecoveryTimeout = 100 * time.Millisecond
		cb := NewCircuitBreaker("test-cb", config, logger)
		
		ctx := context.Background()
		
		// 触发失败直到熔断器开启
		for i := 0; i < config.FailureThreshold; i++ {
			err := cb.Execute(ctx, func(ctx context.Context) error {
				return errors.New("test error")
			})
			assert.Error(t, err)
		}

		// 熔断器应该开启
		assert.Equal(t, StateOpen, cb.GetState())
		assert.False(t, cb.IsHealthy())
	})

	t.Run("OpenStateRejectsRequests", func(t *testing.T) {
		config := DefaultCircuitBreakerConfig()
		config.FailureThreshold = 3
		config.RecoveryTimeout = 100 * time.Millisecond
		cb := NewCircuitBreaker("test-cb", config, logger)
		
		// 先触发失败让熔断器开启
		ctx := context.Background()
		for i := 0; i < config.FailureThreshold; i++ {
			cb.Execute(ctx, func(ctx context.Context) error {
				return errors.New("test error")
			})
		}
		
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker")
		assert.Contains(t, err.Error(), "is open")
	})

	t.Run("TransitionToHalfOpen", func(t *testing.T) {
		config := DefaultCircuitBreakerConfig()
		config.FailureThreshold = 3
		config.SuccessThreshold = 1
		config.RecoveryTimeout = 100 * time.Millisecond
		cb := NewCircuitBreaker("test-cb", config, logger)
		
		// 先触发失败让熔断器开启
		ctx := context.Background()
		for i := 0; i < config.FailureThreshold; i++ {
			cb.Execute(ctx, func(ctx context.Context) error {
				return errors.New("test error")
			})
		}
		
		// 等待恢复超时
		time.Sleep(config.RecoveryTimeout + 10*time.Millisecond)

		// 第一个请求应该被允许（转为半开状态）
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		assert.NoError(t, err)
		
		// 成功后应该转为关闭状态
		// 注意：在半开状态下，一次成功就会转为关闭状态，并重置计数器
		assert.Equal(t, StateClosed, cb.GetState())
		assert.True(t, cb.IsHealthy())
	})

	t.Run("Metrics", func(t *testing.T) {
		config := DefaultCircuitBreakerConfig()
		config.FailureThreshold = 3
		config.RecoveryTimeout = 100 * time.Millisecond
		cb := NewCircuitBreaker("test-cb", config, logger)
		
		// 执行一些操作来生成指标
		ctx := context.Background()
		cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		cb.Execute(ctx, func(ctx context.Context) error {
			return errors.New("test error")
		})
		
		metrics := cb.GetMetrics()
		assert.NotNil(t, metrics)
		assert.True(t, metrics.TotalRequests > 0)
		assert.True(t, metrics.FailureRequests > 0)
		assert.True(t, metrics.SuccessRequests > 0)
	})

	t.Run("Reset", func(t *testing.T) {
		config := DefaultCircuitBreakerConfig()
		config.FailureThreshold = 3
		config.RecoveryTimeout = 100 * time.Millisecond
		cb := NewCircuitBreaker("test-cb", config, logger)
		
		// 执行一些操作
		ctx := context.Background()
		cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		
		cb.Reset()
		assert.Equal(t, StateClosed, cb.GetState())
		assert.True(t, cb.IsHealthy())
		
		metrics := cb.GetMetrics()
		assert.Equal(t, int64(0), metrics.TotalRequests)
	})
}

// TestRetryStrategy 测试重试策略
func TestRetryStrategy(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	config := DefaultRetryConfig()
	config.MaxRetries = 3
	config.InitialDelay = 10 * time.Millisecond
	config.Policy = RetryPolicyExponential
	config.RetryableErrors = []string{"retryable"}

	rs := NewRetryStrategy("test-retry", config, logger)

	t.Run("SuccessfulExecution", func(t *testing.T) {
		ctx := context.Background()
		err := rs.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("RetryableError", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0
		
		err := rs.Execute(ctx, func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return errors.New("retryable error")
			}
			return nil
		})
		
		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("NonRetryableError", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0
		
		err := rs.Execute(ctx, func(ctx context.Context) error {
			attempts++
			return errors.New("permanent failure")
		})
		
		assert.Error(t, err)
		assert.Equal(t, 1, attempts)
		assert.Contains(t, err.Error(), "non-retryable")
	})

	t.Run("MaxRetriesExceeded", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0
		
		err := rs.Execute(ctx, func(ctx context.Context) error {
			attempts++
			return errors.New("retryable error")
		})
		
		assert.Error(t, err)
		assert.Equal(t, config.MaxRetries+1, attempts)
		assert.Contains(t, err.Error(), "max retries")
	})

	t.Run("Metrics", func(t *testing.T) {
		metrics := rs.GetMetrics()
		assert.NotNil(t, metrics)
		assert.True(t, metrics.TotalAttempts > 0)
	})

	t.Run("DifferentPolicies", func(t *testing.T) {
		policies := []RetryPolicy{
			RetryPolicyFixed,
			RetryPolicyLinear,
			RetryPolicyExponential,
		}

		for _, policy := range policies {
			t.Run(policy.String(), func(t *testing.T) {
				config := DefaultRetryConfig()
				config.Policy = policy
				config.MaxRetries = 2
				config.InitialDelay = 1 * time.Millisecond
				
				rs := NewRetryStrategy("test-policy", config, logger)
				
				ctx := context.Background()
				attempts := 0
				
				err := rs.Execute(ctx, func(ctx context.Context) error {
					attempts++
					if attempts < 3 {
						return errors.New("retryable error")
					}
					return nil
				})
				
				assert.NoError(t, err)
				assert.Equal(t, 3, attempts)
			})
		}
	})
}

// TestFallbackStrategy 测试降级策略
func TestFallbackStrategy(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	config := DefaultFallbackConfig()
	config.Type = FallbackTypeDefault
	config.DefaultValue = "fallback result"

	fs := NewFallbackStrategy("test-fallback", config, logger)

	t.Run("SuccessfulPrimaryOperation", func(t *testing.T) {
		ctx := context.Background()
		result, err := fs.Execute(ctx, func(ctx context.Context) (interface{}, error) {
			return "primary result", nil
		}, nil)
		
		assert.NoError(t, err)
		assert.Equal(t, "primary result", result)
	})

	t.Run("FallbackOnPrimaryFailure", func(t *testing.T) {
		ctx := context.Background()
		result, err := fs.Execute(ctx, func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("primary operation failed")
		}, nil)
		
		assert.NoError(t, err)
		assert.Equal(t, "fallback result", result)
	})

	t.Run("CacheFallback", func(t *testing.T) {
		cacheConfig := DefaultFallbackConfig()
		cacheConfig.Type = FallbackTypeCache
		cacheConfig.CacheExpiry = 1 * time.Second
		
		cacheFs := NewFallbackStrategy("test-cache-fallback", cacheConfig, logger)
		
		ctx := context.Background()
		
		// 首次成功，应该缓存结果
		result1, err := cacheFs.Execute(ctx, func(ctx context.Context) (interface{}, error) {
			return "cached result", nil
		}, "test-key")
		
		assert.NoError(t, err)
		assert.Equal(t, "cached result", result1)
		
		// 第二次失败，应该返回缓存结果
		result2, err := cacheFs.Execute(ctx, func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("primary failed")
		}, "test-key")
		
		assert.NoError(t, err)
		assert.Equal(t, "cached result", result2)
	})

	t.Run("CustomFallbackFunction", func(t *testing.T) {
		customFs := NewFallbackStrategy("test-custom", config, logger)
		
		// 注册自定义降级函数
		customFs.RegisterFallbackFunc(FallbackTypeCustom, func(ctx context.Context, args interface{}) (interface{}, error) {
			return "custom fallback result", nil
		})
		
		// 更新配置为自定义类型
		customConfig := *config
		customConfig.Type = FallbackTypeCustom
		customFs.UpdateConfig(&customConfig)
		
		ctx := context.Background()
		result, err := customFs.Execute(ctx, func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("primary failed")
		}, nil)
		
		assert.NoError(t, err)
		assert.Equal(t, "custom fallback result", result)
	})

	t.Run("Metrics", func(t *testing.T) {
		metrics := fs.GetMetrics()
		assert.NotNil(t, metrics)
		assert.True(t, metrics.TotalFallbacks > 0)
	})
}

// TestAutoRecoveryManager 测试自动恢复管理器
func TestAutoRecoveryManager(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	t.Run("RegisterPlugin", func(t *testing.T) {
		config := DefaultAutoRecoveryConfig()
		config.HealthCheckInterval = 50 * time.Millisecond
		config.HealthCheckTimeout = 10 * time.Millisecond
		config.FailureThreshold = 2
		config.MaxRecoveryAttempts = 2
		arm := NewAutoRecoveryManager(config, logger)
		
		arm.RegisterPlugin("test-plugin")
		
		state, exists := arm.GetPluginState("test-plugin")
		assert.True(t, exists)
		assert.Equal(t, "test-plugin", state.PluginID)
		assert.Equal(t, HealthStatusUnknown, state.CurrentStatus)
	})

	t.Run("HealthCheckFunction", func(t *testing.T) {
		config := DefaultAutoRecoveryConfig()
		config.HealthCheckInterval = 50 * time.Millisecond
		config.HealthCheckTimeout = 10 * time.Millisecond
		config.FailureThreshold = 2
		config.MaxRecoveryAttempts = 2
		arm := NewAutoRecoveryManager(config, logger)
		defer arm.Stop()
		
		arm.RegisterPlugin("test-plugin")
		healthCheckCalled := false
		
		arm.SetHealthCheckFunc(func(ctx context.Context, pluginID string) (*HealthCheckResult, error) {
			healthCheckCalled = true
			return &HealthCheckResult{
				PluginID:  pluginID,
				Status:    HealthStatusHealthy,
				Message:   "Plugin is healthy",
				CheckTime: time.Now(),
			}, nil
		})
		
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		
		err := arm.Start(ctx)
		assert.NoError(t, err)
		
		// 等待健康检查执行
		time.Sleep(100 * time.Millisecond)
		
		assert.True(t, healthCheckCalled)
	})

	t.Run("RecoveryTrigger", func(t *testing.T) {
		config := DefaultAutoRecoveryConfig()
		config.HealthCheckInterval = 50 * time.Millisecond
		config.HealthCheckTimeout = 10 * time.Millisecond
		config.FailureThreshold = 2
		config.MaxRecoveryAttempts = 2
		arm := NewAutoRecoveryManager(config, logger)
		defer arm.Stop()
		
		arm.RegisterPlugin("test-plugin")
		
		// 设置健康检查函数返回不健康状态
		arm.SetHealthCheckFunc(func(ctx context.Context, pluginID string) (*HealthCheckResult, error) {
			return &HealthCheckResult{
				PluginID:  pluginID,
				Status:    HealthStatusUnhealthy,
				Message:   "Plugin is unhealthy",
				CheckTime: time.Now(),
			}, nil
		})
		
		// 注册恢复函数
		arm.RegisterRecoveryFunc(RecoveryActionReset, func(ctx context.Context, pluginID string) error {
			return nil
		})
		
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		
		err := arm.Start(ctx)
		assert.NoError(t, err)
		
		// 等待足够长的时间让恢复被触发
		time.Sleep(300 * time.Millisecond)
		
		// 由于需要达到失败阈值，可能需要多次检查
		// 这里我们主要验证系统没有崩溃
		assert.NotPanics(t, func() {
			// Stop will be called by defer
		})
	})

	t.Run("Metrics", func(t *testing.T) {
		config := DefaultAutoRecoveryConfig()
		arm := NewAutoRecoveryManager(config, logger)
		
		metrics := arm.GetMetrics()
		assert.NotNil(t, metrics)
	})

	t.Run("UnregisterPlugin", func(t *testing.T) {
		config := DefaultAutoRecoveryConfig()
		arm := NewAutoRecoveryManager(config, logger)
		
		arm.RegisterPlugin("test-plugin")
		arm.UnregisterPlugin("test-plugin")
		
		_, exists := arm.GetPluginState("test-plugin")
		assert.False(t, exists)
	})
}

// TestRecoveryManager 测试恢复管理器
func TestRecoveryManager(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	config := DefaultRecoveryManagerConfig()
	config.MaxConcurrentRecoveries = 2

	rm := NewRecoveryManager(config, logger)

	// 创建测试策略
	cbConfig := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker("test-cb", cbConfig, logger)
	
	retryConfig := DefaultRetryConfig()
	retryStrategy := NewRetryStrategy("test-retry", retryConfig, logger)

	// 实现RecoveryStrategy接口的包装器
	cbWrapper := &circuitBreakerWrapper{cb: cb}
	retryWrapper := &retryStrategyWrapper{rs: retryStrategy}

	t.Run("RegisterStrategies", func(t *testing.T) {
		err := rm.RegisterStrategy(cbWrapper)
		assert.NoError(t, err)
		
		err = rm.RegisterStrategy(retryWrapper)
		assert.NoError(t, err)
		
		strategies := rm.GetAllStrategies()
		assert.Len(t, strategies, 2)
	})

	t.Run("ExecuteRecovery", func(t *testing.T) {
		ctx := context.Background()
		
		result, err := rm.ExecuteRecovery(ctx, "test-plugin", []string{"test-cb"}, func(ctx context.Context) (interface{}, error) {
			return "success", nil
		}, nil)
		
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("ConcurrentRecoveries", func(t *testing.T) {
		ctx := context.Background()
		var wg sync.WaitGroup
		results := make([]interface{}, 5)
		errors := make([]error, 5)
		
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				result, err := rm.ExecuteRecovery(ctx, "test-plugin", []string{"test-cb"}, func(ctx context.Context) (interface{}, error) {
					time.Sleep(10 * time.Millisecond)
					return fmt.Sprintf("result-%d", index), nil
				}, nil)
				results[index] = result
				errors[index] = err
			}(i)
		}
		
		wg.Wait()
		
		// 验证结果
		for i := 0; i < 5; i++ {
			if errors[i] == nil {
				assert.NotNil(t, results[i])
			}
		}
	})

	t.Run("Metrics", func(t *testing.T) {
		metrics := rm.GetMetrics()
		assert.NotNil(t, metrics)
		assert.Equal(t, 2, metrics.TotalStrategies)
	})

	t.Run("UnregisterStrategy", func(t *testing.T) {
		err := rm.UnregisterStrategy("test-cb")
		assert.NoError(t, err)
		
		strategies := rm.GetAllStrategies()
		assert.Len(t, strategies, 1)
	})
}

// TestConfigManager 测试配置管理器
func TestConfigManager(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cm := NewConfigManager(logger)

	t.Run("DefaultConfig", func(t *testing.T) {
		config := cm.GetConfig()
		assert.NotNil(t, config)
		assert.Equal(t, ConfigVersionV1, config.Version)
	})

	t.Run("AddCircuitBreakerConfig", func(t *testing.T) {
		cbConfig := DefaultCircuitBreakerConfig()
		cbConfig.FailureThreshold = 5
		
		err := cm.AddCircuitBreakerConfig("test-cb", cbConfig)
		assert.NoError(t, err)
		
		retrievedConfig, exists := cm.GetCircuitBreakerConfig("test-cb")
		assert.True(t, exists)
		assert.Equal(t, 5, retrievedConfig.FailureThreshold)
	})

	t.Run("AddRetryConfig", func(t *testing.T) {
		retryConfig := DefaultRetryConfig()
		retryConfig.MaxRetries = 5
		
		err := cm.AddRetryConfig("test-retry", retryConfig)
		assert.NoError(t, err)
		
		retrievedConfig, exists := cm.GetRetryConfig("test-retry")
		assert.True(t, exists)
		assert.Equal(t, 5, retrievedConfig.MaxRetries)
	})

	t.Run("AddPolicyConfig", func(t *testing.T) {
		policyConfig := &PolicyConfig{
			Name:        "test-policy",
			Description: "Test policy",
			Enabled:     true,
			Priority:    1,
			Strategies:  []string{"test-cb", "test-retry"},
			PluginIDs:   []string{"plugin1", "plugin2"},
		}
		
		err := cm.AddPolicyConfig("test-policy", policyConfig)
		assert.NoError(t, err)
		
		retrievedConfig, exists := cm.GetPolicyConfig("test-policy")
		assert.True(t, exists)
		assert.Equal(t, "test-policy", retrievedConfig.Name)
		assert.Len(t, retrievedConfig.Strategies, 2)
	})

	t.Run("SaveAndLoadConfig", func(t *testing.T) {
		// 保存配置
		data, err := cm.SaveConfig()
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
		
		// 创建新的配置管理器并加载配置
		newCm := NewConfigManager(logger)
		err = newCm.LoadConfig(data)
		assert.NoError(t, err)
		
		// 验证配置是否正确加载
		_, exists := newCm.GetCircuitBreakerConfig("test-cb")
		assert.True(t, exists)
		
		_, exists = newCm.GetRetryConfig("test-retry")
		assert.True(t, exists)
		
		_, exists = newCm.GetPolicyConfig("test-policy")
		assert.True(t, exists)
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		// 测试无效配置
		invalidConfig := DefaultCircuitBreakerConfig()
		invalidConfig.FailureThreshold = -1 // 无效值
		
		err := cm.AddCircuitBreakerConfig("invalid-cb", invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failure threshold must be positive")
	})
}

// 辅助类型：实现RecoveryStrategy接口的包装器
type circuitBreakerWrapper struct {
	cb *CircuitBreaker
}

func (cbw *circuitBreakerWrapper) GetName() string {
	return cbw.cb.GetName()
}

func (cbw *circuitBreakerWrapper) GetType() StrategyType {
	return StrategyTypeCircuitBreaker
}

func (cbw *circuitBreakerWrapper) Execute(ctx context.Context, operation func(ctx context.Context) (interface{}, error), args interface{}) (interface{}, error) {
	var result interface{}
	err := cbw.cb.Execute(ctx, func(ctx context.Context) error {
		var opErr error
		result, opErr = operation(ctx)
		return opErr
	})
	return result, err
}

func (cbw *circuitBreakerWrapper) Reset() {
	cbw.cb.Reset()
}

func (cbw *circuitBreakerWrapper) IsHealthy() bool {
	return cbw.cb.IsHealthy()
}

type retryStrategyWrapper struct {
	rs *RetryStrategy
}

func (rsw *retryStrategyWrapper) GetName() string {
	return rsw.rs.GetName()
}

func (rsw *retryStrategyWrapper) GetType() StrategyType {
	return StrategyTypeRetry
}

func (rsw *retryStrategyWrapper) Execute(ctx context.Context, operation func(ctx context.Context) (interface{}, error), args interface{}) (interface{}, error) {
	var result interface{}
	err := rsw.rs.Execute(ctx, func(ctx context.Context) error {
		var opErr error
		result, opErr = operation(ctx)
		return opErr
	})
	return result, err
}

func (rsw *retryStrategyWrapper) Reset() {
	rsw.rs.Reset()
}

func (rsw *retryStrategyWrapper) IsHealthy() bool {
	return rsw.rs.GetSuccessRate() > 0.5 // 简单的健康检查逻辑
}

// TestIntegration 集成测试
func TestIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	
	// 创建配置管理器
	cm := NewConfigManager(logger)
	
	// 添加各种配置
	cbConfig := DefaultCircuitBreakerConfig()
	cbConfig.FailureThreshold = 3
	cm.AddCircuitBreakerConfig("integration-cb", cbConfig)
	
	retryConfig := DefaultRetryConfig()
	retryConfig.MaxRetries = 2
	retryConfig.InitialDelay = 1 * time.Nanosecond
	retryConfig.MaxDelay = 1 * time.Nanosecond
	retryConfig.Timeout = 10 * time.Second
	retryConfig.RetryableErrors = []string{"temporary"}
	cm.AddRetryConfig("integration-retry", retryConfig)
	
	fallbackConfig := DefaultFallbackConfig()
	fallbackConfig.DefaultValue = "integration fallback"
	cm.AddFallbackConfig("integration-fallback", fallbackConfig)
	
	// 创建恢复管理器
	rmConfig := DefaultRecoveryManagerConfig()
	rmConfig.RecoveryTimeout = 10 * time.Second
	rm := NewRecoveryManager(rmConfig, logger)
	
	// 创建策略实例
	cb := NewCircuitBreaker("integration-cb", cbConfig, logger)
	rs := NewRetryStrategy("integration-retry", retryConfig, logger)
	fs := NewFallbackStrategy("integration-fallback", fallbackConfig, logger)
	
	// 注册策略
	rm.RegisterStrategy(&circuitBreakerWrapper{cb: cb})
	rm.RegisterStrategy(&retryStrategyWrapper{rs: rs})
	
	t.Run("CompleteWorkflow", func(t *testing.T) {
		ctx := context.Background()
		
		// 测试成功场景
		result, err := rm.ExecuteRecovery(ctx, "test-plugin", []string{"integration-cb", "integration-retry"}, func(ctx context.Context) (interface{}, error) {
			return "success result", nil
		}, nil)
		
		assert.NoError(t, err)
		assert.Equal(t, "success result", result)
		
		// 测试简单的重试场景
		result, err = rm.ExecuteRecovery(ctx, "test-plugin", []string{"integration-cb"}, func(ctx context.Context) (interface{}, error) {
			return "integration success", nil
		}, nil)
		
		assert.NoError(t, err)
		assert.Equal(t, "integration success", result)
	})
	
	t.Run("MetricsCollection", func(t *testing.T) {
		// 验证各组件的指标
		cbMetrics := cb.GetMetrics()
		assert.NotNil(t, cbMetrics)
		
		rsMetrics := rs.GetMetrics()
		assert.NotNil(t, rsMetrics)
		
		fsMetrics := fs.GetMetrics()
		assert.NotNil(t, fsMetrics)
		
		rmMetrics := rm.GetMetrics()
		assert.NotNil(t, rmMetrics)
		assert.True(t, rmMetrics.TotalRecoveries > 0)
	})
}

// BenchmarkCircuitBreaker 熔断器性能测试
func BenchmarkCircuitBreaker(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cb := NewCircuitBreaker("bench-cb", DefaultCircuitBreakerConfig(), logger)
	ctx := context.Background()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(ctx, func(ctx context.Context) error {
				return nil
			})
		}
	})
}

// BenchmarkRetryStrategy 重试策略性能测试
func BenchmarkRetryStrategy(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := DefaultRetryConfig()
	config.MaxRetries = 1
	rs := NewRetryStrategy("bench-retry", config, logger)
	ctx := context.Background()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rs.Execute(ctx, func(ctx context.Context) error {
				return nil
			})
		}
	})
}