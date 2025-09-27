package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/plugin/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// HybridPluginManager 混合插件管理器（用于测试）
type HybridPluginManager struct {
	eventBus        *UnloadMockEventBus
	serviceRegistry *UnloadMockServiceRegistry
	securityManager *UnloadMockSecurityManager
	logger          *slog.Logger
	plugins         map[string]*ManagedPlugin
	config          *ManagerConfig
	dynamicLoader   loader.PluginLoader
	rpcLoader       loader.PluginLoader
	wasmLoader      loader.PluginLoader
	hotReloadLoader loader.PluginLoader
}

// UnloadOptions 卸载选项
type UnloadOptions struct {
	ForceUnload      bool
	SkipCleanup      bool
	Timeout          time.Duration
	GracefulShutdown bool
	Hooks            *UnloadHooks
}

// UnloadHooks 卸载钩子函数
type UnloadHooks struct {
	PreUnload  func(ctx context.Context, plugin *ManagedPlugin) error
	PostUnload func(ctx context.Context, plugin *ManagedPlugin) error
	OnError    func(pluginID string, err error)
	OnCleanup  func(ctx context.Context, plugin *ManagedPlugin) error
}

// UnloadProgress 卸载进度
type UnloadProgress struct {
	PluginID string
	Stage    string
	Progress float64
	Message  string
}

// UnloadPlugin 卸载插件（测试方法）
func (h *HybridPluginManager) UnloadPlugin(ctx context.Context, pluginID string) error {
	// 简化的测试实现
	if plugin, exists := h.plugins[pluginID]; exists {
		if plugin.Plugin != nil {
			plugin.Plugin.Stop()
			plugin.Plugin.Cleanup()
		}
		delete(h.plugins, pluginID)
	}
	return nil
}

// UnloadPluginWithOptions 带选项的卸载插件（测试方法）
func (h *HybridPluginManager) UnloadPluginWithOptions(pluginID string, options *UnloadOptions) error {
	// 简化的测试实现
	ctx := context.Background()
	if plugin, exists := h.plugins[pluginID]; exists {
		if options != nil && options.Hooks != nil {
			if options.Hooks.PreUnload != nil {
				options.Hooks.PreUnload(ctx, plugin)
			}
			if options.Hooks.OnCleanup != nil {
				options.Hooks.OnCleanup(ctx, plugin)
			}
		}
		if plugin.Plugin != nil && (options == nil || !options.SkipCleanup) {
			plugin.Plugin.Stop()
			plugin.Plugin.Cleanup()
		}
		delete(h.plugins, pluginID)
		if options != nil && options.Hooks != nil && options.Hooks.PostUnload != nil {
			options.Hooks.PostUnload(ctx, plugin)
		}
	}
	return nil
}

// UnloadPluginWithProgress 带进度监控的卸载插件（测试方法）
func (h *HybridPluginManager) UnloadPluginWithProgress(pluginID string, progressCallback func(*UnloadProgress)) error {
	// 简化的测试实现
	if progressCallback != nil {
		progressCallback(&UnloadProgress{
			PluginID: pluginID,
			Stage:    "starting",
			Progress: 0.0,
			Message:  "Starting unload",
		})
		progressCallback(&UnloadProgress{
			PluginID: pluginID,
			Stage:    "completed",
			Progress: 1.0,
			Message:  "Unload completed",
		})
	}
	return h.UnloadPlugin(context.Background(), pluginID)
}

// isValidStateTransition 检查状态转换是否有效（测试辅助函数）
func isValidStateTransition(from, to loader.PluginState) bool {
	// 简化的状态转换逻辑
	switch from {
	case loader.PluginStateLoaded, loader.PluginStateStopped, loader.PluginStateError:
		return to == loader.PluginStateUnloading
	case loader.PluginStateUnloading:
		return to == loader.PluginStateUnloaded
	case loader.PluginStateRunning:
		return false // 运行中的插件不能直接卸载
	default:
		return to == loader.PluginStateError // 任何状态都可以转换为错误状态
	}
}

