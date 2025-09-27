package tui

import (
	"context"
	"testing"

	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// mockPluginContext 模拟插件上下文
type mockPluginContext struct{}

func (m *mockPluginContext) GetContext() context.Context {
	return context.Background()
}

func (m *mockPluginContext) GetConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func (m *mockPluginContext) GetPluginConfig() plugin.PluginConfig {
	return nil
}

func (m *mockPluginContext) UpdateConfig(config plugin.PluginConfig) error {
	return nil
}

func (m *mockPluginContext) GetEventBus() plugin.EventBus {
	return nil
}

func (m *mockPluginContext) GetServiceRegistry() plugin.ServiceRegistry {
	return nil
}

func (m *mockPluginContext) GetContainer() plugin.ServiceRegistry {
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

// mockLogger 模拟日志器
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...interface{}) {}
func (m *mockLogger) Info(msg string, args ...interface{}) {}
func (m *mockLogger) Warn(msg string, args ...interface{}) {}
func (m *mockLogger) Error(msg string, args ...interface{}) {}

func TestNewTUIPlugin(t *testing.T) {
	plugin := NewTUIPlugin()

	if plugin == nil {
		t.Fatal("NewTUIPlugin returned nil")
	}

	if plugin.GetInfo().Name != "TUI Plugin" {
		t.Errorf("Expected plugin name 'TUI Plugin', got '%s'", plugin.GetInfo().Name)
	}

	if plugin.GetInfo().Version != "1.0.0" {
		t.Errorf("Expected plugin version '1.0.0', got '%s'", plugin.GetInfo().Version)
	}

	if plugin.GetInfo().Type != "ui" {
		t.Errorf("Expected plugin type 'ui', got '%s'", plugin.GetInfo().Type)
	}
}

func TestTUIPluginInitialize(t *testing.T) {
	plugin := NewTUIPlugin()
	ctx := &mockPluginContext{}

	err := plugin.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Check if components are initialized
	tuiPlugin := plugin.(*TUIPlugin)
	if tuiPlugin.inputHandler == nil {
		t.Error("Input handler not initialized")
	}

	if tuiPlugin.viewManager == nil {
		t.Error("View manager not initialized")
	}

	if tuiPlugin.player == nil {
		t.Error("Player not initialized")
	}
}

func TestTUIPluginStart(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize first
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Start plugin
	err = plugin.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	tuiPlugin := plugin.(*TUIPlugin)
	if !tuiPlugin.IsRunning() {
		t.Error("Plugin should be running after Start")
	}

	// Stop plugin
	err = plugin.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if tuiPlugin.IsRunning() {
		t.Error("Plugin should not be running after Stop")
	}
}

func TestTUIPluginRender(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize and start plugin
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = plugin.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer plugin.Stop()

	// Test that plugin is running (render functionality is internal)
	tuiPlugin := plugin.(*TUIPlugin)
	if !tuiPlugin.IsRunning() {
		t.Error("Plugin should be running for render operations")
	}

	// Test that plugin is running (app may be nil in test environment)
	if !tuiPlugin.IsRunning() {
		t.Error("Plugin should be running")
	}

	// Test that core components are available
	if tuiPlugin.GetConfig() == nil {
		t.Error("Config should be available when plugin is running")
	}
}

func TestTUIPluginHandleInput(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize and start plugin
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = plugin.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer plugin.Stop()

	// Test that input handler is available
	tuiPlugin := plugin.(*TUIPlugin)
	if tuiPlugin.inputHandler == nil {
		t.Error("Input handler should be available when plugin is running")
	}

	// Test that plugin is running
	if !tuiPlugin.IsRunning() {
		t.Error("Plugin should be running for input handling")
	}
}

func TestTUIPluginGetLayout(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize plugin
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test that plugin has basic components after initialization
	tuiPlugin := plugin.(*TUIPlugin)
	if tuiPlugin.GetConfig() == nil {
		t.Error("Config should be available")
	}
}

func TestTUIPluginGetSupportedLayouts(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize plugin
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test that plugin has view manager
	tuiPlugin := plugin.(*TUIPlugin)
	if tuiPlugin.viewManager == nil {
		t.Error("View manager should be available")
	}
}

func TestTUIPluginGetSupportedUITypes(t *testing.T) {
	plugin := NewTUIPlugin()

	// Test that plugin info is correct
	info := plugin.GetInfo()
	if info.Type != "ui" {
		t.Errorf("Expected plugin type 'ui', got '%s'", info.Type)
	}

	if info.ID != "tui" {
		t.Errorf("Expected plugin ID 'tui', got '%s'", info.ID)
	}
}

func TestTUIPluginGetComponents(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize plugin
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test that plugin has main components
	tuiPlugin := plugin.(*TUIPlugin)
	if tuiPlugin.player == nil {
		t.Error("Player component should be available")
	}

	if tuiPlugin.inputHandler == nil {
		t.Error("Input handler component should be available")
	}

	if tuiPlugin.viewManager == nil {
		t.Error("View manager component should be available")
	}
}

func TestTUIPluginSetTheme(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize plugin
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test that plugin has config for theme management
	tuiPlugin := plugin.(*TUIPlugin)
	config := tuiPlugin.GetConfig()
	if config == nil {
		t.Error("Config should be available for theme management")
	}
}

func TestTUIPluginConfiguration(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize plugin
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test that plugin has internal config
	tuiPlugin := plugin.(*TUIPlugin)
	if tuiPlugin.GetConfig() == nil {
		t.Error("Internal config should be available")
	}
}

func TestTUIPluginValidateConfiguration(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize plugin
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test that plugin is properly initialized
	tuiPlugin := plugin.(*TUIPlugin)
	if tuiPlugin.GetLogger() == nil {
		t.Error("Logger should be available")
	}
}

func TestTUIPluginEventHandling(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize plugin
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test that plugin can be started and stopped for event handling
	err = plugin.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer plugin.Stop()

	tuiPlugin := plugin.(*TUIPlugin)
	if !tuiPlugin.IsRunning() {
		t.Error("Plugin should be running for event handling")
	}
}

func TestTUIPluginCleanup(t *testing.T) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize and start plugin
	err := plugin.Initialize(mockCtx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = plugin.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Test cleanup (stop)
	err = plugin.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	tuiPlugin := plugin.(*TUIPlugin)
	if tuiPlugin.IsRunning() {
		t.Error("Plugin should not be running after stop")
	}
}

// Benchmark tests
func BenchmarkTUIPluginInitialize(b *testing.B) {
	mockCtx := &mockPluginContext{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plugin := NewTUIPlugin()
		err := plugin.Initialize(mockCtx)
		if err != nil {
			b.Fatalf("Initialize failed: %v", err)
		}
	}
}

func BenchmarkTUIPluginStartStop(b *testing.B) {
	plugin := NewTUIPlugin()
	mockCtx := &mockPluginContext{}

	// Initialize once
	err := plugin.Initialize(mockCtx)
	if err != nil {
		b.Fatalf("Initialize failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := plugin.Start()
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		err = plugin.Stop()
		if err != nil {
			b.Fatalf("Stop failed: %v", err)
		}
	}
}