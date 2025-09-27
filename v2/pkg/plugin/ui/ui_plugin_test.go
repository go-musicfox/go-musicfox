package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPluginContext 模拟插件上下文
type mockPluginContext struct{}

func (m *mockPluginContext) GetContext() context.Context {
	return context.Background()
}

func (m *mockPluginContext) GetContainer() plugin.ServiceRegistry {
	return nil
}

func (m *mockPluginContext) GetEventBus() plugin.EventBus {
	return nil
}

func (m *mockPluginContext) GetServiceRegistry() plugin.ServiceRegistry {
	return nil
}

func (m *mockPluginContext) GetPluginConfig() plugin.PluginConfig {
	return nil
}

func (m *mockPluginContext) UpdateConfig(config plugin.PluginConfig) error {
	return nil
}

func (m *mockPluginContext) GetDataDir() string {
	return "/tmp/test"
}

func (m *mockPluginContext) GetTempDir() string {
	return "/tmp/test/temp"
}

func (m *mockPluginContext) GetLogger() plugin.Logger {
	return &mockLogger{}
}

// mockLogger 模拟日志器
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...interface{}) {}
func (m *mockLogger) Info(msg string, args ...interface{}) {}
func (m *mockLogger) Warn(msg string, args ...interface{}) {}
func (m *mockLogger) Error(msg string, args ...interface{}) {}

func (m *mockPluginContext) Shutdown() error {
	return nil
}

func (m *mockPluginContext) SendMessage(topic string, data interface{}) error {
	return nil
}

func (m *mockPluginContext) Subscribe(topic string, handler plugin.EventHandler) error {
	return nil
}

func (m *mockPluginContext) Unsubscribe(topic string, handler plugin.EventHandler) error {
	return nil
}

func (m *mockPluginContext) BroadcastMessage(message interface{}) error {
	return nil
}

func (m *mockPluginContext) GetResourceMonitor() *plugin.ResourceMonitor {
	return nil
}

func (m *mockPluginContext) GetSecurityManager() *plugin.SecurityManager {
	return nil
}

func (m *mockPluginContext) GetIsolationGroup() *plugin.IsolationGroup {
	return nil
}

func TestNewBaseUIExtensionPlugin(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:        "TestUIPlugin",
		Version:     "1.0.0",
		Description: "Test UI plugin",
		Author:      "Test Author",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop, UITypeWeb},
		DefaultTheme:     "light",
		DefaultLayout:    "grid",
		HotReloadEnabled: true,
		MaxComponents:    100,
		Settings:         map[string]string{"test": "value"},
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	
	assert.NotNil(t, uiPlugin)
	assert.Equal(t, info.Name, uiPlugin.GetInfo().Name)
	assert.Equal(t, info.Version, uiPlugin.GetVersion())
	assert.Equal(t, UITypeDesktop, uiPlugin.GetCurrentUIType())
	
	// 初始化插件以设置hotReloader
	ctx := &mockPluginContext{}
	err := uiPlugin.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, uiPlugin.CanHotReload())
}

func TestBaseUIExtensionPlugin_Initialize(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:    "TestUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		DefaultTheme:     "light",
		DefaultLayout:    "grid",
		HotReloadEnabled: false,
		MaxComponents:    10,
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	
	// 创建模拟的插件上下文
	ctx := &mockPluginContext{}
	
	err := uiPlugin.Initialize(ctx)
	assert.NoError(t, err)
	
	// 验证组件是否正确初始化
	assert.NotNil(t, uiPlugin.renderer)
	assert.NotNil(t, uiPlugin.inputHandler)
	assert.NotNil(t, uiPlugin.layoutManager)
	assert.NotNil(t, uiPlugin.themeManager)
	assert.Nil(t, uiPlugin.hotReloader) // 热重载未启用
}

func TestBaseUIExtensionPlugin_ComponentManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:    "TestUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		MaxComponents:    5,
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	ctx := context.Background()
	
	// 测试注册组件
	component := &UIComponent{
		ID:      "test-component",
		Name:    "Test Component",
		Type:    ComponentTypeButton,
		Version: "1.0.0",
		Visible: true,
		Enabled: true,
		Position: &Position{X: 0, Y: 0, Z: 0},
		Size:     &Size{Width: 100, Height: 50},
	}
	
	err := uiPlugin.RegisterComponent(ctx, component)
	assert.NoError(t, err)
	
	// 测试获取组件
	retrievedComponent, err := uiPlugin.GetComponent(ctx, "test-component")
	assert.NoError(t, err)
	assert.Equal(t, component.ID, retrievedComponent.ID)
	assert.Equal(t, component.Name, retrievedComponent.Name)
	
	// 测试列出组件
	components, err := uiPlugin.ListComponents(ctx)
	assert.NoError(t, err)
	assert.Len(t, components, 1)
	
	// 测试注销组件
	err = uiPlugin.UnregisterComponent(ctx, "test-component")
	assert.NoError(t, err)
	
	// 验证组件已被删除
	_, err = uiPlugin.GetComponent(ctx, "test-component")
	assert.Error(t, err)
}

