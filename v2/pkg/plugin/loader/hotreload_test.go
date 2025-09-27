package loader

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockEventBus is a mock implementation of EventBus
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(event string, data interface{}) error {
	args := m.Called(event, data)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(event string, handler func(interface{})) error {
	args := m.Called(event, handler)
	return args.Error(0)
}

func (m *MockEventBus) Unsubscribe(event string, handler func(interface{})) error {
	args := m.Called(event, handler)
	return args.Error(0)
}

// MockPlugin is a mock implementation of Plugin
type MockPlugin struct {
	mock.Mock
	info *PluginInfo
}

func (m *MockPlugin) GetInfo() *PluginInfo {
	if m.info == nil {
		m.info = &PluginInfo{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test Plugin",
			Author:      "Test Author",
			LoadTime:    time.Now(),
		}
	}
	return m.info
}

func (m *MockPlugin) GetCapabilities() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{"test"}
	}
	return args.Get(0).([]string)
}

func (m *MockPlugin) GetDependencies() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockPlugin) Initialize(ctx PluginContext) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPlugin) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPlugin) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPlugin) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPlugin) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPlugin) GetMetrics() (*PluginMetrics, error) {
	args := m.Called()
	return args.Get(0).(*PluginMetrics), args.Error(1)
}

// Test helper functions

func createTestPluginFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	pluginPath := filepath.Join(tempDir, "test-plugin.js")
	err := os.WriteFile(pluginPath, []byte(content), 0644)
	require.NoError(t, err)
	return pluginPath
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func TestNewHotReloadPluginLoader(t *testing.T) {
	mockEventBus := &MockEventBus{}
	logger := createTestLogger()

	loader := NewHotReloadPluginLoader(mockEventBus, logger)

	assert.NotNil(t, loader)
	assert.Equal(t, mockEventBus, loader.eventBus)
	assert.Equal(t, logger, loader.logger)
	assert.NotNil(t, loader.plugins)
	assert.NotNil(t, loader.ctx)
	assert.NotNil(t, loader.cancel)
	assert.NotNil(t, loader.stopChan)

	// Test default configuration
	config := loader.GetConfig()
	assert.Equal(t, time.Second*2, config.WatchInterval)
	assert.Equal(t, 10, config.MaxVersions)
	assert.True(t, config.AutoReload)
	assert.True(t, config.BackupEnabled)
	assert.True(t, config.StatePreservation)

	// Cleanup
	loader.Cleanup()
}

