package kernel

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBootstrap_Initialize tests bootstrap initialization
func TestBootstrap_Initialize(t *testing.T) {
	tests := []struct {
		name   string
		config *BootstrapConfig
		wantErr bool
	}{
		{
			name:   "default config",
			config: nil,
			wantErr: false,
		},
		{
			name: "custom config",
			config: &BootstrapConfig{
				ConfigPath:   "test_config.yaml",
				LogLevel:     "debug",
				LogFormat:    "json",
				StartTimeout: 10 * time.Second,
				StopTimeout:  5 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrap := NewBootstrap(tt.config)
			require.NotNil(t, bootstrap)

			err := bootstrap.Initialize()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bootstrap.GetKernel())
				assert.NotNil(t, bootstrap.GetLogger())
			}
		})
	}
}

// TestBootstrap_StartStop tests bootstrap start and stop
func TestBootstrap_StartStop(t *testing.T) {
	bootstrap := NewBootstrap(nil)
	require.NotNil(t, bootstrap)

	// Initialize
	err := bootstrap.Initialize()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start
	err = bootstrap.Start(ctx)
	assert.NoError(t, err)

	// Verify kernel is running
	kernel := bootstrap.GetKernel()
	assert.True(t, kernel.IsRunning())

	// Stop
	err = bootstrap.Stop(ctx)
	assert.NoError(t, err)

	// Verify kernel is stopped
	assert.False(t, kernel.IsRunning())
}

// TestBootstrap_Shutdown tests bootstrap shutdown
func TestBootstrap_Shutdown(t *testing.T) {
	bootstrap := NewBootstrap(nil)
	require.NotNil(t, bootstrap)

	// Initialize and start
	err := bootstrap.Initialize()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = bootstrap.Start(ctx)
	require.NoError(t, err)

	// Shutdown
	err = bootstrap.Shutdown(ctx)
	assert.NoError(t, err)

	// Verify kernel is not running
	kernel := bootstrap.GetKernel()
	assert.False(t, kernel.IsRunning())
}

