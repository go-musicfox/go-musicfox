package kernel

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"log/slog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigManager_Initialize tests config manager initialization
func TestConfigManager_Initialize(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)
	require.NotNil(t, cm)

	ctx := context.Background()
	err := cm.Initialize(ctx)
	assert.NoError(t, err)

	// Verify default config is set
	assert.True(t, cm.Exists("kernel.name"))
	assert.True(t, cm.Exists("kernel.version"))
	assert.Equal(t, "go-musicfox", cm.Get("kernel.name"))
}

// TestConfigManager_LoadYAMLConfig tests YAML config loading
func TestConfigManager_LoadYAMLConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	// Create temporary YAML config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test.yaml")

	configContent := `
kernel:
  name: "yaml-test"
  version: "1.0.0"
  log_level: "debug"
plugins:
  enabled: true
  auto_load: false
  scan_dirs:
    - "plugins"
    - "custom_plugins"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Initialize and load config
	ctx := context.Background()
	err = cm.Initialize(ctx)
	require.NoError(t, err)

	err = cm.Load(configPath)
	assert.NoError(t, err)

	// Verify config values
	assert.Equal(t, "yaml-test", cm.Get("kernel.name"))
	assert.Equal(t, "1.0.0", cm.Get("kernel.version"))
	assert.Equal(t, "debug", cm.Get("kernel.log_level"))
	assert.Equal(t, true, cm.Get("plugins.enabled"))
	assert.Equal(t, false, cm.Get("plugins.auto_load"))
}

// TestConfigManager_LoadJSONConfig tests JSON config loading
func TestConfigManager_LoadJSONConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	// Create temporary JSON config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test.json")

	configContent := `{
  "kernel": {
    "name": "json-test",
    "version": "2.0.0",
    "log_level": "info"
  },
  "plugins": {
    "enabled": false,
    "auto_load": true
  }
}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Initialize and load config
	ctx := context.Background()
	err = cm.Initialize(ctx)
	require.NoError(t, err)

	err = cm.Load(configPath)
	assert.NoError(t, err)

	// Verify config values
	assert.Equal(t, "json-test", cm.Get("kernel.name"))
	assert.Equal(t, "2.0.0", cm.Get("kernel.version"))
	assert.Equal(t, "info", cm.Get("kernel.log_level"))
	assert.Equal(t, false, cm.Get("plugins.enabled"))
	assert.Equal(t, true, cm.Get("plugins.auto_load"))
}

// TestConfigManager_SetGet tests setting and getting config values
func TestConfigManager_SetGet(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	ctx := context.Background()
	err := cm.Initialize(ctx)
	require.NoError(t, err)

	// Test setting and getting values
	err = cm.Set("test.string", "hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello", cm.Get("test.string"))

	err = cm.Set("test.number", 42)
	assert.NoError(t, err)
	assert.Equal(t, 42, cm.Get("test.number"))

	err = cm.Set("test.boolean", true)
	assert.NoError(t, err)
	assert.Equal(t, true, cm.Get("test.boolean"))

	// Test exists
	assert.True(t, cm.Exists("test.string"))
	assert.True(t, cm.Exists("test.number"))
	assert.True(t, cm.Exists("test.boolean"))
	assert.False(t, cm.Exists("test.nonexistent"))
}

// TestConfigManager_Validation tests config validation
func TestConfigManager_Validation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	ctx := context.Background()
	err := cm.Initialize(ctx)
	require.NoError(t, err)

	// Test valid config
	err = cm.Validate()
	assert.NoError(t, err)

	// Test invalid log level
	err = cm.Set("kernel.log_level", "invalid")
	require.NoError(t, err)

	err = cm.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")

	// Fix log level
	err = cm.Set("kernel.log_level", "info")
	require.NoError(t, err)

	err = cm.Validate()
	assert.NoError(t, err)
}

// TestConfigManager_HotReload tests hot reload functionality
func TestConfigManager_HotReload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "hotreload.yaml")

	initialContent := `
kernel:
  name: "initial"
  version: "1.0.0"
`

	err := os.WriteFile(configPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Initialize and load config
	ctx := context.Background()
	err = cm.Initialize(ctx)
	require.NoError(t, err)

	err = cm.Load(configPath)
	require.NoError(t, err)

	// Verify initial config
	assert.Equal(t, "initial", cm.Get("kernel.name"))

	// Enable hot reload
	err = cm.EnableHotReload()
	assert.NoError(t, err)
	assert.True(t, cm.IsHotReloadEnabled())

	// Start config manager to enable file watching
	err = cm.Start(ctx)
	assert.NoError(t, err)

	// Update config file
	updatedContent := `
kernel:
  name: "updated"
  version: "2.0.0"
`

	err = os.WriteFile(configPath, []byte(updatedContent), 0644)
	require.NoError(t, err)

	// Wait a bit for file watcher to detect change
	time.Sleep(200 * time.Millisecond)

	// Verify config was reloaded
	assert.Equal(t, "updated", cm.Get("kernel.name"))
	assert.Equal(t, "2.0.0", cm.Get("kernel.version"))

	// Disable hot reload
	err = cm.DisableHotReload()
	assert.NoError(t, err)
	assert.False(t, cm.IsHotReloadEnabled())

	// Stop config manager
	err = cm.Stop(ctx)
	assert.NoError(t, err)
}

