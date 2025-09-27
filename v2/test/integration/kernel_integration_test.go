package integration

import (
	"context"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// KernelIntegrationTestSuite 微内核集成测试套件
type KernelIntegrationTestSuite struct {
	suite.Suite
	kernel kernel.Kernel
	ctx    context.Context
	cancel context.CancelFunc
}

// SetupSuite 设置测试套件
func (suite *KernelIntegrationTestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 30*time.Second)
}

// TearDownSuite 清理测试套件
func (suite *KernelIntegrationTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// SetupTest 设置每个测试
func (suite *KernelIntegrationTestSuite) SetupTest() {
	suite.kernel = kernel.NewMicroKernel()
}

// TearDownTest 清理每个测试
func (suite *KernelIntegrationTestSuite) TearDownTest() {
	if suite.kernel != nil {
		_ = suite.kernel.Shutdown(suite.ctx)
	}
}

// TestKernelLifecycle 测试微内核完整生命周期
func (suite *KernelIntegrationTestSuite) TestKernelLifecycle() {
	// 1. 测试初始状态
	assert.False(suite.T(), suite.kernel.IsRunning())
	status := suite.kernel.GetStatus()
	assert.Equal(suite.T(), kernel.KernelStateUninitialized, status.State)

	// 2. 测试初始化
	err := suite.kernel.Initialize(suite.ctx)
	assert.NoError(suite.T(), err)
	status = suite.kernel.GetStatus()
	assert.Equal(suite.T(), kernel.KernelStateInitialized, status.State)

	// 3. 测试启动
	err = suite.kernel.Start(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), suite.kernel.IsRunning())
	status = suite.kernel.GetStatus()
	assert.Equal(suite.T(), kernel.KernelStateRunning, status.State)
	assert.True(suite.T(), status.StartedAt > 0)

	// 4. 测试停止
	err = suite.kernel.Stop(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), suite.kernel.IsRunning())
	status = suite.kernel.GetStatus()
	assert.Equal(suite.T(), kernel.KernelStateStopped, status.State)

	// 5. 测试关闭
	err = suite.kernel.Shutdown(suite.ctx)
	assert.NoError(suite.T(), err)
}

// TestCoreComponentsIntegration 测试核心组件集成
func (suite *KernelIntegrationTestSuite) TestCoreComponentsIntegration() {
	// 初始化和启动内核
	err := suite.kernel.Initialize(suite.ctx)
	assert.NoError(suite.T(), err)
	err = suite.kernel.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	// 测试获取核心组件
	pluginManager := suite.kernel.GetPluginManager()
	assert.NotNil(suite.T(), pluginManager)

	eventBus := suite.kernel.GetEventBus()
	assert.NotNil(suite.T(), eventBus)

	serviceRegistry := suite.kernel.GetServiceRegistry()
	assert.NotNil(suite.T(), serviceRegistry)

	securityManager := suite.kernel.GetSecurityManager()
	assert.NotNil(suite.T(), securityManager)

	config := suite.kernel.GetConfig()
	assert.NotNil(suite.T(), config)

	logger := suite.kernel.GetLogger()
	assert.NotNil(suite.T(), logger)

	container := suite.kernel.GetContainer()
	assert.NotNil(suite.T(), container)
}

// TestEventBusIntegration 测试事件总线集成
func (suite *KernelIntegrationTestSuite) TestEventBusIntegration() {
	// 初始化和启动内核
	err := suite.kernel.Initialize(suite.ctx)
	assert.NoError(suite.T(), err)
	err = suite.kernel.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	eventBus := suite.kernel.GetEventBus()
	assert.NotNil(suite.T(), eventBus)

	// 测试事件发布和订阅
	received := make(chan bool, 1)
	testEventType := event.EventType("test.event")
	testData := map[string]interface{}{"message": "hello world"}

	// 订阅事件
	_, err = eventBus.Subscribe(testEventType, func(ctx context.Context, e event.Event) error {
		received <- true
		return nil
	})
	assert.NoError(suite.T(), err)

	// 创建测试事件
	testEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      testEventType,
		Data:      testData,
		Source:    "test",
		Timestamp: time.Now(),
	}

	// 发布事件
	err = eventBus.Publish(suite.ctx, testEvent)
	assert.NoError(suite.T(), err)

	// 验证事件接收
	select {
	case <-received:
		// 事件接收成功
	case <-time.After(5 * time.Second):
		suite.T().Fatal("Event not received within timeout")
	}
}

