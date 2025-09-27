package main

import (
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStatusMonitor 测试创建状态监控器
func TestNewStatusMonitor(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	monitor := NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)
	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.audioPlugin)
	assert.NotNil(t, monitor.playlistPlugin)
	assert.NotNil(t, monitor.eventBus)
	assert.NotNil(t, monitor.logger)
	assert.False(t, monitor.running)
	assert.Equal(t, 0, len(monitor.subscriptions))
}

// TestStatusMonitor_StartStop 测试状态监控器启动和停止
func TestStatusMonitor_StartStop(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	monitor := NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)
	require.NotNil(t, monitor)

	// 测试启动
	err := monitor.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, monitor.IsRunning())
	assert.Greater(t, len(monitor.subscriptions), 0)

	// 测试重复启动
	err = monitor.Start(ctx)
	assert.Error(t, err)

	// 测试停止
	monitor.Stop()
	assert.False(t, monitor.IsRunning())
	assert.Equal(t, 0, len(monitor.subscriptions))

	// 测试重复停止
	monitor.Stop() // 应该不会出错
	assert.False(t, monitor.IsRunning())
}

// TestStatusMonitor_GetStatus 测试获取状态
func TestStatusMonitor_GetStatus(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	monitor := NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)
	require.NotNil(t, monitor)

	// 测试未启动时的状态
	status := monitor.GetStatus()
	assert.NotNil(t, status)
	assert.False(t, status["running"].(bool))
	assert.Equal(t, 0, status["subscriptions"].(int))

	// 启动监控器
	err := monitor.Start(ctx)
	require.NoError(t, err)

	// 测试启动后的状态
	status = monitor.GetStatus()
	assert.NotNil(t, status)
	assert.True(t, status["running"].(bool))
	assert.Greater(t, status["subscriptions"].(int), 0)

	// 验证音频和播放列表状态
	assert.Contains(t, status, "audio")
	assert.Contains(t, status, "playlist")

	audioStatus := status["audio"].(map[string]interface{})
	assert.Contains(t, audioStatus, "status")
	assert.Contains(t, audioStatus, "volume")

	playlistStatus := status["playlist"].(map[string]interface{})
	assert.Contains(t, playlistStatus, "queue_size")
	assert.Contains(t, playlistStatus, "play_mode")

	monitor.Stop()
}

// TestStatusMonitor_EventHandling 测试事件处理
func TestStatusMonitor_EventHandling(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	monitor := NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)
	require.NotNil(t, monitor)

	// 启动监控器
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop()

	// 测试播放开始事件处理
	playStartEvent := &event.BaseEvent{
		ID:        "test-play-start",
		Type:      "audio.play.start",
		Source:    "test",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"song_id": "test-song-123",
		},
	}

	err = monitor.handlePlayStartEvent(ctx, playStartEvent)
	assert.NoError(t, err)

	// 测试播放暂停事件处理
	playPauseEvent := &event.BaseEvent{
		ID:        "test-play-pause",
		Type:      "audio.play.pause",
		Source:    "test",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	err = monitor.handlePlayPauseEvent(ctx, playPauseEvent)
	assert.NoError(t, err)

	// 测试音量变化事件处理
	volumeChangeEvent := &event.BaseEvent{
		ID:        "test-volume-change",
		Type:      "audio.volume.change",
		Source:    "test",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"volume": 75,
		},
	}

	err = monitor.handleVolumeChangeEvent(ctx, volumeChangeEvent)
	assert.NoError(t, err)

	// 测试播放错误事件处理
	playErrorEvent := &event.BaseEvent{
		ID:        "test-play-error",
		Type:      "audio.play.error",
		Source:    "test",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"error": "Failed to load audio file",
		},
	}

	err = monitor.handlePlayErrorEvent(ctx, playErrorEvent)
	assert.NoError(t, err)
}

// TestStatusMonitor_PlaylistEventHandling 测试播放列表事件处理
func TestStatusMonitor_PlaylistEventHandling(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	monitor := NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)
	require.NotNil(t, monitor)

	// 启动监控器
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop()

	// 测试播放列表创建事件处理
	playlistCreateEvent := &event.BaseEvent{
		ID:        "test-playlist-create",
		Type:      "playlist.created",
		Source:    "test",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"playlist_name": "My Test Playlist",
			"playlist_id":   "playlist-123",
		},
	}

	err = monitor.handlePlaylistCreateEvent(ctx, playlistCreateEvent)
	assert.NoError(t, err)

	// 测试播放列表更新事件处理
	playlistUpdateEvent := &event.BaseEvent{
		ID:        "test-playlist-update",
		Type:      "playlist.updated",
		Source:    "test",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"playlist_id": "playlist-123",
			"action":      "song_added",
		},
	}

	err = monitor.handlePlaylistUpdateEvent(ctx, playlistUpdateEvent)
	assert.NoError(t, err)

	// 测试队列变化事件处理
	queueChangeEvent := &event.BaseEvent{
		ID:        "test-queue-change",
		Type:      "queue.changed",
		Source:    "test",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"queue_size": 5,
			"action":     "song_added",
		},
	}

	err = monitor.handleQueueChangeEvent(ctx, queueChangeEvent)
	assert.NoError(t, err)

	// 测试播放模式变化事件处理
	playModeChangeEvent := &event.BaseEvent{
		ID:        "test-playmode-change",
		Type:      "playmode.changed",
		Source:    "test",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"play_mode": "random",
		},
	}

	err = monitor.handlePlayModeChangeEvent(ctx, playModeChangeEvent)
	assert.NoError(t, err)
}

// TestStatusMonitor_EventSubscription 测试事件订阅
func TestStatusMonitor_EventSubscription(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	monitor := NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)
	require.NotNil(t, monitor)

	// 启动监控器
	err := monitor.Start(ctx)
	require.NoError(t, err)

	// 验证已订阅事件
	assert.Greater(t, len(monitor.subscriptions), 0)

	// 发布测试事件并验证处理
	testEvent := &event.BaseEvent{
		ID:        "test-event",
		Type:      "audio.play.start",
		Source:    "test",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"test": "data",
		},
	}

	// 发布事件
	err = app.eventBus.Publish(ctx, testEvent)
	assert.NoError(t, err)

	// 等待事件处理
	time.Sleep(100 * time.Millisecond)

	// 停止监控器
	monitor.Stop()
	assert.Equal(t, 0, len(monitor.subscriptions))
}

// TestStatusMonitor_PositionUpdate 测试位置更新
func TestStatusMonitor_PositionUpdate(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	monitor := NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)
	require.NotNil(t, monitor)

	// 启动监控器
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop()

	// 测试位置更新（当没有当前歌曲时）
	monitor.updatePlaybackPosition()

	// 这里主要测试函数不会崩溃
	// 实际的位置更新测试需要有正在播放的歌曲
}

// TestStatusMonitor_ConcurrentAccess 测试并发访问
func TestStatusMonitor_ConcurrentAccess(t *testing.T) {
	app, ctx, cancel := setupTestApp(t)
	defer teardownTestApp(app, ctx, cancel)

	monitor := NewStatusMonitor(app.audioPlugin, app.playlistPlugin, app.eventBus, app.logger)
	require.NotNil(t, monitor)

	// 启动监控器
	err := monitor.Start(ctx)
	require.NoError(t, err)

	// 并发访问状态
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 10; j++ {
				status := monitor.GetStatus()
				assert.NotNil(t, status)
				isRunning := monitor.IsRunning()
				assert.True(t, isRunning)
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	// 等待所有协程完成
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent access test")
		}
	}

	monitor.Stop()
}