func TestHotReloadPluginLoader_LoadPlugin(t *testing.T) {
	mockEventBus := &MockEventBus{}
	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Create test plugin file
	pluginContent := `
		// Test JavaScript plugin
		function initialize() {
			console.log('Plugin initialized');
		}
	`
	pluginPath := createTestPluginFile(t, pluginContent)

	// Test loading plugin
	ctx := context.Background()
	plugin, err := loader.LoadPlugin(ctx, pluginPath)

	assert.NoError(t, err)
	assert.NotNil(t, plugin)

	// Verify plugin is loaded
	pluginID := loader.getPluginID(pluginPath)
	assert.True(t, loader.IsPluginLoaded(pluginID))

	// Verify plugin info
	info, err := loader.GetPluginInfo(pluginID)
	assert.NoError(t, err)
	assert.NotNil(t, info)

	// Test loading same plugin again (should fail)
	_, err = loader.LoadPlugin(ctx, pluginPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin already loaded")
}

func TestHotReloadPluginLoader_UnloadPlugin(t *testing.T) {
	mockEventBus := &MockEventBus{}
	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Create and load test plugin
	pluginContent := `console.log('test plugin');`
	pluginPath := createTestPluginFile(t, pluginContent)
	ctx := context.Background()

	_, err := loader.LoadPlugin(ctx, pluginPath)
	require.NoError(t, err)

	pluginID := loader.getPluginID(pluginPath)
	assert.True(t, loader.IsPluginLoaded(pluginID))

	// Test unloading plugin
	err = loader.UnloadPlugin(ctx, pluginID)
	assert.NoError(t, err)
	assert.False(t, loader.IsPluginLoaded(pluginID))

	// Test unloading non-existent plugin
	err = loader.UnloadPlugin(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestHotReloadPluginLoader_GetLoadedPlugins(t *testing.T) {
	mockEventBus := &MockEventBus{}
	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Initially no plugins loaded
	plugins := loader.GetLoadedPlugins()
	assert.Empty(t, plugins)

	// Load multiple plugins with unique names
	ctx := context.Background()
	pluginPaths := make([]string, 3)
	for i := 0; i < 3; i++ {
		content := `console.log('test plugin ` + string(rune('0'+i)) + `');`
		// Create unique plugin files
		tempDir := t.TempDir()
		pluginPath := filepath.Join(tempDir, "test-plugin-"+string(rune('0'+i))+".js")
		err := os.WriteFile(pluginPath, []byte(content), 0644)
		require.NoError(t, err)
		pluginPaths[i] = pluginPath
		_, err = loader.LoadPlugin(ctx, pluginPaths[i])
		require.NoError(t, err)
	}

	// Verify all plugins are loaded
	plugins = loader.GetLoadedPlugins()
	assert.Len(t, plugins, 3)
}

func TestHotReloadPluginLoader_ValidatePlugin(t *testing.T) {
	mockEventBus := &MockEventBus{}
	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Test valid JavaScript plugin
	validContent := `console.log('valid plugin');`
	validPath := createTestPluginFile(t, validContent)
	err := loader.ValidatePlugin(validPath)
	assert.NoError(t, err)

	// Test non-existent file
	err = loader.ValidatePlugin("/non/existent/path.js")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")

	// Test unsupported extension
	unsupportedPath := createTestPluginFile(t, "content")
	unsupportedPath = unsupportedPath[:len(unsupportedPath)-3] + ".txt"
	os.WriteFile(unsupportedPath, []byte("content"), 0644)
	err = loader.ValidatePlugin(unsupportedPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported plugin file extension")

	// Test empty file
	emptyPath := createTestPluginFile(t, "")
	err = loader.ValidatePlugin(emptyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin file is empty")
}

func TestHotReloadPluginLoader_GetLoaderType(t *testing.T) {
	mockEventBus := &MockEventBus{}
	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	loaderType := loader.GetLoaderType()
	assert.Equal(t, PluginTypeHotReload, loaderType)
}

func TestHotReloadPluginLoader_HotReload(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockEventBus.On("Publish", EventPluginHotReloaded, mock.Anything).Return(nil)

	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Load initial plugin
	initialContent := `console.log('initial version');`
	pluginPath := createTestPluginFile(t, initialContent)
	ctx := context.Background()

	_, err := loader.LoadPlugin(ctx, pluginPath)
	require.NoError(t, err)

	pluginID := loader.getPluginID(pluginPath)

	// Get initial version history
	initialVersions := loader.GetVersionHistory(pluginID)
	assert.Len(t, initialVersions, 1)

	// Perform hot reload with new content
	newContent := []byte(`console.log('updated version');`)
	err = loader.HotReload(pluginID, newContent)
	assert.NoError(t, err)

	// Verify version history updated
	updatedVersions := loader.GetVersionHistory(pluginID)
	assert.Len(t, updatedVersions, 2)

	// Verify event was published
	mockEventBus.AssertExpectations(t)
}

func TestHotReloadPluginLoader_AutoReload(t *testing.T) {
	mockEventBus := &MockEventBus{}
	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Load plugin
	initialContent := `console.log('initial');`
	pluginPath := createTestPluginFile(t, initialContent)
	ctx := context.Background()

	_, err := loader.LoadPlugin(ctx, pluginPath)
	require.NoError(t, err)

	pluginID := loader.getPluginID(pluginPath)

	// Test enabling auto-reload
	err = loader.EnableAutoReload(pluginID)
	assert.NoError(t, err)

	// Verify auto-reload is enabled
	stats, err := loader.GetPluginStats(pluginID)
	require.NoError(t, err)
	assert.True(t, stats.AutoReload)

	// Test disabling auto-reload
	err = loader.DisableAutoReload(pluginID)
	assert.NoError(t, err)

	// Verify auto-reload is disabled
	stats, err = loader.GetPluginStats(pluginID)
	require.NoError(t, err)
	assert.False(t, stats.AutoReload)

	// Test with non-existent plugin
	err = loader.EnableAutoReload("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestHotReloadPluginLoader_VersionManagement(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockEventBus.On("Publish", mock.Anything, mock.Anything).Return(nil)

	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Load plugin
	initialContent := `console.log('v1');`
	pluginPath := createTestPluginFile(t, initialContent)
	ctx := context.Background()

	_, err := loader.LoadPlugin(ctx, pluginPath)
	require.NoError(t, err)

	pluginID := loader.getPluginID(pluginPath)

	// Create multiple versions through hot reload
	for i := 2; i <= 5; i++ {
		newContent := []byte(`console.log('v` + string(rune('0'+i)) + `');`)
		err = loader.HotReload(pluginID, newContent)
		require.NoError(t, err)
	}

	// Verify version history
	versions := loader.GetVersionHistory(pluginID)
	assert.Len(t, versions, 5)

	// Test rollback to specific version
	err = loader.RollbackToVersion(pluginID, "1.2.0")
	assert.NoError(t, err)

	// Verify current version changed
	stats, err := loader.GetPluginStats(pluginID)
	require.NoError(t, err)
	assert.Equal(t, "1.2.0", stats.CurrentVersion)

	// Test rollback to non-existent version
	err = loader.RollbackToVersion(pluginID, "999.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version not found")

	// Test rollback for non-existent plugin
	err = loader.RollbackToVersion("non-existent", "1.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestHotReloadPluginLoader_Configuration(t *testing.T) {
	mockEventBus := &MockEventBus{}
	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Test default configuration
	defaultConfig := loader.GetConfig()
	assert.Equal(t, time.Second*2, defaultConfig.WatchInterval)
	assert.Equal(t, 10, defaultConfig.MaxVersions)
	assert.True(t, defaultConfig.AutoReload)

	// Test updating configuration
	newConfig := HotReloadConfig{
		WatchInterval:     time.Second * 5,
		MaxVersions:       20,
		AutoReload:        false,
		BackupEnabled:     false,
		StatePreservation: false,
	}

	loader.SetConfig(newConfig)
	updatedConfig := loader.GetConfig()
	assert.Equal(t, newConfig, updatedConfig)
}

func TestHotReloadPluginLoader_PluginStats(t *testing.T) {
	mockEventBus := &MockEventBus{}
	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Load plugin
	content := `console.log('test');`
	pluginPath := createTestPluginFile(t, content)
	ctx := context.Background()

	loadTime := time.Now()
	_, err := loader.LoadPlugin(ctx, pluginPath)
	require.NoError(t, err)

	pluginID := loader.getPluginID(pluginPath)

	// Get plugin stats
	stats, err := loader.GetPluginStats(pluginID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, pluginID, stats.PluginID)
	assert.Equal(t, pluginPath, stats.Path)
	assert.Equal(t, "1.0.0", stats.CurrentVersion)
	assert.Equal(t, 1, stats.VersionCount)
	assert.True(t, stats.LoadTime.After(loadTime.Add(-time.Second)))
	assert.True(t, stats.AutoReload)

	// Test with non-existent plugin
	_, err = loader.GetPluginStats("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestHotReloadPluginLoader_ReloadPlugin(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockEventBus.On("Publish", mock.Anything, mock.Anything).Return(nil)

	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Load plugin
	initialContent := `console.log('initial');`
	pluginPath := createTestPluginFile(t, initialContent)
	ctx := context.Background()

	_, err := loader.LoadPlugin(ctx, pluginPath)
	require.NoError(t, err)

	pluginID := loader.getPluginID(pluginPath)

	// Update plugin file
	updatedContent := `console.log('updated');`
	err = os.WriteFile(pluginPath, []byte(updatedContent), 0644)
	require.NoError(t, err)

	// Reload plugin
	err = loader.ReloadPlugin(ctx, pluginID)
	assert.NoError(t, err)

	// Verify version history updated
	versions := loader.GetVersionHistory(pluginID)
	assert.Len(t, versions, 2)

	// Test reloading unchanged plugin (should skip)
	err = loader.ReloadPlugin(ctx, pluginID)
	assert.NoError(t, err) // Should not error, but should skip reload

	// Test reloading non-existent plugin
	err = loader.ReloadPlugin(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestHotReloadPluginLoader_Cleanup(t *testing.T) {
	mockEventBus := &MockEventBus{}
	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)

	// Load some plugins with unique names
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		content := `console.log('test ` + string(rune('0'+i)) + `');`
		// Create unique plugin files
		tempDir := t.TempDir()
		pluginPath := filepath.Join(tempDir, "cleanup-test-plugin-"+string(rune('0'+i))+".js")
		err := os.WriteFile(pluginPath, []byte(content), 0644)
		require.NoError(t, err)
		_, err = loader.LoadPlugin(ctx, pluginPath)
		require.NoError(t, err)
	}

	// Verify plugins are loaded
	plugins := loader.GetLoadedPlugins()
	assert.Len(t, plugins, 3)

	// Cleanup
	err := loader.Cleanup()
	assert.NoError(t, err)

	// Verify all plugins are unloaded
	plugins = loader.GetLoadedPlugins()
	assert.Empty(t, plugins)
}

func TestHotReloadPluginLoader_ConcurrentOperations(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockEventBus.On("Publish", mock.Anything, mock.Anything).Return(nil)

	logger := createTestLogger()
	loader := NewHotReloadPluginLoader(mockEventBus, logger)
	defer loader.Cleanup()

	// Load plugin
	content := `console.log('test');`
	pluginPath := createTestPluginFile(t, content)
	ctx := context.Background()

	_, err := loader.LoadPlugin(ctx, pluginPath)
	require.NoError(t, err)

	pluginID := loader.getPluginID(pluginPath)

	// Perform concurrent operations
	done := make(chan bool, 10)

	// Concurrent hot reloads
	for i := 0; i < 5; i++ {
		go func(version int) {
			newContent := []byte(`console.log('version ` + string(rune('0'+version)) + `');`)
			err := loader.HotReload(pluginID, newContent)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Concurrent stats queries
	for i := 0; i < 5; i++ {
		go func() {
			_, err := loader.GetPluginStats(pluginID)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all operations to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Operation completed
		case <-time.After(time.Second * 5):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	// Verify final state
	versions := loader.GetVersionHistory(pluginID)
	assert.True(t, len(versions) >= 1) // At least initial version should exist
}