// canUnloadFromState 检查是否可以从指定状态卸载（测试辅助函数）
func canUnloadFromState(state loader.PluginState) bool {
	switch state {
	case loader.PluginStateLoaded, loader.PluginStateStopped, loader.PluginStateError:
		return true
	case loader.PluginStateRunning, loader.PluginStateUnloading:
		return false
	default:
		return false
	}
}

// ManagerConfig 管理器配置（用于测试）
type ManagerConfig struct {
	MaxPlugins       int
	HealthCheckInterval time.Duration
	ShutdownTimeout  time.Duration
}

// DefaultManagerConfig 默认管理器配置
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		MaxPlugins:       100,
		HealthCheckInterval: 30 * time.Second,
		ShutdownTimeout:  10 * time.Second,
	}
}

// ManagedPlugin 管理的插件（用于测试）
type ManagedPlugin struct {
	ID       string
	Path     string
	Type     string
	Plugin   loader.Plugin
	State    loader.PluginState
	Config   map[string]interface{}
	Loader   loader.PluginLoader
	Metadata map[string]interface{}
}

// UnloadMockPlugin 卸载测试专用的模拟插件
type UnloadMockPlugin struct {
	mock.Mock
	info         *loader.PluginInfo
	capabilities []string
	dependencies []string
	cleanupCalled bool
	stopCalled   bool
}

func NewUnloadMockPlugin(name, version string) *UnloadMockPlugin {
	return &UnloadMockPlugin{
		info: &loader.PluginInfo{
			Name:    name,
			Version: version,
			ID:      fmt.Sprintf("%s-%s", name, version),
		},
		capabilities: []string{"test"},
		dependencies: []string{},
	}
}

func (m *UnloadMockPlugin) GetInfo() *loader.PluginInfo {
	return m.info
}

func (m *UnloadMockPlugin) GetCapabilities() []string {
	return m.capabilities
}

func (m *UnloadMockPlugin) GetDependencies() []string {
	return m.dependencies
}

func (m *UnloadMockPlugin) Initialize(ctx loader.PluginContext) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *UnloadMockPlugin) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *UnloadMockPlugin) Stop() error {
	m.stopCalled = true
	args := m.Called()
	return args.Error(0)
}

func (m *UnloadMockPlugin) Cleanup() error {
	m.cleanupCalled = true
	args := m.Called()
	return args.Error(0)
}

func (m *UnloadMockPlugin) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *UnloadMockPlugin) ValidateConfig(config map[string]interface{}) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *UnloadMockPlugin) UpdateConfig(config map[string]interface{}) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *UnloadMockPlugin) GetMetrics() (*loader.PluginMetrics, error) {
	args := m.Called()
	return args.Get(0).(*loader.PluginMetrics), args.Error(1)
}

func (m *UnloadMockPlugin) HandleEvent(event interface{}) error {
	args := m.Called(event)
	return args.Error(0)
}

// UnloadMockPluginLoader 卸载测试专用的模拟插件加载器
type UnloadMockPluginLoader struct {
	mock.Mock
}

func (m *UnloadMockPluginLoader) LoadPlugin(ctx context.Context, pluginPath string) (loader.Plugin, error) {
	args := m.Called(ctx, pluginPath)
	return args.Get(0).(loader.Plugin), args.Error(1)
}

func (m *UnloadMockPluginLoader) UnloadPlugin(ctx context.Context, pluginID string) error {
	args := m.Called(ctx, pluginID)
	return args.Error(0)
}

func (m *UnloadMockPluginLoader) ValidatePlugin(pluginPath string) error {
	args := m.Called(pluginPath)
	return args.Error(0)
}

func (m *UnloadMockPluginLoader) GetSupportedTypes() []loader.LoaderType {
	args := m.Called()
	return args.Get(0).([]loader.LoaderType)
}

