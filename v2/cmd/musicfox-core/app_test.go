package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCoreApp 测试创建核心应用
func TestNewCoreApp(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		logLevel   string
		wantErr    bool
	}{
		{
			name:       "default config",
			configPath: "",
			logLevel:   "info",
			wantErr:    false,
		},
		{
			name:       "debug log level",
			configPath: "",
			logLevel:   "debug",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := NewCoreApp(tt.configPath, tt.logLevel)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, app)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app)
				assert.NotNil(t, app.kernel)
				assert.NotNil(t, app.logger)
			}
		})
	}
}

// TestCoreAppInitialization 测试核心应用初始化
func TestCoreAppInitialization(t *testing.T) {
	app, err := NewCoreApp("", "info")
	require.NoError(t, err)
	require.NotNil(t, app)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 测试初始化
	err = app.kernel.Initialize(ctx)
	assert.NoError(t, err)

	// 测试启动
	err = app.kernel.Start(ctx)
	assert.NoError(t, err)

	// 验证组件已创建
	assert.NotNil(t, app.kernel.GetEventBus())
	assert.NotNil(t, app.kernel.GetServiceRegistry())
	assert.NotNil(t, app.kernel.GetPluginManager())

	// 清理
	app.kernel.Shutdown(ctx)
}

// TestCoreAppPluginInitialization 测试插件初始化
func TestCoreAppPluginInitialization(t *testing.T) {
	app, err := NewCoreApp("", "info")
	require.NoError(t, err)
	require.NotNil(t, app)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 初始化内核
	err = app.kernel.Initialize(ctx)
	require.NoError(t, err)

	err = app.kernel.Start(ctx)
	require.NoError(t, err)

	// 获取事件总线
	app.eventBus = app.kernel.GetEventBus()
	require.NotNil(t, app.eventBus)

	// 测试插件初始化
	err = app.initializePlugins(ctx)
	assert.NoError(t, err)

	// 验证插件已创建
	assert.NotNil(t, app.audioPlugin)
	assert.NotNil(t, app.playlistPlugin)

	// 验证插件信息
	audioInfo := app.audioPlugin.GetInfo()
	assert.NotNil(t, audioInfo)
	assert.Equal(t, "Audio Processor Plugin", audioInfo.Name)

	playlistInfo := app.playlistPlugin.GetInfo()
	assert.NotNil(t, playlistInfo)
	assert.Equal(t, "Playlist Plugin", playlistInfo.Name)

	// 清理
	app.audioPlugin.Stop()
	app.playlistPlugin.Stop()
	app.kernel.Shutdown(ctx)
}

// TestCoreAppCommandHandling 测试命令处理
func TestCoreAppCommandHandling(t *testing.T) {
	app, err := NewCoreApp("", "info")
	require.NoError(t, err)
	require.NotNil(t, app)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 初始化应用
	err = app.kernel.Initialize(ctx)
	require.NoError(t, err)

	err = app.kernel.Start(ctx)
	require.NoError(t, err)

	app.eventBus = app.kernel.GetEventBus()
	err = app.initializePlugins(ctx)
	require.NoError(t, err)

	// 创建命令处理器
	app.commandHandler = NewCommandHandler(app.audioPlugin, app.playlistPlugin, app.logger)
	require.NotNil(t, app.commandHandler)

	// 测试状态命令
	err = app.commandHandler.HandleStatus(ctx)
	assert.NoError(t, err)

	// 测试音量命令
	err = app.commandHandler.HandleVolume(ctx, []string{})
	assert.NoError(t, err)

	err = app.commandHandler.HandleVolume(ctx, []string{"50"})
	assert.NoError(t, err)

	// 测试无效音量
	err = app.commandHandler.HandleVolume(ctx, []string{"invalid"})
	assert.Error(t, err)

	err = app.commandHandler.HandleVolume(ctx, []string{"150"})
	assert.Error(t, err)

	// 清理
	app.audioPlugin.Stop()
	app.playlistPlugin.Stop()
	app.kernel.Shutdown(ctx)
}

// TestCoreAppStatusMonitor 测试状态监控
func TestCoreAppStatusMonitor(t *testing.T) {
	app, err := NewCoreApp("", "info")
	require.NoError(t, err)
	require.NotNil(t, app)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 初始化应用
	err = app.kernel.Initialize(ctx)
	require.NoError(t, err)

	err = app.kernel.Start(ctx)
	require.NoError(t, err)

	app.eventBus = app.kernel.GetEventBus()
	err = app.initializePlugins(ctx)
	require.NoError(t, err)

	// 创建状态监控器
	app.statusMonitor = NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)
	require.NotNil(t, app.statusMonitor)

	// 测试启动监控器
	err = app.statusMonitor.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, app.statusMonitor.IsRunning())

	// 获取状态
	status := app.statusMonitor.GetStatus()
	assert.NotNil(t, status)
	assert.True(t, status["running"].(bool))
	assert.Greater(t, status["subscriptions"].(int), 0)

	// 停止监控器
	app.statusMonitor.Stop()
	assert.False(t, app.statusMonitor.IsRunning())

	// 清理
	app.audioPlugin.Stop()
	app.playlistPlugin.Stop()
	app.kernel.Shutdown(ctx)
}

// TestCoreAppShutdown 测试应用关闭
func TestCoreAppShutdown(t *testing.T) {
	app, err := NewCoreApp("", "info")
	require.NoError(t, err)
	require.NotNil(t, app)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 初始化应用
	err = app.kernel.Initialize(ctx)
	require.NoError(t, err)

	err = app.kernel.Start(ctx)
	require.NoError(t, err)

	app.eventBus = app.kernel.GetEventBus()
	err = app.initializePlugins(ctx)
	require.NoError(t, err)

	app.statusMonitor = NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)
	err = app.statusMonitor.Start(ctx)
	require.NoError(t, err)

	// 测试关闭
	err = app.Shutdown(ctx)
	assert.NoError(t, err)

	// 验证组件已停止
	assert.False(t, app.statusMonitor.IsRunning())
}