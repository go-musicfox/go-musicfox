package kernel

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLauncher_NewLauncher tests launcher creation
func TestLauncher_NewLauncher(t *testing.T) {
	launcher := NewLauncher()
	require.NotNil(t, launcher)

	// Verify default config
	config := launcher.GetConfig()
	assert.NotNil(t, config)
	assert.Equal(t, LaunchModeNormal, config.Mode)
	assert.Equal(t, "config/kernel.yaml", config.ConfigPath)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, "text", config.LogFormat)
	assert.True(t, config.EnableSignals)
	assert.True(t, config.AutoLoadPlugins)
}

// TestLauncher_NewLauncherWithConfig tests launcher creation with custom config
func TestLauncher_NewLauncherWithConfig(t *testing.T) {
	customConfig := &LaunchConfig{
		Mode:         LaunchModeDebug,
		LogLevel:     "debug",
		LogFormat:    "json",
		EnableSignals: false,
	}

	launcher := NewLauncherWithConfig(customConfig)
	require.NotNil(t, launcher)

	config := launcher.GetConfig()
	assert.Equal(t, LaunchModeDebug, config.Mode)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, "json", config.LogFormat)
	assert.False(t, config.EnableSignals)
}

// TestLauncher_ParseFlags tests command line flag parsing
func TestLauncher_ParseFlags(t *testing.T) {
	// Save original args and flags
	originalArgs := os.Args
	originalCommandLine := flag.CommandLine
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = originalCommandLine
	}()

	tests := []struct {
		name     string
		args     []string
		expected func(*LaunchConfig)
		wantErr  bool
	}{
		{
			name: "default flags",
			args: []string{"launcher"},
			expected: func(config *LaunchConfig) {
				assert.Equal(t, LaunchModeNormal, config.Mode)
				assert.Equal(t, "info", config.LogLevel)
			},
			wantErr: false,
		},
		{
			name: "debug mode",
			args: []string{"launcher", "--mode", "debug", "--log-level", "debug", "--verbose"},
			expected: func(config *LaunchConfig) {
				assert.Equal(t, LaunchModeDebug, config.Mode)
				assert.Equal(t, "debug", config.LogLevel)
				assert.True(t, config.Verbose)
			},
			wantErr: false,
		},
		{
			name: "daemon mode with pid file",
			args: []string{"launcher", "--mode", "daemon", "--pid-file", "/tmp/test.pid"},
			expected: func(config *LaunchConfig) {
				assert.Equal(t, LaunchModeDaemon, config.Mode)
				assert.Equal(t, "/tmp/test.pid", config.PidFile)
			},
			wantErr: false,
		},
		{
			name: "json log format",
			args: []string{"launcher", "--log-format", "json", "--log-file", "/tmp/test.log"},
			expected: func(config *LaunchConfig) {
				assert.Equal(t, "json", config.LogFormat)
				assert.Equal(t, "/tmp/test.log", config.LogFile)
			},
			wantErr: false,
		},
		{
			name:    "invalid mode",
			args:    []string{"launcher", "--mode", "invalid"},
			expected: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag.CommandLine for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			os.Args = tt.args

			launcher := NewLauncher()
			err := launcher.ParseFlags()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected != nil {
					tt.expected(launcher.GetConfig())
				}
			}
		})
	}
}

// TestLauncher_Initialize tests launcher initialization
func TestLauncher_Initialize(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	config := &LaunchConfig{
		Mode:       LaunchModeNormal,
		LogLevel:   "info",
		LogFormat:  "text",
		WorkDir:    tempDir,
		ConfigPath: filepath.Join(tempDir, "config", "kernel.yaml"),
		PidFile:    filepath.Join(tempDir, "test.pid"),
	}

	launcher := NewLauncherWithConfig(config)
	err := launcher.Initialize()
	assert.NoError(t, err)

	// Verify logger is created
	logger := launcher.GetLogger()
	assert.NotNil(t, logger)

	// Verify kernel is created
	kernel := launcher.GetKernel()
	assert.NotNil(t, kernel)

	// Verify PID file is created
	pidFileContent, err := os.ReadFile(config.PidFile)
	assert.NoError(t, err)
	assert.Contains(t, string(pidFileContent), "")

	// Verify directories are created
	configDir := filepath.Dir(config.ConfigPath)
	_, err = os.Stat(configDir)
	assert.NoError(t, err)
}