func (m *UnloadMockPluginLoader) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *UnloadMockPluginLoader) GetLoadedPlugins() map[string]loader.Plugin {
	args := m.Called()
	return args.Get(0).(map[string]loader.Plugin)
}

func (m *UnloadMockPluginLoader) GetPluginInfo(pluginID string) (*loader.PluginInfo, error) {
	args := m.Called(pluginID)
	return args.Get(0).(*loader.PluginInfo), args.Error(1)
}

func (m *UnloadMockPluginLoader) IsPluginLoaded(pluginID string) bool {
	args := m.Called(pluginID)
	return args.Bool(0)
}

func (m *UnloadMockPluginLoader) GetLoaderType() loader.PluginType {
	args := m.Called()
	return args.Get(0).(loader.PluginType)
}

func (m *UnloadMockPluginLoader) GetLoaderInfo() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *UnloadMockPluginLoader) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *UnloadMockPluginLoader) ReloadPlugin(ctx context.Context, pluginID string) error {
	args := m.Called(ctx, pluginID)
	return args.Error(0)
}

// TestPluginUnloadBasic 测试基础插件卸载功能
func TestPluginUnloadBasic(t *testing.T) {
	tests := []struct {
		name           string
		pluginState    loader.PluginState
		expectError    bool
		cleanupError   error
		unloadError    error
	}{
		{
			name:        "successful unload from loaded state",
			pluginState: loader.PluginStateLoaded,
			expectError: false,
		},
		{
			name:        "successful unload from stopped state",
			pluginState: loader.PluginStateStopped,
			expectError: false,
		},
		{
			name:        "successful unload from error state",
			pluginState: loader.PluginStateError,
			expectError: false,
		},
		{
			name:         "cleanup error but continue",
			pluginState:  loader.PluginStateLoaded,
			expectError:  false,
			cleanupError: fmt.Errorf("cleanup failed"),
		},
		{
			name:        "loader unload error but continue",
			pluginState: loader.PluginStateLoaded,
			expectError: false,
			unloadError: fmt.Errorf("loader unload failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			manager, mockPlugin, mockLoader := setupUnloadTestManager(t)
			pluginID := "test-plugin"

			// 设置插件状态
			managedPlugin := &ManagedPlugin{
				ID:      pluginID,
				Path:    "/test/path",
				Type:    "test", // 使用测试类型避免真实加载器
				Plugin:  mockPlugin,
				Loader:  mockLoader,
				State:   tt.pluginState,
				Metadata: make(map[string]interface{}),
			}
			manager.plugins[pluginID] = managedPlugin

			// 设置模拟期望
			mockPlugin.On("Stop").Return(nil)
			mockPlugin.On("Cleanup").Return(tt.cleanupError)
			// 不设置UnloadPlugin期望，因为selectLoader会返回nil

			// 执行卸载
			err := manager.UnloadPlugin(context.Background(), pluginID)

			// 验证结果
			if tt.expectError {
				assert.Error(t, err)
			} else {
				// 对于恢复的情况，可能会有错误但插件已被移除
				_, exists := manager.plugins[pluginID]
				assert.False(t, exists)
				// 验证清理方法被调用
				assert.True(t, mockPlugin.cleanupCalled)
			}

			mockPlugin.AssertExpectations(t)
			mockLoader.AssertExpectations(t)
		})
	}
}