func TestBaseUIExtensionPlugin_ComponentLimits(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:    "TestUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		MaxComponents:    2, // 限制为2个组件
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	ctx := context.Background()
	
	// 注册第一个组件
	component1 := &UIComponent{
		ID:   "component-1",
		Name: "Component 1",
		Type: ComponentTypeButton,
	}
	err := uiPlugin.RegisterComponent(ctx, component1)
	assert.NoError(t, err)
	
	// 注册第二个组件
	component2 := &UIComponent{
		ID:   "component-2",
		Name: "Component 2",
		Type: ComponentTypeInput,
	}
	err = uiPlugin.RegisterComponent(ctx, component2)
	assert.NoError(t, err)
	
	// 尝试注册第三个组件（应该失败）
	component3 := &UIComponent{
		ID:   "component-3",
		Name: "Component 3",
		Type: ComponentTypeLabel,
	}
	err = uiPlugin.RegisterComponent(ctx, component3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum number of components reached")
}

func TestBaseUIExtensionPlugin_ThemeManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:    "TestUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		DefaultTheme:     "light",
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	ctx := &mockPluginContext{}
	err := uiPlugin.Initialize(ctx)
	require.NoError(t, err)
	
	// 测试获取支持的主题
	themes := uiPlugin.GetSupportedThemes()
	assert.NotEmpty(t, themes)
	
	// 测试应用主题
	darkTheme := uiPlugin.createDefaultTheme("dark")
	err = uiPlugin.ApplyTheme(darkTheme)
	assert.NoError(t, err)
	
	// 验证当前主题
	currentTheme := uiPlugin.GetCurrentTheme()
	assert.NotNil(t, currentTheme)
	assert.Equal(t, "dark", currentTheme.ID)
	assert.True(t, currentTheme.IsDark)
}

func TestBaseUIExtensionPlugin_LayoutManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:    "TestUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		DefaultLayout:    "grid",
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	ctx := &mockPluginContext{}
	err := uiPlugin.Initialize(ctx)
	require.NoError(t, err)
	
	// 测试获取支持的布局
	layouts := uiPlugin.GetSupportedLayouts()
	assert.NotEmpty(t, layouts)
	
	// 测试设置布局
	flexLayout := uiPlugin.createDefaultLayout("flex")
	err = uiPlugin.SetLayout(flexLayout)
	assert.NoError(t, err)
	
	// 验证当前布局
	currentLayout := uiPlugin.GetLayout()
	assert.NotNil(t, currentLayout)
	assert.Equal(t, "flex", currentLayout.ID)
	assert.Equal(t, LayoutTypeFlex, currentLayout.Type)
}

func TestBaseUIExtensionPlugin_UITypeManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:    "TestUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop, UITypeWeb, UITypeMobile},
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	
	// 测试获取支持的UI类型
	supportedTypes := uiPlugin.GetSupportedUITypes()
	assert.Len(t, supportedTypes, 3)
	assert.Contains(t, supportedTypes, UITypeDesktop)
	assert.Contains(t, supportedTypes, UITypeWeb)
	assert.Contains(t, supportedTypes, UITypeMobile)
	
	// 测试设置支持的UI类型
	err := uiPlugin.SetUIType(UITypeWeb)
	assert.NoError(t, err)
	assert.Equal(t, UITypeWeb, uiPlugin.GetCurrentUIType())
	
	// 测试设置不支持的UI类型
	err = uiPlugin.SetUIType(UITypeTerminal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestBaseUIExtensionPlugin_EventHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:    "TestUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		MaxComponents:    10,
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	ctx := context.Background()
	
	// 创建模拟事件处理器
	mockHandler := &MockEventHandler{
		eventType: "test-event",
		priority:  10,
	}
	
	// 注册事件处理器
	err := uiPlugin.RegisterEventHandler("test-event", mockHandler)
	assert.NoError(t, err)
	
	// 创建测试事件
	event := &UIEvent{
		Type:      "test-event",
		Source:    "test",
		Target:    "component",
		Data:      map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
	}
	
	// 处理事件
	err = uiPlugin.HandleUIEvent(ctx, event)
	assert.NoError(t, err)
	assert.True(t, mockHandler.handled)
	
	// 注销事件处理器
	err = uiPlugin.UnregisterEventHandler("test-event")
	assert.NoError(t, err)
	
	// 再次处理事件（应该没有处理器）
	mockHandler.handled = false
	err = uiPlugin.HandleUIEvent(ctx, event)
	assert.NoError(t, err)
	assert.False(t, mockHandler.handled)
}

