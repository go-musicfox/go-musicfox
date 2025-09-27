package plugin

import (
	"context"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

// TestPluginContext 测试插件上下文基本功能
func TestPluginContext(t *testing.T) {
	// 创建测试配置
	config := NewBasePluginConfig("test-plugin", "Test Plugin", "1.0.0", PluginTypeDynamicLibrary)
	
	// 创建插件上下文
	ctx := context.Background()
	pluginCtx := NewPluginContext(ctx, config)
	assert.NotNil(t, pluginCtx)
	
	// 测试基本信息
	assert.Equal(t, config, pluginCtx.GetPluginConfig())
	assert.NotNil(t, pluginCtx.GetLogger())
	assert.NotNil(t, pluginCtx.GetContainer())
	assert.NotNil(t, pluginCtx.GetEventBus())
	
	// 测试关闭
	err := pluginCtx.Shutdown()
	assert.NoError(t, err)
}

// TestServiceRegistry 测试服务注册表
func TestServiceRegistry(t *testing.T) {
	registry := NewServiceRegistry()
	
	// 测试服务注册
	testService := "test-service-data"
	err := registry.RegisterService("test-service", testService)
	if err != nil {
		t.Errorf("Failed to register service: %v", err)
	}
	
	// 测试服务获取
	service, err := registry.GetService("test-service")
	if err != nil {
		t.Errorf("Failed to get service: %v", err)
	}
	
	if service != testService {
		t.Errorf("Expected service data '%s', got '%v'", testService, service)
	}
	
	// 测试服务存在检查
	if !registry.HasService("test-service") {
		t.Error("Service should exist")
	}
	
	if registry.HasService("non-existent-service") {
		t.Error("Non-existent service should not exist")
	}
	
	// 测试服务列表
	services := registry.ListServices()
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
	
	// 测试服务注销
	err = registry.UnregisterService("test-service")
	if err != nil {
		t.Errorf("Failed to unregister service: %v", err)
	}
	
	if registry.HasService("test-service") {
		t.Error("Service should not exist after unregistration")
	}
}

// TestEventBus 测试事件总线
func TestEventBus(t *testing.T) {
	eventBus := NewEventBus()
	
	// 使用channel来同步事件处理
	received := make(chan bool, 1)
	handler := func(ctx context.Context, data interface{}) error {
		if data != "test-data" {
			t.Errorf("Expected 'test-data', got '%v'", data)
		}
		select {
		case received <- true:
		default:
			// 如果channel已满，忽略
		}
		return nil
	}
	
	// 订阅事件
	err := eventBus.Subscribe("test-event", handler)
	if err != nil {
		t.Errorf("Failed to subscribe: %v", err)
	}
	
	// 检查订阅者数量
	count := eventBus.GetSubscriberCount("test-event")
	if count != 1 {
		t.Errorf("Expected 1 subscriber, got %d", count)
	}
	
	// 发布事件
	err = eventBus.Publish("test-event", "test-data")
	if err != nil {
		t.Errorf("Failed to publish event: %v", err)
	}
	
	// 等待事件处理
	select {
	case <-received:
		// 事件已接收
	case <-time.After(1 * time.Second):
		t.Error("Event should have been received")
	}
	
	// 测试取消订阅
	err = eventBus.Unsubscribe("test-event", handler)
	if err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
	
	count = eventBus.GetSubscriberCount("test-event")
	if count != 0 {
		t.Errorf("Expected 0 subscribers after unsubscribe, got %d", count)
	}
}

// TestSecurityManager 测试安全管理器
func TestSecurityManager(t *testing.T) {
	config := &SecurityConfig{
		Permissions: []Permission{PermissionEventAccess, PermissionConfigAccess},
		SandboxEnabled: true,
		SecurityLevel: SecurityLevelMedium,
	}
	
	sm := NewSecurityManager(config)
	assert.NotNil(t, sm)
	
	// 测试权限检查
	if !sm.CheckPermission(PermissionConfigAccess) {
		t.Error("Should have config access permission")
	}
	
	if sm.CheckPermission(PermissionFileWrite) {
		t.Error("Should not have file write permission")
	}
	
	// 测试配置更新
	newConfig := &SecurityConfig{
		Permissions: []Permission{PermissionAll},
		SandboxEnabled: false,
		SecurityLevel: SecurityLevelHigh,
	}
	
	err := sm.UpdateConfig(newConfig)
	assert.NoError(t, err)
}

// TestResourceMonitor 测试资源监控器
func TestResourceMonitor(t *testing.T) {
	limits := &ResourceLimits{
		Enabled:         true,
		MaxMemoryMB:     100,
		MaxCPUPercent:   50.0,
		MaxGoroutines:   10,
		EnforceMode:     EnforceModeWarn,
	}
	
	rm := NewResourceMonitor(limits)
	defer rm.Stop()
	
	// 测试指标获取
	metrics := rm.GetMetrics()
	if metrics == nil {
		t.Error("GetMetrics() should not return nil")
	}
	
	// 测试限制更新
	newLimits := &ResourceLimits{
		Enabled:         true,
		MaxMemoryMB:     200,
		MaxCPUPercent:   80.0,
		MaxGoroutines:   20,
		EnforceMode:     EnforceModeLimit,
	}
	
	rm.UpdateLimits(newLimits)
	
	// 等待一段时间让监控器运行
	time.Sleep(100 * time.Millisecond)
}

// TestIsolationGroup 测试隔离组
func TestIsolationGroup(t *testing.T) {
	ig := NewIsolationGroup("test-plugin")
	defer ig.Cleanup()
	
	// 测试资源分配
	testResource := "test-resource-data"
	ig.AllocateResource("test-resource", testResource)
	
	// 测试资源获取
	resource, exists := ig.GetResource("test-resource")
	if !exists {
		t.Error("Resource should exist")
	}
	
	if resource != testResource {
		t.Errorf("Expected resource '%s', got '%v'", testResource, resource)
	}
	
	// 测试不存在的资源
	_, exists = ig.GetResource("non-existent-resource")
	if exists {
		t.Error("Non-existent resource should not exist")
	}
	
	// 测试资源清理
	ig.Cleanup()
	_, exists = ig.GetResource("test-resource")
	if exists {
		t.Error("Resource should not exist after cleanup")
	}
}