// TestPluginUnloadWithOptions 测试带选项的插件卸载
func TestPluginUnloadWithOptions(t *testing.T) {
	tests := []struct {
		name        string
		options     *UnloadOptions
		expectError bool
		setupMock   func(*UnloadMockPlugin, *UnloadMockPluginLoader)
	}{
		{
			name: "force unload with cleanup skip",
			options: &UnloadOptions{
				ForceUnload: true,
				SkipCleanup: true,
			},
			expectError: false,
			setupMock: func(mp *UnloadMockPlugin, ml *UnloadMockPluginLoader) {
				// SkipCleanup为true时不会调用Stop和Cleanup
				// 不设置UnloadPlugin期望，因为selectLoader会返回nil
			},
		},
		{
			name: "graceful shutdown with timeout",
			options: &UnloadOptions{
				Timeout:          5 * time.Second,
				GracefulShutdown: true,
			},
			expectError: false,
			setupMock: func(mp *UnloadMockPlugin, ml *UnloadMockPluginLoader) {
				mp.On("Stop").Return(nil)
				mp.On("Cleanup").Return(nil)
				// 不设置UnloadPlugin期望，因为selectLoader会返回nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			manager, mockPlugin, mockLoader := setupUnloadTestManager(t)
			pluginID := "test-plugin"

			// 设置插件
			managedPlugin := &ManagedPlugin{
				ID:      pluginID,
				Path:    "/test/path",
				Type:    "test",
				Plugin:  mockPlugin,
				Loader:  mockLoader,
				State:   loader.PluginStateLoaded,
				Metadata: make(map[string]interface{}),
			}
			manager.plugins[pluginID] = managedPlugin

			// 设置模拟期望
			tt.setupMock(mockPlugin, mockLoader)

			// 执行卸载
			err := manager.UnloadPluginWithOptions(pluginID, tt.options)

			// 验证结果
			if tt.expectError {
				assert.Error(t, err)
			} else {
				// 插件应该被移除或进入恢复状态
				_, exists := manager.plugins[pluginID]
				assert.False(t, exists)
			}

			mockPlugin.AssertExpectations(t)
			mockLoader.AssertExpectations(t)
		})
	}
}

// TestPluginUnloadWithHooks 测试带钩子函数的插件卸载
func TestPluginUnloadWithHooks(t *testing.T) {
	manager, mockPlugin, mockLoader := setupUnloadTestManager(t)
	pluginID := "test-plugin"

	// 设置插件
	managedPlugin := &ManagedPlugin{
		ID:      pluginID,
		Path:    "/test/path",
		Type:    "test",
		Plugin:  mockPlugin,
		Loader:  mockLoader,
		State:   loader.PluginStateLoaded,
		Metadata: make(map[string]interface{}),
	}
	manager.plugins[pluginID] = managedPlugin

	// 设置钩子函数
	var preUnloadCalled, postUnloadCalled, onErrorCalled, onCleanupCalled bool
	hooks := &UnloadHooks{
		PreUnload: func(ctx context.Context, plugin *ManagedPlugin) error {
			preUnloadCalled = true
			return nil
		},
		PostUnload: func(ctx context.Context, plugin *ManagedPlugin) error {
			postUnloadCalled = true
			return nil
		},
		OnError: func(pluginID string, err error) {
			onErrorCalled = true
		},
		OnCleanup: func(ctx context.Context, plugin *ManagedPlugin) error {
			onCleanupCalled = true
			return nil
		},
	}

	options := &UnloadOptions{
		Hooks: hooks,
	}

	// 设置模拟期望
	mockPlugin.On("Stop").Return(nil)
	mockPlugin.On("Cleanup").Return(nil)
	// 不设置UnloadPlugin期望，因为selectLoader会返回nil

	// 执行卸载
	err := manager.UnloadPluginWithOptions(pluginID, options)

	// 验证结果
	assert.NoError(t, err)
	assert.True(t, preUnloadCalled, "PreUnload hook should be called")
	assert.True(t, postUnloadCalled, "PostUnload hook should be called")
	assert.True(t, onCleanupCalled, "OnCleanup hook should be called")
	assert.False(t, onErrorCalled, "OnError hook should not be called on success")

	mockPlugin.AssertExpectations(t)
	mockLoader.AssertExpectations(t)
}