// TestBootstrap_ConfigLoading tests configuration loading
func TestBootstrap_ConfigLoading(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	configContent := `
kernel:
  name: "test-kernel"
  version: "1.0.0"
  log_level: "debug"
plugins:
  enabled: true
  auto_load: false
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test with config file
	config := &BootstrapConfig{
		ConfigPath: configPath,
		LogLevel:   "info",
	}

	bootstrap := NewBootstrap(config)
	err = bootstrap.Initialize()
	assert.NoError(t, err)

	// Verify config was loaded
	kernel := bootstrap.GetKernel()
	kernelConfig := kernel.GetConfig()
	assert.Equal(t, "test-kernel", kernelConfig.String("kernel.name"))
	assert.Equal(t, "1.0.0", kernelConfig.String("kernel.version"))
	assert.Equal(t, "debug", kernelConfig.String("kernel.log_level"))
	assert.Equal(t, true, kernelConfig.Bool("plugins.enabled"))
	assert.Equal(t, false, kernelConfig.Bool("plugins.auto_load"))
}

// TestBootstrap_LogLevels tests different log levels
func TestBootstrap_LogLevels(t *testing.T) {
	logLevels := []string{"debug", "info", "warn", "error"}

	for _, level := range logLevels {
		t.Run(level, func(t *testing.T) {
			config := &BootstrapConfig{
				LogLevel: level,
			}

			bootstrap := NewBootstrap(config)
			err := bootstrap.Initialize()
			assert.NoError(t, err)

			logger := bootstrap.GetLogger()
			assert.NotNil(t, logger)
		})
	}
}

// TestBootstrap_LogFormats tests different log formats
func TestBootstrap_LogFormats(t *testing.T) {
	logFormats := []string{"text", "json"}

	for _, format := range logFormats {
		t.Run(format, func(t *testing.T) {
			config := &BootstrapConfig{
				LogFormat: format,
			}

			bootstrap := NewBootstrap(config)
			err := bootstrap.Initialize()
			assert.NoError(t, err)

			logger := bootstrap.GetLogger()
			assert.NotNil(t, logger)
		})
	}
}

// TestBootstrap_Timeouts tests timeout configurations
func TestBootstrap_Timeouts(t *testing.T) {
	config := &BootstrapConfig{
		StartTimeout: 1 * time.Second,
		StopTimeout:  500 * time.Millisecond,
	}

	bootstrap := NewBootstrap(config)
	err := bootstrap.Initialize()
	assert.NoError(t, err)

	// Test with short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start should succeed within timeout
	err = bootstrap.Start(ctx)
	assert.NoError(t, err)

	// Stop should succeed within timeout
	err = bootstrap.Stop(ctx)
	assert.NoError(t, err)
}

// TestBootstrap_EnvironmentVariables tests environment variable loading
func TestBootstrap_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("MUSICFOX_KERNEL_NAME", "env-test-kernel")
	os.Setenv("MUSICFOX_KERNEL_LOG_LEVEL", "warn")
	defer func() {
		os.Unsetenv("MUSICFOX_KERNEL_NAME")
		os.Unsetenv("MUSICFOX_KERNEL_LOG_LEVEL")
	}()

	bootstrap := NewBootstrap(nil)
	err := bootstrap.Initialize()
	assert.NoError(t, err)

	// Verify environment variables were loaded
	kernel := bootstrap.GetKernel()
	kernelConfig := kernel.GetConfig()
	assert.Equal(t, "env-test-kernel", kernelConfig.String("kernel.name"))
	assert.Equal(t, "warn", kernelConfig.String("kernel.log_level"))
}

// TestBootstrap_DefaultConfig tests default configuration values
func TestBootstrap_DefaultConfig(t *testing.T) {
	defaultConfig := DefaultBootstrapConfig()
	assert.NotNil(t, defaultConfig)
	assert.Equal(t, "config/kernel.yaml", defaultConfig.ConfigPath)
	assert.Equal(t, "info", defaultConfig.LogLevel)
	assert.Equal(t, "text", defaultConfig.LogFormat)
	assert.Equal(t, 30*time.Second, defaultConfig.StartTimeout)
	assert.Equal(t, 10*time.Second, defaultConfig.StopTimeout)
}

// TestBootstrap_ConcurrentOperations tests concurrent bootstrap operations
func TestBootstrap_ConcurrentOperations(t *testing.T) {
	bootstrap := NewBootstrap(nil)
	err := bootstrap.Initialize()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start
	err = bootstrap.Start(ctx)
	require.NoError(t, err)

	// Try to start again (should not cause issues)
	err = bootstrap.Start(ctx)
	assert.Error(t, err) // Should return error for already started

	// Stop
	err = bootstrap.Stop(ctx)
	assert.NoError(t, err)

	// Try to stop again (should not cause issues)
	err = bootstrap.Stop(ctx)
	assert.Error(t, err) // Should return error for already stopped
}

// TestBootstrap_InvalidConfig tests invalid configuration handling
func TestBootstrap_InvalidConfig(t *testing.T) {
	// Test with non-existent config file
	config := &BootstrapConfig{
		ConfigPath: "/non/existent/config.yaml",
	}

	bootstrap := NewBootstrap(config)
	err := bootstrap.Initialize()
	// Should not fail even if config file doesn't exist
	assert.NoError(t, err)
}

// BenchmarkBootstrap_Initialize benchmarks bootstrap initialization
func BenchmarkBootstrap_Initialize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bootstrap := NewBootstrap(nil)
		err := bootstrap.Initialize()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBootstrap_StartStop benchmarks bootstrap start/stop cycle
func BenchmarkBootstrap_StartStop(b *testing.B) {
	bootstrap := NewBootstrap(nil)
	err := bootstrap.Initialize()
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := bootstrap.Start(ctx)
		if err != nil {
			b.Fatal(err)
		}

		err = bootstrap.Stop(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}