func TestBaseUIExtensionPlugin_Render(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:    "TestUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		MaxComponents:    10,
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	ctx := &mockPluginContext{}
	err := uiPlugin.Initialize(ctx)
	require.NoError(t, err)
	
	// 启动插件
	err = uiPlugin.Start()
	require.NoError(t, err)
	
	// 注册一个可见组件
	component := &UIComponent{
		ID:      "visible-component",
		Name:    "Visible Component",
		Type:    ComponentTypeButton,
		Visible: true,
	}
	err = uiPlugin.RegisterComponent(context.Background(), component)
	require.NoError(t, err)
	
	// 创建应用状态
	appState := &AppState{
		CurrentView: "main",
		Config:      map[string]string{"ui_type": "desktop"},
		UpdatedAt:   time.Now(),
	}
	
	// 测试渲染
	err = uiPlugin.Render(context.Background(), appState)
	assert.NoError(t, err)
}

func TestBaseUIExtensionPlugin_InputHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:    "TestUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		MaxComponents:    10,
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	ctx := &mockPluginContext{}
	err := uiPlugin.Initialize(ctx)
	require.NoError(t, err)
	
	// 启动插件
	err = uiPlugin.Start()
	require.NoError(t, err)
	
	// 创建输入事件
	inputEvent := &InputEvent{
		Type:      InputTypeKeyboard,
		Key:       "Enter",
		Modifiers: []string{},
		Timestamp: time.Now(),
	}
	
	// 测试输入处理
	err = uiPlugin.HandleInput(context.Background(), inputEvent)
	assert.NoError(t, err)
}

func TestBaseUIExtensionPlugin_HotReload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	info := &plugin.PluginInfo{
		Name:    "TestUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		HotReloadEnabled: true,
		MaxComponents:    10,
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	ctx := &mockPluginContext{}
	err := uiPlugin.Initialize(ctx)
	require.NoError(t, err)
	
	// 测试热重载能力
	assert.True(t, uiPlugin.CanHotReload())
	
	// 测试版本获取
	version := uiPlugin.GetVersion()
	assert.Equal(t, "1.0.0", version)
	
	// 测试热重载
	newData := []byte("new plugin data")
	err = uiPlugin.HotReload(newData)
	assert.NoError(t, err)
}

// MockEventHandler 模拟事件处理器
type MockEventHandler struct {
	eventType string
	priority  int
	handled   bool
}

func (m *MockEventHandler) Handle(ctx context.Context, event *UIEvent) error {
	m.handled = true
	return nil
}

func (m *MockEventHandler) GetType() string {
	return m.eventType
}

func (m *MockEventHandler) GetPriority() int {
	return m.priority
}

// 基准测试

func BenchmarkBaseUIExtensionPlugin_RegisterComponent(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	info := &plugin.PluginInfo{
		Name:    "BenchUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		MaxComponents:    10000,
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component := &UIComponent{
			ID:   fmt.Sprintf("component-%d", i),
			Name: fmt.Sprintf("Component %d", i),
			Type: ComponentTypeButton,
		}
		uiPlugin.RegisterComponent(ctx, component)
	}
}

func BenchmarkBaseUIExtensionPlugin_Render(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	info := &plugin.PluginInfo{
		Name:    "BenchUIPlugin",
		Version: "1.0.0",
	}
	
	config := &UIPluginConfig{
		SupportedUITypes: []UIType{UITypeDesktop},
		MaxComponents:    100,
	}
	
	uiPlugin := NewBaseUIExtensionPlugin(info, config, logger)
	ctx := &mockPluginContext{}
	uiPlugin.Initialize(ctx)
	uiPlugin.Start()
	
	// 注册一些组件
	for i := 0; i < 10; i++ {
		component := &UIComponent{
			ID:      fmt.Sprintf("component-%d", i),
			Name:    fmt.Sprintf("Component %d", i),
			Type:    ComponentTypeButton,
			Visible: true,
		}
		uiPlugin.RegisterComponent(context.Background(), component)
	}
	
	appState := &AppState{
		CurrentView: "main",
		Config:      map[string]string{"ui_type": "desktop"},
		UpdatedAt:   time.Now(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uiPlugin.Render(context.Background(), appState)
	}
}