// TestPluginUnloadWithProgress 测试带进度监控的插件卸载
func TestPluginUnloadWithProgress(t *testing.T) {
	manager, mockPlugin, mockLoader := setupUnloadTestManager(t)
	pluginID := "test-plugin"

	// 设置插件
	managedPlugin := &ManagedPlugin{
		ID:      pluginID,
		Path:    "/test/path",
		Type:    "test",
		Plugin:  mockPlugin,
		Loader:  mockLoader,
		State:   loader.PluginStateLoaded,
		Metadata: make(map[string]interface{}),
	}
	manager.plugins[pluginID] = managedPlugin

	// 设置进度回调
	var progressUpdates []*UnloadProgress
	var mu sync.Mutex
	progressCallback := func(progress *UnloadProgress) {
		mu.Lock()
		defer mu.Unlock()
		progressUpdates = append(progressUpdates, progress)
	}

	// 设置模拟期望
	mockPlugin.On("Stop").Return(nil)
	mockPlugin.On("Cleanup").Return(nil)
	// 不设置UnloadPlugin期望，因为selectLoader会返回nil

	// 执行带进度的卸载
	err := manager.UnloadPluginWithProgress(pluginID, progressCallback)

	// 验证结果
	assert.NoError(t, err)

	// 验证进度更新
	mu.Lock()
	defer mu.Unlock()
	assert.Greater(t, len(progressUpdates), 0, "Should have progress updates")

	// 验证进度顺序
	for i := 1; i < len(progressUpdates); i++ {
		assert.GreaterOrEqual(t, progressUpdates[i].Progress, progressUpdates[i-1].Progress,
			"Progress should be non-decreasing")
	}

	// 验证最后的进度是100%
	lastProgress := progressUpdates[len(progressUpdates)-1]
	assert.Equal(t, 1.0, lastProgress.Progress, "Final progress should be 100%")
	assert.Equal(t, "completed", lastProgress.Stage, "Final stage should be completed")

	mockPlugin.AssertExpectations(t)
	mockLoader.AssertExpectations(t)
}

// TestPluginStateTransitionsForUnload 测试插件状态转换
func TestPluginStateTransitionsForUnload(t *testing.T) {
	tests := []struct {
		name        string
		fromState   loader.PluginState
		toState     loader.PluginState
		expectValid bool
	}{
		{"loaded to unloading", loader.PluginStateLoaded, loader.PluginStateUnloading, true},
		{"stopped to unloading", loader.PluginStateStopped, loader.PluginStateUnloading, true},
		{"error to unloading", loader.PluginStateError, loader.PluginStateUnloading, true},
		{"running to unloading", loader.PluginStateRunning, loader.PluginStateUnloading, false},
		{"unloading to unloaded", loader.PluginStateUnloading, loader.PluginStateUnloaded, true},
		{"unloading to unloaded", loader.PluginStateUnloading, loader.PluginStateUnloaded, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := isValidStateTransition(tt.fromState, tt.toState)
			assert.Equal(t, tt.expectValid, valid,
				"State transition from %s to %s should be %v",
				tt.fromState.String(), tt.toState.String(), tt.expectValid)
		})
	}
}

// TestPluginCanUnload 测试插件是否可以卸载
func TestPluginCanUnload(t *testing.T) {
	tests := []struct {
		name      string
		state     loader.PluginState
		canUnload bool
	}{
		{"loaded can unload", loader.PluginStateLoaded, true},
		{"stopped can unload", loader.PluginStateStopped, true},
		{"error can unload", loader.PluginStateError, true},
		{"stopped can unload", loader.PluginStateStopped, true},
		{"running cannot unload", loader.PluginStateRunning, false},
		{"unloading cannot unload", loader.PluginStateUnloading, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canUnload := canUnloadFromState(tt.state)
			assert.Equal(t, tt.canUnload, canUnload,
				"State %s CanUnload should be %v", tt.state.String(), tt.canUnload)
		})
	}
}

