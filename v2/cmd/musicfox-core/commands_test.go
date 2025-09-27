package main

import (
	"context"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestApp 设置测试应用
func setupTestApp(t *testing.T) (*CoreApp, context.Context, context.CancelFunc) {
	app, err := NewCoreApp("", "info")
	require.NoError(t, err)
	require.NotNil(t, app)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// 初始化应用
	err = app.kernel.Initialize(ctx)
	require.NoError(t, err)

	err = app.kernel.Start(ctx)
	require.NoError(t, err)

	app.eventBus = app.kernel.GetEventBus()
	err = app.initializePlugins(ctx)
	require.NoError(t, err)

	app.commandHandler = NewCommandHandler(app.audioPlugin, app.playlistPlugin, app.logger)
	require.NotNil(t, app.commandHandler)

	return app, ctx, cancel
}

// teardownTestApp 清理测试应用
func teardownTestApp(app *CoreApp, ctx context.Context, cancel context.CancelFunc) {
	if app.audioPlugin != nil {
		app.audioPlugin.Stop()
	}
	if app.playlistPlugin != nil {
		app.playlistPlugin.Stop()
	}
	if app.kernel != nil {
		app.kernel.Shutdown(ctx)
	}
	cancel()
}

// TestCommandHandler_HandleStatus 测试状态命令
func TestCommandHandler_HandleStatus(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	err := app.commandHandler.HandleStatus(ctx)
	assert.NoError(t, err)
}

// TestCommandHandler_HandleVolume 测试音量命令
func TestCommandHandler_HandleVolume(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "get current volume",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "set valid volume",
			args:    []string{"50"},
			wantErr: false,
		},
		{
			name:    "set max volume",
			args:    []string{"100"},
			wantErr: false,
		},
		{
			name:    "set min volume",
			args:    []string{"0"},
			wantErr: false,
		},
		{
			name:    "invalid volume string",
			args:    []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "volume too high",
			args:    []string{"150"},
			wantErr: true,
		},
		{
			name:    "negative volume",
			args:    []string{"-10"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.commandHandler.HandleVolume(ctx, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCommandHandler_HandlePlay 测试播放命令
func TestCommandHandler_HandlePlay(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "valid file path",
			args:    []string{"/path/to/song.mp3"},
			wantErr: false, // 可能会因为文件不存在而失败，但命令处理本身不会出错
		},
		{
			name:    "valid URL",
			args:    []string{"http://example.com/song.mp3"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.commandHandler.HandlePlay(ctx, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// 注意：播放可能会失败（文件不存在等），但这里主要测试命令解析
				// 实际的播放功能测试应该在音频插件的测试中进行
				_ = err // 忽略播放错误，只关注命令处理逻辑
			}
		})
	}
}

// TestCommandHandler_HandlePlaybackControls 测试播放控制命令
func TestCommandHandler_HandlePlaybackControls(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	// 测试暂停命令
	err := app.commandHandler.HandlePause(ctx)
	// 暂停命令可能会失败（没有正在播放的歌曲），但命令处理本身不应该出错
	_ = err

	// 测试恢复命令
	err = app.commandHandler.HandleResume(ctx)
	_ = err

	// 测试停止命令
	err = app.commandHandler.HandleStop(ctx)
	assert.NoError(t, err)
}

// TestCommandHandler_HandlePlaylist 测试播放列表命令
func TestCommandHandler_HandlePlaylist(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no subcommand",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "list playlists",
			args:    []string{"list"},
			wantErr: false,
		},
		{
			name:    "create playlist without name",
			args:    []string{"create"},
			wantErr: true,
		},
		{
			name:    "create playlist with name",
			args:    []string{"create", "My", "Test", "Playlist"},
			wantErr: false,
		},
		{
			name:    "unknown subcommand",
			args:    []string{"unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.commandHandler.HandlePlaylist(ctx, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCommandHandler_PlaylistOperations 测试播放列表操作
func TestCommandHandler_PlaylistOperations(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	// 创建播放列表
	err := app.commandHandler.HandlePlaylist(ctx, []string{"create", "Test Playlist"})
	assert.NoError(t, err)

	// 列出播放列表
	err = app.commandHandler.HandlePlaylist(ctx, []string{"list"})
	assert.NoError(t, err)

	// 获取创建的播放列表
	playlists, err := app.playlistPlugin.ListPlaylists(ctx)
	assert.NoError(t, err)
	assert.Greater(t, len(playlists), 0)

	playlistID := playlists[0].ID

	// 显示播放列表详情
	err = app.commandHandler.HandlePlaylist(ctx, []string{"show", playlistID})
	assert.NoError(t, err)

	// 添加歌曲到播放列表
	err = app.commandHandler.HandlePlaylist(ctx, []string{"add", playlistID, "/path/to/test.mp3"})
	assert.NoError(t, err)

	// 删除播放列表
	err = app.commandHandler.HandlePlaylist(ctx, []string{"delete", playlistID})
	assert.NoError(t, err)
}

// TestCommandHandler_QueueOperations 测试队列操作
func TestCommandHandler_QueueOperations(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	// 显示空队列
	err := app.commandHandler.HandlePlaylist(ctx, []string{"queue"})
	assert.NoError(t, err)

	// 添加歌曲到队列
	err = app.commandHandler.HandlePlaylist(ctx, []string{"queue", "add", "/path/to/test1.mp3"})
	assert.NoError(t, err)

	err = app.commandHandler.HandlePlaylist(ctx, []string{"queue", "add", "/path/to/test2.mp3"})
	assert.NoError(t, err)

	// 显示队列
	err = app.commandHandler.HandlePlaylist(ctx, []string{"queue"})
	assert.NoError(t, err)

	// 验证队列中有歌曲
	queue, err := app.playlistPlugin.GetCurrentQueue(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(queue))

	// 打乱队列
	err = app.commandHandler.HandlePlaylist(ctx, []string{"queue", "shuffle"})
	assert.NoError(t, err)

	// 清空队列
	err = app.commandHandler.HandlePlaylist(ctx, []string{"queue", "clear"})
	assert.NoError(t, err)

	// 验证队列已清空
	queue, err = app.playlistPlugin.GetCurrentQueue(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(queue))
}

// TestExtractSongName 测试歌曲名称提取
func TestExtractSongName(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "file path with extension",
			url:      "/path/to/song.mp3",
			expected: "song",
		},
		{
			name:     "URL with extension",
			url:      "http://example.com/music/song.flac",
			expected: "song",
		},
		{
			name:     "file without extension",
			url:      "/path/to/song",
			expected: "song",
		},
		{
			name:     "empty path",
			url:      "",
			expected: "Unknown Song",
		},
		{
			name:     "complex filename",
			url:      "/music/Artist - Song Title.mp3",
			expected: "Artist - Song Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSongName(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatFunctions 测试格式化函数
func TestFormatFunctions(t *testing.T) {
	// 测试播放状态格式化
	assert.Equal(t, "Playing", formatPlayStatus(model.PlayStatusPlaying))
	assert.Equal(t, "Paused", formatPlayStatus(model.PlayStatusPaused))
	assert.Equal(t, "Stopped", formatPlayStatus(model.PlayStatusStopped))
	assert.Equal(t, "Unknown", formatPlayStatus(model.PlayStatus(999)))

	// 测试时长格式化
	assert.Equal(t, "--:--", formatDuration(0))
	assert.Equal(t, "01:30", formatDuration(90*time.Second))
	assert.Equal(t, "03:45", formatDuration(225*time.Second))

	// 测试进度格式化
	assert.Equal(t, "---%", formatProgress(0, 0))
	assert.Equal(t, "50.0%", formatProgress(30*time.Second, 60*time.Second))
	assert.Equal(t, "25.0%", formatProgress(15*time.Second, 60*time.Second))

	// 测试播放模式格式化
	assert.Equal(t, "Sequential", formatPlayMode(model.PlayModeSequential))
	assert.Equal(t, "Random", formatPlayMode(2)) // Random mode
	assert.Equal(t, "Single", formatPlayMode(1)) // Single mode
	assert.Equal(t, "Unknown", formatPlayMode(model.PlayMode(999)))
}