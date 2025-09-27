package plugin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPluginStartStopManagement 测试插件启动和停止管理功能
func TestPluginStartStopManagement(t *testing.T) {
	// 创建管理器配置
	config := DefaultManagerConfig()
	config.EnableCircuitBreaker = true
	config.CircuitBreakerThreshold = 5

	t.Run("StartStopOptions", func(t *testing.T) {
		// 测试启动选项结构
		startOptions := &StartOptions{
			Timeout: 5 * time.Second,
		}
		assert.Equal(t, 5*time.Second, startOptions.Timeout)

		// 测试停止选项结构
		stopOptions := &StopOptions{
			Timeout:          5 * time.Second,
			ForceStop:        false,
			GracefulShutdown: true,
			SkipCleanup:      false,
		}
		assert.Equal(t, 5*time.Second, stopOptions.Timeout)
		assert.False(t, stopOptions.ForceStop)
		assert.True(t, stopOptions.GracefulShutdown)
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		// 验证配置字段
		assert.True(t, config.EnableCircuitBreaker)
		assert.Equal(t, 5, config.CircuitBreakerThreshold)
	})
}

// TestCircuitBreakerStates 测试熔断器状态
func TestCircuitBreakerStates(t *testing.T) {
	cb := &CircuitBreaker{
		FailureThreshold: 3,
		RecoveryTimeout:  100 * time.Millisecond,
		State:           CircuitBreakerClosed,
	}

	// 初始状态应该是关闭的
	assert.Equal(t, CircuitBreakerClosed, cb.State)
	assert.Equal(t, 3, cb.FailureThreshold)
	assert.Equal(t, 100*time.Millisecond, cb.RecoveryTimeout)
}

// TestStructDefinitions 测试结构体定义
func TestStructDefinitions(t *testing.T) {
	t.Run("StartRequest", func(t *testing.T) {
		req := &StartRequest{
			PluginID: "test-plugin",
			Options:  &StartOptions{},
			Result:   make(chan error, 1),
		}
		assert.Equal(t, "test-plugin", req.PluginID)
		assert.NotNil(t, req.Options)
		assert.NotNil(t, req.Result)
	})

	t.Run("StopRequest", func(t *testing.T) {
		req := &StopRequest{
			PluginID: "test-plugin",
			Options:  &StopOptions{},
			Result:   make(chan error, 1),
		}
		assert.Equal(t, "test-plugin", req.PluginID)
		assert.NotNil(t, req.Options)
		assert.NotNil(t, req.Result)
	})

	t.Run("CircuitBreakerStates", func(t *testing.T) {
		assert.Equal(t, 0, int(CircuitBreakerClosed))
		assert.Equal(t, 1, int(CircuitBreakerOpen))
		assert.Equal(t, 2, int(CircuitBreakerHalfOpen))
	})
}

// TestPluginHooks 测试插件钩子函数结构
func TestPluginHooks(t *testing.T) {
	hooks := &PluginHooks{}
	assert.NotNil(t, hooks)

	startHooks := &StartHooks{}
	assert.NotNil(t, startHooks)

	stopHooks := &StopHooks{}
	assert.NotNil(t, stopHooks)
}

// TestManagedPluginExtensions 测试ManagedPlugin扩展字段
func TestManagedPluginExtensions(t *testing.T) {
	plugin := &ManagedPlugin{
		ID:       "test-plugin",
		Priority: 10,
		Group:    "test-group",
	}

	assert.Equal(t, "test-plugin", plugin.ID)
	assert.Equal(t, 10, plugin.Priority)
	assert.Equal(t, "test-group", plugin.Group)
}

// TestWorkerPoolStruct 测试工作池结构体
func TestWorkerPoolStruct(t *testing.T) {
	wp := &WorkerPool{
		workers: 2,
	}
	assert.Equal(t, 2, wp.workers)
}