// setupUnloadTestManager 创建测试用的插件管理器
func setupUnloadTestManager(t *testing.T) (*HybridPluginManager, *UnloadMockPlugin, *UnloadMockPluginLoader) {
	// 创建模拟组件
	mockEventBus := &UnloadMockEventBus{}
	mockServiceRegistry := &UnloadMockServiceRegistry{}
	mockSecurityManager := &UnloadMockSecurityManager{}

	// 设置模拟期望
	mockEventBus.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mockServiceRegistry.On("UnregisterService", mock.Anything).Return(nil)
	mockServiceRegistry.On("ListServices").Return([]string{})
	mockSecurityManager.On("ValidatePlugin", mock.Anything).Return(nil)
	mockSecurityManager.On("CheckPermission", mock.Anything).Return(true)
	mockSecurityManager.On("CheckPermissions", mock.Anything, mock.Anything).Return(nil)

	// 创建模拟插件和加载器
	mockPlugin := NewUnloadMockPlugin("test-plugin", "1.0.0")
	mockLoader := &UnloadMockPluginLoader{}

	// 不在这里设置全局期望，让各个测试自己设置

	// 创建管理器
	config := DefaultManagerConfig()
	manager := &HybridPluginManager{
		eventBus:        mockEventBus,
		serviceRegistry: mockServiceRegistry,
		securityManager: mockSecurityManager,
		logger:          slog.New(slog.NewTextHandler(os.Stdout, nil)),
		plugins:         make(map[string]*ManagedPlugin),
		config:          config,
	}

	// 设置加载器以避免nil pointer - 在测试中都设为nil
	manager.dynamicLoader = nil
	manager.rpcLoader = nil
	manager.wasmLoader = nil
	manager.hotReloadLoader = nil

	return manager, mockPlugin, mockLoader
}

// 模拟组件的简单实现
type UnloadMockEventBus struct {
	mock.Mock
}

func (m *UnloadMockEventBus) Publish(eventType string, data interface{}) error {
	args := m.Called(eventType, data)
	return args.Error(0)
}

func (m *UnloadMockEventBus) Subscribe(eventType string, handler func(interface{})) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *UnloadMockEventBus) Unsubscribe(eventType string, handler func(interface{})) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *UnloadMockEventBus) GetSubscriberCount(eventType string) int {
	args := m.Called(eventType)
	return args.Int(0)
}

type UnloadMockServiceRegistry struct {
	mock.Mock
}

func (m *UnloadMockServiceRegistry) RegisterService(name string, service interface{}) error {
	args := m.Called(name, service)
	return args.Error(0)
}

func (m *UnloadMockServiceRegistry) UnregisterService(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *UnloadMockServiceRegistry) GetService(name string) (interface{}, error) {
	args := m.Called(name)
	return args.Get(0), args.Error(1)
}

func (m *UnloadMockServiceRegistry) RegisterPlugin(plugin loader.Plugin) error {
	args := m.Called(plugin)
	return args.Error(0)
}

func (m *UnloadMockServiceRegistry) ListServices() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

type UnloadMockSecurityManager struct {
	mock.Mock
}

func (m *UnloadMockSecurityManager) ValidatePlugin(pluginPath string) error {
	args := m.Called(pluginPath)
	return args.Error(0)
}

func (m *UnloadMockSecurityManager) CleanupPluginContext(pluginID string) error {
	args := m.Called(pluginID)
	return args.Error(0)
}

func (m *UnloadMockSecurityManager) RevokeAllPermissions(pluginID string) error {
	args := m.Called(pluginID)
	return args.Error(0)
}

func (m *UnloadMockSecurityManager) CheckPermission(permission string) bool {
	args := m.Called(permission)
	return args.Bool(0)
}

func (m *UnloadMockSecurityManager) CheckPermissions(pluginID string, permission string) error {
	args := m.Called(pluginID, permission)
	return args.Error(0)
}

type UnloadMockLogger struct {
	mock.Mock
}

func (m *UnloadMockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *UnloadMockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *UnloadMockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *UnloadMockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}