// TestServiceRegistryIntegration 测试服务注册表集成
func (suite *KernelIntegrationTestSuite) TestServiceRegistryIntegration() {
	// 初始化和启动内核
	err := suite.kernel.Initialize(suite.ctx)
	assert.NoError(suite.T(), err)
	err = suite.kernel.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	serviceRegistry := suite.kernel.GetServiceRegistry()
	assert.NotNil(suite.T(), serviceRegistry)

	// 测试服务注册和发现
	serviceInfo := &kernel.ServiceInfo{
		ID:      "test-service-1",
		Name:    "test.service",
		Version: "1.0.0",
		Address: "localhost",
		Port:    8080,
		Tags:    []string{"test"},
		Metadata: map[string]string{"env": "test"},
		Weight:  100,
	}

	// 注册服务
	err = serviceRegistry.Register(suite.ctx, serviceInfo)
	assert.NoError(suite.T(), err)

	// 发现服务
	services, err := serviceRegistry.Discover(suite.ctx, "test.service")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), services, 1)
	assert.Equal(suite.T(), "test-service-1", services[0].Info.ID)

	// 注销服务
	err = serviceRegistry.Deregister(suite.ctx, "test-service-1")
	assert.NoError(suite.T(), err)

	// 验证服务已注销
	services, err = serviceRegistry.Discover(suite.ctx, "test.service")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), services, 0)
}

// TestPluginManagerIntegration 测试插件管理器集成
func (suite *KernelIntegrationTestSuite) TestPluginManagerIntegration() {
	// 初始化和启动内核
	err := suite.kernel.Initialize(suite.ctx)
	assert.NoError(suite.T(), err)
	err = suite.kernel.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	pluginManager := suite.kernel.GetPluginManager()
	assert.NotNil(suite.T(), pluginManager)

	// 测试插件管理器基本功能
	loadedPlugins := pluginManager.GetLoadedPlugins()
	assert.NotNil(suite.T(), loadedPlugins)

	pluginCount := pluginManager.GetLoadedPluginCount()
	assert.Equal(suite.T(), len(loadedPlugins), pluginCount)

	// 测试插件查询
	isLoaded := pluginManager.IsPluginLoaded("non-existent-plugin")
	assert.False(suite.T(), isLoaded)
}

// TestConcurrentOperations 测试并发操作
func (suite *KernelIntegrationTestSuite) TestConcurrentOperations() {
	// 初始化和启动内核
	err := suite.kernel.Initialize(suite.ctx)
	assert.NoError(suite.T(), err)
	err = suite.kernel.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	// 并发获取组件
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			// 并发访问各个组件
			pluginManager := suite.kernel.GetPluginManager()
			assert.NotNil(suite.T(), pluginManager)

			eventBus := suite.kernel.GetEventBus()
			assert.NotNil(suite.T(), eventBus)

			serviceRegistry := suite.kernel.GetServiceRegistry()
			assert.NotNil(suite.T(), serviceRegistry)

			status := suite.kernel.GetStatus()
			assert.Equal(suite.T(), kernel.KernelStateRunning, status.State)
		}()
	}

	// 等待所有协程完成
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// 协程完成
		case <-time.After(10 * time.Second):
			suite.T().Fatal("Concurrent operations timeout")
		}
	}
}

// TestKernelRestart 测试内核重启
func (suite *KernelIntegrationTestSuite) TestKernelRestart() {
	// 第一次启动
	err := suite.kernel.Initialize(suite.ctx)
	assert.NoError(suite.T(), err)
	err = suite.kernel.Start(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), suite.kernel.IsRunning())

	// 停止内核
	err = suite.kernel.Stop(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), suite.kernel.IsRunning())

	// 重新启动（应该失败，因为需要重新初始化）
	err = suite.kernel.Start(suite.ctx)
	assert.Error(suite.T(), err)

	// 重新初始化和启动
	suite.kernel = kernel.NewMicroKernel()
	err = suite.kernel.Initialize(suite.ctx)
	assert.NoError(suite.T(), err)
	err = suite.kernel.Start(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), suite.kernel.IsRunning())
}

// TestKernelIntegration 运行微内核集成测试
func TestKernelIntegration(t *testing.T) {
	suite.Run(t, new(KernelIntegrationTestSuite))
}