package kernel

import (
	"context"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMicroKernel(t *testing.T) {
	kernel := NewMicroKernel()
	require.NotNil(t, kernel)

	mk, ok := kernel.(*MicroKernel)
	require.True(t, ok)

	// 验证初始状态
	assert.Equal(t, KernelStateUninitialized, mk.state)
	assert.NotNil(t, mk.container)
	assert.NotNil(t, mk.logger)
	assert.NotNil(t, mk.config)
	assert.NotNil(t, mk.ctx)
	assert.NotNil(t, mk.cancel)
}

func TestMicroKernelLifecycle(t *testing.T) {
	kernel := NewMicroKernel()
	ctx := context.Background()

	// 测试初始化
	err := kernel.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, kernel.GetStatus().State == KernelStateInitialized)

	// 测试启动
	err = kernel.Start(ctx)
	require.NoError(t, err)
	assert.True(t, kernel.IsRunning())
	assert.True(t, kernel.GetStatus().State == KernelStateRunning)

	// 测试停止
	err = kernel.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, kernel.IsRunning())
	assert.True(t, kernel.GetStatus().State == KernelStateStopped)

	// 测试关闭
	err = kernel.Shutdown(ctx)
	require.NoError(t, err)
}

func TestMicroKernelComponents(t *testing.T) {
	kernel := NewMicroKernel()
	ctx := context.Background()

	// 初始化内核
	err := kernel.Initialize(ctx)
	require.NoError(t, err)

	// 验证组件获取
	assert.NotNil(t, kernel.GetPluginManager())
	assert.NotNil(t, kernel.GetEventBus())
	assert.NotNil(t, kernel.GetServiceRegistry())
	assert.NotNil(t, kernel.GetSecurityManager())
	assert.NotNil(t, kernel.GetConfig())
	assert.NotNil(t, kernel.GetContainer())

	// 清理
	kernel.Shutdown(ctx)
}

func TestDependencyInjectionContainer(t *testing.T) {
	kernel := NewMicroKernel()
	ctx := context.Background()

	// 初始化内核
	err := kernel.Initialize(ctx)
	require.NoError(t, err)

	container := kernel.GetContainer()
	require.NotNil(t, container)

	// 测试从容器中获取组件
	var eventBus event.EventBus
	err = container.Invoke(func(eb event.EventBus) {
		eventBus = eb
	})
	require.NoError(t, err)
	assert.NotNil(t, eventBus)
	assert.Equal(t, kernel.GetEventBus(), eventBus)

	var serviceRegistry ServiceRegistry
	err = container.Invoke(func(sr ServiceRegistry) {
		serviceRegistry = sr
	})
	require.NoError(t, err)
	assert.NotNil(t, serviceRegistry)
	assert.Equal(t, kernel.GetServiceRegistry(), serviceRegistry)

	var securityManager SecurityManager
	err = container.Invoke(func(sm SecurityManager) {
		securityManager = sm
	})
	require.NoError(t, err)
	assert.NotNil(t, securityManager)
	assert.Equal(t, kernel.GetSecurityManager(), securityManager)

	var pluginManager PluginManager
	err = container.Invoke(func(pm PluginManager) {
		pluginManager = pm
	})
	require.NoError(t, err)
	assert.NotNil(t, pluginManager)
	assert.Equal(t, kernel.GetPluginManager(), pluginManager)

	var cfg *koanf.Koanf
	err = container.Invoke(func(c *koanf.Koanf) {
		cfg = c
	})
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// 清理
	kernel.Shutdown(ctx)
}

func TestKernelStatus(t *testing.T) {
	kernel := NewMicroKernel()
	ctx := context.Background()

	// 测试未初始化状态
	status := kernel.GetStatus()
	assert.Equal(t, KernelStateUninitialized, status.State)
	assert.Equal(t, int64(0), status.StartedAt)
	assert.Equal(t, int64(0), status.Uptime)

	// 初始化并启动
	err := kernel.Initialize(ctx)
	require.NoError(t, err)

	err = kernel.Start(ctx)
	require.NoError(t, err)

	// 测试运行状态
	status = kernel.GetStatus()
	assert.Equal(t, KernelStateRunning, status.State)
	assert.True(t, status.StartedAt > 0)
	assert.Equal(t, "1.0.0", status.Version)

	// 等待一小段时间测试运行时间
	time.Sleep(100 * time.Millisecond)
	status = kernel.GetStatus()
	assert.True(t, status.Uptime > 0)

	// 清理
	kernel.Shutdown(ctx)
}

func TestKernelStateTransitions(t *testing.T) {
	kernel := NewMicroKernel()
	ctx := context.Background()

	// 测试重复初始化
	err := kernel.Initialize(ctx)
	require.NoError(t, err)

	err = kernel.Initialize(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already initialized")

	// 测试未启动就停止
	mk := kernel.(*MicroKernel)
	mk.state = KernelStateStopped
	err = kernel.Stop(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// 清理
	kernel.Shutdown(ctx)
}

func TestKernelStateString(t *testing.T) {
	tests := []struct {
		state    KernelState
		expected string
	}{
		{KernelStateUninitialized, "uninitialized"},
		{KernelStateInitialized, "initialized"},
		{KernelStateStarting, "starting"},
		{KernelStateRunning, "running"},
		{KernelStateStopping, "stopping"},
		{KernelStateStopped, "stopped"},
		{KernelStateError, "error"},
		{KernelState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}