// TestConfigManager_ConfigChangeCallbacks tests config change callbacks
func TestConfigManager_ConfigChangeCallbacks(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	ctx := context.Background()
	err := cm.Initialize(ctx)
	require.NoError(t, err)

	// Track callback invocations
	callbackInvoked := false
	var callbackKey string
	var callbackOldValue, callbackNewValue interface{}

	// Register callback
	callback := func(key string, oldValue, newValue interface{}) error {
		callbackInvoked = true
		callbackKey = key
		callbackOldValue = oldValue
		callbackNewValue = newValue
		return nil
	}

	cm.OnConfigChanged(callback)

	// Change config value
	err = cm.Set("test.callback", "new_value")
	assert.NoError(t, err)

	// Wait for callback to be invoked
	time.Sleep(50 * time.Millisecond)

	// Verify callback was invoked
	assert.True(t, callbackInvoked)
	assert.Equal(t, "test.callback", callbackKey)
	assert.Nil(t, callbackOldValue)
	assert.Equal(t, "new_value", callbackNewValue)

	// Remove callback
	cm.RemoveConfigChangeCallback(callback)

	// Reset tracking
	callbackInvoked = false

	// Change config value again
	err = cm.Set("test.callback2", "another_value")
	assert.NoError(t, err)

	// Wait and verify callback was not invoked
	time.Sleep(50 * time.Millisecond)
	assert.False(t, callbackInvoked)
}

// TestConfigManager_EnvironmentVariables tests environment variable loading
func TestConfigManager_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("MUSICFOX_KERNEL_NAME", "env-test")
	os.Setenv("MUSICFOX_PLUGINS_ENABLED", "false")
	defer func() {
		os.Unsetenv("MUSICFOX_KERNEL_NAME")
		os.Unsetenv("MUSICFOX_PLUGINS_ENABLED")
	}()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	ctx := context.Background()
	err := cm.Initialize(ctx)
	assert.NoError(t, err)

	// Verify environment variables were loaded
	assert.Equal(t, "env-test", cm.Get("kernel.name"))
	// Note: Environment variables are loaded as strings
	assert.Equal(t, "false", cm.Get("plugins.enabled"))
}

// TestConfigManager_UnsupportedFileFormat tests unsupported file format handling
func TestConfigManager_UnsupportedFileFormat(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	// Create temporary file with unsupported extension
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test.xml")

	err := os.WriteFile(configPath, []byte("<config></config>"), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	err = cm.Initialize(ctx)
	require.NoError(t, err)

	// Try to load unsupported format
	err = cm.Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file extension")
}

// TestConfigManager_NonExistentFile tests non-existent file handling
func TestConfigManager_NonExistentFile(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	ctx := context.Background()
	err := cm.Initialize(ctx)
	require.NoError(t, err)

	// Try to load non-existent file
	err = cm.Load("/non/existent/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

// TestConfigManager_Reload tests config reload functionality
func TestConfigManager_Reload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "reload.yaml")

	initialContent := `
kernel:
  name: "reload-test"
  version: "1.0.0"
`

	err := os.WriteFile(configPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Initialize and load config
	ctx := context.Background()
	err = cm.Initialize(ctx)
	require.NoError(t, err)

	err = cm.Load(configPath)
	require.NoError(t, err)

	// Verify initial config
	assert.Equal(t, "reload-test", cm.Get("kernel.name"))
	assert.Equal(t, "1.0.0", cm.Get("kernel.version"))

	// Update config file
	updatedContent := `
kernel:
  name: "reload-test-updated"
  version: "2.0.0"
`

	err = os.WriteFile(configPath, []byte(updatedContent), 0644)
	require.NoError(t, err)

	// Reload config
	err = cm.Reload()
	assert.NoError(t, err)

	// Verify config was reloaded
	assert.Equal(t, "reload-test-updated", cm.Get("kernel.name"))
	assert.Equal(t, "2.0.0", cm.Get("kernel.version"))
}

// TestConfigManager_Lifecycle tests config manager lifecycle
func TestConfigManager_Lifecycle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	ctx := context.Background()

	// Initialize
	err := cm.Initialize(ctx)
	assert.NoError(t, err)

	// Start
	err = cm.Start(ctx)
	assert.NoError(t, err)

	// Stop
	err = cm.Stop(ctx)
	assert.NoError(t, err)

	// Shutdown
	err = cm.Shutdown(ctx)
	assert.NoError(t, err)
}

// BenchmarkConfigManager_Get benchmarks config value retrieval
func BenchmarkConfigManager_Get(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	ctx := context.Background()
	err := cm.Initialize(ctx)
	if err != nil {
		b.Fatal(err)
	}

	// Set some test values
	cm.Set("benchmark.test1", "value1")
	cm.Set("benchmark.test2", 42)
	cm.Set("benchmark.test3", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cm.Get("benchmark.test1")
		_ = cm.Get("benchmark.test2")
		_ = cm.Get("benchmark.test3")
	}
}

// BenchmarkConfigManager_Set benchmarks config value setting
func BenchmarkConfigManager_Set(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := NewConfigManager(logger)

	ctx := context.Background()
	err := cm.Initialize(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.Set("benchmark.dynamic", i)
	}
}