// TestLauncher_RunTestMode tests test mode execution
func TestLauncher_RunTestMode(t *testing.T) {
	config := &LaunchConfig{
		Mode:            LaunchModeTest,
		LogLevel:        "info",
		StartTimeout:    5 * time.Second,
		StopTimeout:     2 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}

	launcher := NewLauncherWithConfig(config)
	err := launcher.Initialize()
	require.NoError(t, err)

	// Run in test mode (should complete quickly)
	start := time.Now()
	err = launcher.Run()
	assert.NoError(t, err)

	// Verify it completed in reasonable time (test mode runs for ~5 seconds)
	duration := time.Since(start)
	assert.Greater(t, duration, 4*time.Second)
	assert.Less(t, duration, 10*time.Second)
}

// TestLauncher_Shutdown tests launcher shutdown
func TestLauncher_Shutdown(t *testing.T) {
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "test.pid")

	config := &LaunchConfig{
		Mode:            LaunchModeNormal,
		LogLevel:        "info",
		PidFile:         pidFile,
		StartTimeout:    5 * time.Second,
		StopTimeout:     2 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}

	launcher := NewLauncherWithConfig(config)
	err := launcher.Initialize()
	require.NoError(t, err)

	// Verify PID file exists
	_, err = os.Stat(pidFile)
	assert.NoError(t, err)

	// Shutdown
	err = launcher.Shutdown()
	assert.NoError(t, err)

	// Verify PID file is removed
	_, err = os.Stat(pidFile)
	assert.True(t, os.IsNotExist(err))
}

