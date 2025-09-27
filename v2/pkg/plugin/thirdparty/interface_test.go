// Package thirdparty 第三方插件接口测试
package thirdparty

import (
	"context"
	"testing"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestThirdPartyPluginInterface 测试第三方插件接口
func TestThirdPartyPluginInterface(t *testing.T) {
	// 创建插件信息
	info := &core.PluginInfo{
		ID:          "test-wasm-plugin",
		Name:        "Test WASM Plugin",
		Version:     "1.0.0",
		Description: "Test WebAssembly plugin",
		Author:      "Test Author",
		Type:        core.PluginTypeWebAssembly,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 创建模拟WASM模块
	wasmModule := []byte("mock wasm module data")

	// 创建WASM插件
	plugin := NewWASMPlugin(info, wasmModule)
	require.NotNil(t, plugin)

	// 验证插件实现了ThirdPartyPlugin接口
	var _ ThirdPartyPlugin = plugin

	// 测试基本信息
	assert.Equal(t, info.Name, plugin.GetInfo().Name)
	assert.Equal(t, info.Version, plugin.GetInfo().Version)

	// 测试WASM模块
	module := plugin.GetWASMModule()
	assert.Equal(t, wasmModule, module)

	// 测试资源限制
	limits := plugin.GetResourceLimits()
	assert.NotNil(t, limits)
	assert.Greater(t, limits.MaxMemory, int64(0))

	// 测试沙箱配置
	sandboxConfig := plugin.GetSandboxConfig()
	assert.NotNil(t, sandboxConfig)
	assert.True(t, sandboxConfig.Enabled)
}

// TestWASMPluginLifecycle 测试WASM插件生命周期
func TestWASMPluginLifecycle(t *testing.T) {
	info := &core.PluginInfo{
		ID:      "test-lifecycle-plugin",
		Name:    "Test Lifecycle Plugin",
		Version: "1.0.0",
		Type:    core.PluginTypeWebAssembly,
	}

	wasmModule := []byte("mock wasm module")
	plugin := NewWASMPlugin(info, wasmModule)

	// 创建模拟插件上下文
	ctx := createMockPluginContext()

	// 测试初始化
	err := plugin.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, plugin.IsInitialized())

	// 测试启动
	err = plugin.Start()
	assert.NoError(t, err)
	assert.True(t, plugin.IsStarted())

	// 测试健康检查
	err = plugin.HealthCheck()
	assert.NoError(t, err)

	// 测试停止
	err = plugin.Stop()
	assert.NoError(t, err)
	assert.False(t, plugin.IsStarted())

	// 测试清理
	err = plugin.Cleanup()
	assert.NoError(t, err)
	assert.False(t, plugin.IsInitialized())
}

// TestWASMPluginFunctionExecution 测试WASM插件函数执行
func TestWASMPluginFunctionExecution(t *testing.T) {
	info := &core.PluginInfo{
		ID:      "test-execution-plugin",
		Name:    "Test Execution Plugin",
		Version: "1.0.0",
		Type:    core.PluginTypeWebAssembly,
	}

	wasmModule := []byte("mock wasm module")
	plugin := NewWASMPlugin(info, wasmModule)

	// 初始化和启动插件
	ctx := createMockPluginContext()
	err := plugin.Initialize(ctx)
	require.NoError(t, err)

	err = plugin.Start()
	require.NoError(t, err)

	// 测试获取导出函数
	functions := plugin.GetExportedFunctions()
	assert.NotEmpty(t, functions)
	assert.Contains(t, functions, "add")
	assert.Contains(t, functions, "multiply")

	// 测试执行加法函数
	result, err := plugin.ExecuteFunction("add", []interface{}{5.0, 3.0})
	assert.NoError(t, err)
	assert.Equal(t, 8.0, result)

	// 测试执行乘法函数
	result, err = plugin.ExecuteFunction("multiply", []interface{}{4.0, 2.0})
	assert.NoError(t, err)
	assert.Equal(t, 8.0, result)

	// 测试执行不存在的函数
	_, err = plugin.ExecuteFunction("nonexistent", []interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// 清理
	plugin.Stop()
	plugin.Cleanup()
}

// TestResourceLimits 测试资源限制
func TestResourceLimits(t *testing.T) {
	info := &core.PluginInfo{
		ID:      "test-resource-plugin",
		Name:    "Test Resource Plugin",
		Version: "1.0.0",
		Type:    core.PluginTypeWebAssembly,
	}

	wasmModule := []byte("mock wasm module")
	plugin := NewWASMPlugin(info, wasmModule)

	// 测试默认资源限制
	limits := plugin.GetResourceLimits()
	assert.NotNil(t, limits)
	assert.Greater(t, limits.MaxMemory, int64(0))
	assert.Greater(t, limits.MaxCPU, 0.0)

	// 测试设置新的资源限制
	newLimits := &ResourceLimits{
		MaxMemory:     128 * 1024 * 1024, // 128MB
		MaxCPU:        0.8,               // 80%
		MaxDiskIO:     20 * 1024 * 1024,  // 20MB/s
		MaxNetworkIO:  10 * 1024 * 1024,  // 10MB/s
		Timeout:       60 * time.Second,
		MaxGoroutines: 20,
		MaxFileSize:   20 * 1024 * 1024, // 20MB
		MaxOpenFiles:  50,
	}

	err := plugin.SetResourceLimits(newLimits)
	assert.NoError(t, err)

	// 验证资源限制已更新
	updatedLimits := plugin.GetResourceLimits()
	assert.Equal(t, newLimits.MaxMemory, updatedLimits.MaxMemory)
	assert.Equal(t, newLimits.MaxCPU, updatedLimits.MaxCPU)
	assert.Equal(t, newLimits.Timeout, updatedLimits.Timeout)

	// 测试设置nil资源限制
	err = plugin.SetResourceLimits(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSandboxConfig 测试沙箱配置
func TestSandboxConfig(t *testing.T) {
	info := &core.PluginInfo{
		ID:      "test-sandbox-plugin",
		Name:    "Test Sandbox Plugin",
		Version: "1.0.0",
		Type:    core.PluginTypeWebAssembly,
	}

	wasmModule := []byte("mock wasm module")
	plugin := NewWASMPlugin(info, wasmModule)

	// 测试默认沙箱配置
	config := plugin.GetSandboxConfig()
	assert.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Equal(t, IsolationLevelStrict, config.IsolationLevel)

	// 测试设置新的沙箱配置
	newConfig := &SandboxConfig{
		Enabled:          true,
		AllowedPaths:     []string{"/tmp", "/var/tmp"},
		AllowedNetworks:  []string{"localhost", "127.0.0.1"},
		AllowedSyscalls:  []string{"read", "write", "open", "close", "stat"},
		TrustedSources:   []string{"trusted.example.com"},
		IsolationLevel:   IsolationLevelBasic,
		NetworkAccess:    true,
		FileSystemAccess: true,
	}

	err := plugin.SetSandboxConfig(newConfig)
	assert.NoError(t, err)

	// 验证沙箱配置已更新
	updatedConfig := plugin.GetSandboxConfig()
	assert.Equal(t, newConfig.IsolationLevel, updatedConfig.IsolationLevel)
	assert.Equal(t, newConfig.NetworkAccess, updatedConfig.NetworkAccess)
	assert.Equal(t, newConfig.FileSystemAccess, updatedConfig.FileSystemAccess)
	assert.Equal(t, len(newConfig.AllowedPaths), len(updatedConfig.AllowedPaths))

	// 测试设置nil沙箱配置
	err = plugin.SetSandboxConfig(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestResourceMonitoring 测试资源监控
func TestResourceMonitoring(t *testing.T) {
	info := &core.PluginInfo{
		ID:      "test-monitoring-plugin",
		Name:    "Test Monitoring Plugin",
		Version: "1.0.0",
		Type:    core.PluginTypeWebAssembly,
	}

	wasmModule := []byte("mock wasm module")
	plugin := NewWASMPlugin(info, wasmModule)

	// 初始化插件
	ctx := createMockPluginContext()
	err := plugin.Initialize(ctx)
	require.NoError(t, err)

	// 测试启动资源监控
	monitorCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = plugin.StartResourceMonitoring(monitorCtx)
	assert.NoError(t, err)

	// 等待一小段时间让监控器收集数据
	time.Sleep(100 * time.Millisecond)

	// 测试获取资源使用情况
	usage := plugin.GetResourceUsage()
	assert.NotNil(t, usage)
	assert.NotZero(t, usage.LastUpdated)

	// 测试停止资源监控
	err = plugin.StopResourceMonitoring()
	assert.NoError(t, err)

	// 清理
	plugin.Cleanup()
}

// TestPluginCapabilities 测试插件能力
func TestPluginCapabilities(t *testing.T) {
	info := &core.PluginInfo{
		ID:      "test-capabilities-plugin",
		Name:    "Test Capabilities Plugin",
		Version: "1.0.0",
		Type:    core.PluginTypeWebAssembly,
	}

	wasmModule := []byte("mock wasm module")
	plugin := NewWASMPlugin(info, wasmModule)

	// 测试获取插件能力
	capabilities := plugin.GetCapabilities()
	assert.NotEmpty(t, capabilities)
	assert.Contains(t, capabilities, "wasm_execution")
	assert.Contains(t, capabilities, "sandboxed_execution")
	assert.Contains(t, capabilities, "resource_monitoring")
	assert.Contains(t, capabilities, "function_execution")
}

// createMockPluginContext 创建模拟插件上下文
func createMockPluginContext() core.PluginContext {
	return &mockPluginContext{}
}

// mockPluginContext 模拟插件上下文
type mockPluginContext struct{}

func (m *mockPluginContext) GetContext() context.Context {
	return context.Background()
}

func (m *mockPluginContext) GetContainer() core.ServiceRegistry {
	return &mockServiceRegistry{}
}

func (m *mockPluginContext) GetEventBus() core.EventBus {
	return &mockEventBus{}
}

func (m *mockPluginContext) GetServiceRegistry() core.ServiceRegistry {
	return &mockServiceRegistry{}
}

func (m *mockPluginContext) GetLogger() core.Logger {
	return &mockLogger{}
}

func (m *mockPluginContext) GetPluginConfig() core.PluginConfig {
	return &mockPluginConfig{}
}

func (m *mockPluginContext) UpdateConfig(config core.PluginConfig) error {
	return nil
}

func (m *mockPluginContext) GetDataDir() string {
	return "/tmp/test"
}

func (m *mockPluginContext) GetTempDir() string {
	return "/tmp/test/temp"
}

func (m *mockPluginContext) SendMessage(topic string, data interface{}) error {
	return nil
}

func (m *mockPluginContext) Subscribe(topic string, handler core.EventHandler) error {
	return nil
}

func (m *mockPluginContext) Unsubscribe(topic string, handler core.EventHandler) error {
	return nil
}

func (m *mockPluginContext) BroadcastMessage(message interface{}) error {
	return nil
}

func (m *mockPluginContext) GetResourceMonitor() *core.ResourceMonitor {
	return nil
}

func (m *mockPluginContext) GetSecurityManager() *core.SecurityManager {
	return nil
}

func (m *mockPluginContext) GetIsolationGroup() *core.IsolationGroup {
	return nil
}

func (m *mockPluginContext) Shutdown() error {
	return nil
}

// mockServiceRegistry 模拟服务注册表
type mockServiceRegistry struct{}

func (m *mockServiceRegistry) RegisterService(name string, service interface{}) error {
	return nil
}

func (m *mockServiceRegistry) GetService(name string) (interface{}, error) {
	return nil, nil
}

func (m *mockServiceRegistry) UnregisterService(name string) error {
	return nil
}

func (m *mockServiceRegistry) ListServices() []string {
	return []string{}
}

func (m *mockServiceRegistry) HasService(name string) bool {
	return false
}

// mockEventBus 模拟事件总线
type mockEventBus struct{}

func (m *mockEventBus) Publish(eventType string, data interface{}) error {
	return nil
}

func (m *mockEventBus) Subscribe(eventType string, handler core.EventHandler) error {
	return nil
}

func (m *mockEventBus) Unsubscribe(eventType string, handler core.EventHandler) error {
	return nil
}

func (m *mockEventBus) GetSubscriberCount(eventType string) int {
	return 0
}

// mockLogger 模拟日志器
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...interface{}) {}
func (m *mockLogger) Info(msg string, args ...interface{})  {}
func (m *mockLogger) Warn(msg string, args ...interface{})  {}
func (m *mockLogger) Error(msg string, args ...interface{}) {}

// mockPluginConfig 模拟插件配置
type mockPluginConfig struct{}

func (m *mockPluginConfig) GetID() string {
	return "test-plugin"
}

func (m *mockPluginConfig) GetName() string {
	return "Test Plugin"
}

func (m *mockPluginConfig) GetVersion() string {
	return "1.0.0"
}

func (m *mockPluginConfig) GetEnabled() bool {
	return true
}

func (m *mockPluginConfig) GetPriority() core.PluginPriority {
	return core.PluginPriorityNormal
}

func (m *mockPluginConfig) GetDependencies() []string {
	return []string{}
}

func (m *mockPluginConfig) GetResourceLimits() *core.ResourceLimits {
	return nil
}

func (m *mockPluginConfig) GetSecurityConfig() *core.SecurityConfig {
	return nil
}

func (m *mockPluginConfig) GetCustomConfig() map[string]interface{} {
	return make(map[string]interface{})
}

func (m *mockPluginConfig) Validate() error {
	return nil
}