// TestLauncher_ValidateConfig tests config validation
func TestLauncher_ValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *LaunchConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			config:  DefaultLaunchConfig(),
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: &LaunchConfig{
				LogLevel:        "invalid",
				LogFormat:       "text",
				StartTimeout:    30 * time.Second,
				StopTimeout:     10 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name: "invalid log format",
			config: &LaunchConfig{
				LogLevel:        "info",
				LogFormat:       "invalid",
				StartTimeout:    30 * time.Second,
				StopTimeout:     10 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "invalid log format",
		},
		{
			name: "invalid start timeout",
			config: &LaunchConfig{
				LogLevel:        "info",
				LogFormat:       "text",
				StartTimeout:    -1 * time.Second,
				StopTimeout:     10 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "start timeout must be positive",
		},
		{
			name: "invalid debug port",
			config: &LaunchConfig{
				LogLevel:        "info",
				LogFormat:       "text",
				StartTimeout:    30 * time.Second,
				StopTimeout:     10 * time.Second,
				ShutdownTimeout: 30 * time.Second,
				DebugPort:       70000,
			},
			wantErr: true,
			errMsg:  "invalid debug port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			launcher := NewLauncherWithConfig(tt.config)
			
			// Use reflection to call private validateConfig method
			// For testing purposes, we'll test through ParseFlags which calls validateConfig
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
			os.Args = []string{"test"}
			
			err := launcher.ParseFlags()
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLauncher_LogLevels tests different log levels
func TestLauncher_LogLevels(t *testing.T) {
	logLevels := []string{"debug", "info", "warn", "error"}

	for _, level := range logLevels {
		t.Run(level, func(t *testing.T) {
			config := &LaunchConfig{
				Mode:            LaunchModeNormal,
				LogLevel:        level,
				LogFormat:       "text",
				StartTimeout:    5 * time.Second,
				StopTimeout:     2 * time.Second,
				ShutdownTimeout: 5 * time.Second,
			}

			launcher := NewLauncherWithConfig(config)
			err := launcher.Initialize()
			assert.NoError(t, err)

			logger := launcher.GetLogger()
			assert.NotNil(t, logger)
		})
	}
}

// TestLauncher_LogFormats tests different log formats
func TestLauncher_LogFormats(t *testing.T) {
	logFormats := []string{"text", "json"}

	for _, format := range logFormats {
		t.Run(format, func(t *testing.T) {
			config := &LaunchConfig{
				Mode:            LaunchModeNormal,
				LogLevel:        "info",
				LogFormat:       format,
				StartTimeout:    5 * time.Second,
				StopTimeout:     2 * time.Second,
				ShutdownTimeout: 5 * time.Second,
			}

			launcher := NewLauncherWithConfig(config)
			err := launcher.Initialize()
			assert.NoError(t, err)

			logger := launcher.GetLogger()
			assert.NotNil(t, logger)
		})
	}
}

// TestLauncher_LogFile tests log file output
func TestLauncher_LogFile(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &LaunchConfig{
		Mode:            LaunchModeNormal,
		LogLevel:        "info",
		LogFormat:       "text",
		LogFile:         logFile,
		StartTimeout:    5 * time.Second,
		StopTimeout:     2 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}

	launcher := NewLauncherWithConfig(config)
	err := launcher.Initialize()
	assert.NoError(t, err)

	// Verify log file is created
	_, err = os.Stat(logFile)
	assert.NoError(t, err)

	// Write a log message
	logger := launcher.GetLogger()
	logger.Info("Test log message")

	// Verify log file contains content
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test log message")
}

// TestLaunchMode_String tests launch mode string representation
func TestLaunchMode_String(t *testing.T) {
	tests := []struct {
		mode     LaunchMode
		expected string
	}{
		{LaunchModeNormal, "normal"},
		{LaunchModeDaemon, "daemon"},
		{LaunchModeDebug, "debug"},
		{LaunchModeTest, "test"},
		{LaunchMode(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.String())
		})
	}
}

// TestDefaultLaunchConfig tests default launch configuration
func TestDefaultLaunchConfig(t *testing.T) {
	config := DefaultLaunchConfig()
	assert.NotNil(t, config)

	// Verify default values
	assert.Equal(t, LaunchModeNormal, config.Mode)
	assert.Equal(t, "config/kernel.yaml", config.ConfigPath)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, "text", config.LogFormat)
	assert.Equal(t, "", config.LogFile)
	assert.Equal(t, "", config.PidFile)
	assert.Equal(t, "", config.WorkDir)
	assert.Equal(t, 30*time.Second, config.StartTimeout)
	assert.Equal(t, 10*time.Second, config.StopTimeout)
	assert.Equal(t, 30*time.Second, config.ShutdownTimeout)
	assert.True(t, config.EnableSignals)
	assert.False(t, config.EnableHotReload)
	assert.False(t, config.EnableMetrics)
	assert.False(t, config.EnableProfile)
	assert.Equal(t, []string{"plugins"}, config.PluginDirs)
	assert.True(t, config.AutoLoadPlugins)
	assert.Equal(t, 0, config.DebugPort)
	assert.Equal(t, 0, config.ProfilePort)
	assert.False(t, config.Verbose)
}

// TestLauncher_WorkingDirectory tests working directory change
func TestLauncher_WorkingDirectory(t *testing.T) {
	// Get current working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd) // Restore original working directory

	// Create temporary directory
	tempDir := t.TempDir()

	config := &LaunchConfig{
		Mode:            LaunchModeNormal,
		LogLevel:        "info",
		WorkDir:         tempDir,
		StartTimeout:    5 * time.Second,
		StopTimeout:     2 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}

	launcher := NewLauncherWithConfig(config)
	err = launcher.Initialize()
	assert.NoError(t, err)

	// Verify working directory was changed
	currentWd, err := os.Getwd()
	assert.NoError(t, err)
	
	// Resolve symlinks to handle macOS /private prefix
	expectedWd, err := filepath.EvalSymlinks(tempDir)
	assert.NoError(t, err)
	actualWd, err := filepath.EvalSymlinks(currentWd)
	assert.NoError(t, err)
	assert.Equal(t, expectedWd, actualWd)
}

// TestLauncher_DirectoryCreation tests automatic directory creation
func TestLauncher_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()

	config := &LaunchConfig{
		Mode:            LaunchModeNormal,
		LogLevel:        "info",
		ConfigPath:      filepath.Join(tempDir, "config", "subdir", "kernel.yaml"),
		LogFile:         filepath.Join(tempDir, "logs", "app.log"),
		PidFile:         filepath.Join(tempDir, "run", "app.pid"),
		PluginDirs:      []string{filepath.Join(tempDir, "plugins"), filepath.Join(tempDir, "custom")},
		StartTimeout:    5 * time.Second,
		StopTimeout:     2 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}

	launcher := NewLauncherWithConfig(config)
	err := launcher.Initialize()
	assert.NoError(t, err)

	// Verify directories were created
	dirs := []string{
		filepath.Dir(config.ConfigPath),
		filepath.Dir(config.LogFile),
		filepath.Dir(config.PidFile),
		config.PluginDirs[0],
		config.PluginDirs[1],
	}

	for _, dir := range dirs {
		_, err := os.Stat(dir)
		assert.NoError(t, err, "Directory should exist: %s", dir)
	}
}

// BenchmarkLauncher_Initialize benchmarks launcher initialization
func BenchmarkLauncher_Initialize(b *testing.B) {
	config := DefaultLaunchConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		launcher := NewLauncherWithConfig(config)
		err := launcher.Initialize()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLauncher_ParseFlags benchmarks flag parsing
func BenchmarkLauncher_ParseFlags(b *testing.B) {
	// Save original state
	originalArgs := os.Args
	originalCommandLine := flag.CommandLine
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = originalCommandLine
	}()

	args := []string{"launcher", "--mode", "debug", "--log-level", "debug", "--verbose"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = args

		launcher := NewLauncher()
		err := launcher.ParseFlags()
		if err != nil {
			b.Fatal(err)
		}
	}
}