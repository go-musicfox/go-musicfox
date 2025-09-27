package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/plugins/audio"
	"github.com/go-musicfox/go-musicfox/v2/plugins/playlist"
)

// StatusMonitor 状态监控器
type StatusMonitor struct {
	audioPlugin    *audio.AudioPlugin
	playlistPlugin *playlist.PlaylistPluginImpl
	eventBus       event.EventBus
	logger         *slog.Logger
	subscriptions  []string
	audioSub       string
	playlistSub    string
	running        bool
	mutex          sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewStatusMonitor 创建状态监控器
func NewStatusMonitor(audioPlugin *audio.AudioPlugin, playlistPlugin *playlist.PlaylistPluginImpl, eventBus event.EventBus, logger *slog.Logger) *StatusMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &StatusMonitor{
		audioPlugin:    audioPlugin,
		playlistPlugin: playlistPlugin,
		eventBus:       eventBus,
		logger:         logger,
		subscriptions:  make([]string, 0),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start 启动状态监控
func (sm *StatusMonitor) Start(ctx context.Context) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.running {
		return fmt.Errorf("status monitor already running")
	}

	sm.logger.Info("Starting status monitor")

	// 订阅音频事件
	if err := sm.subscribeToAudioEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to audio events: %w", err)
	}

	// 订阅播放列表事件
	if err := sm.subscribeToPlaylistEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to playlist events: %w", err)
	}

	// 启动状态更新协程
	go sm.statusUpdateLoop()

	sm.running = true
	sm.logger.Info("Status monitor started successfully")
	return nil
}

// Stop 停止状态监控
func (sm *StatusMonitor) Stop() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if !sm.running {
		return
	}

	sm.logger.Info("Stopping status monitor")

	// 取消上下文
	sm.cancel()

	// 取消所有事件订阅
	for _, subID := range sm.subscriptions {
		if err := sm.eventBus.Unsubscribe(subID); err != nil {
			sm.logger.Warn("Failed to unsubscribe from event", "subscription_id", subID, "error", err)
		}
	}
	sm.subscriptions = sm.subscriptions[:0]

	sm.running = false
	sm.logger.Info("Status monitor stopped")
}

// subscribeToAudioEvents 订阅音频事件
func (sm *StatusMonitor) subscribeToAudioEvents() error {
	// 订阅播放开始事件
	playStartSub, err := sm.eventBus.Subscribe("audio.play.start", sm.handlePlayStartEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to play start event: %w", err)
	}
	sm.subscriptions = append(sm.subscriptions, playStartSub.ID)

	// 订阅播放暂停事件
	playPauseSub, err := sm.eventBus.Subscribe("audio.play.pause", sm.handlePlayPauseEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to play pause event: %w", err)
	}
	sm.subscriptions = append(sm.subscriptions, playPauseSub.ID)

	// 订阅播放停止事件
	playStopSub, err := sm.eventBus.Subscribe("audio.play.stop", sm.handlePlayStopEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to play stop event: %w", err)
	}
	sm.subscriptions = append(sm.subscriptions, playStopSub.ID)

	// 订阅播放恢复事件
	playResumeSub, err := sm.eventBus.Subscribe("audio.play.resume", sm.handlePlayResumeEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to play resume event: %w", err)
	}
	sm.subscriptions = append(sm.subscriptions, playResumeSub.ID)

	// 订阅音量变化事件
	volumeChangeSub, err := sm.eventBus.Subscribe("audio.volume.change", sm.handleVolumeChangeEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to volume change event: %w", err)
	}
	sm.subscriptions = append(sm.subscriptions, volumeChangeSub.ID)

	// 订阅播放错误事件
	playErrorSub, err := sm.eventBus.Subscribe("audio.play.error", sm.handlePlayErrorEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to play error event: %w", err)
	}
	sm.subscriptions = append(sm.subscriptions, playErrorSub.ID)

	return nil
}

// subscribeToPlaylistEvents 订阅播放列表事件
func (sm *StatusMonitor) subscribeToPlaylistEvents() error {
	// 订阅播放列表创建事件
	playlistCreateSub, err := sm.eventBus.Subscribe("playlist.created", sm.handlePlaylistCreateEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to playlist create event: %w", err)
	}
	sm.subscriptions = append(sm.subscriptions, playlistCreateSub.ID)

	// 订阅播放列表更新事件
	playlistUpdateSub, err := sm.eventBus.Subscribe("playlist.updated", sm.handlePlaylistUpdateEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to playlist update event: %w", err)
	}
	sm.subscriptions = append(sm.subscriptions, playlistUpdateSub.ID)

	// 订阅队列变化事件
	queueChangeSub, err := sm.eventBus.Subscribe("queue.changed", sm.handleQueueChangeEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to queue change event: %w", err)
	}
	sm.subscriptions = append(sm.subscriptions, queueChangeSub.ID)

	// 订阅播放模式变化事件
	playModeChangeSub, err := sm.eventBus.Subscribe("playmode.changed", sm.handlePlayModeChangeEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to play mode change event: %w", err)
	}
	sm.subscriptions = append(sm.subscriptions, playModeChangeSub.ID)

	return nil
}

// statusUpdateLoop 状态更新循环
func (sm *StatusMonitor) statusUpdateLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.updatePlaybackPosition()
		}
	}
}

// updatePlaybackPosition 更新播放位置
func (sm *StatusMonitor) updatePlaybackPosition() {
	state := sm.audioPlugin.GetState()
	if state == nil || state.CurrentSong == nil {
		return
	}

	// 获取当前播放位置
	position := sm.audioPlugin.GetPosition()
	duration := sm.audioPlugin.GetDuration()

	// 发布位置更新事件
	if sm.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("position-update-%d", time.Now().UnixNano()),
			Type:      "audio.position.update",
			Source:    "status_monitor",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"position": position,
				"duration": duration,
				"song_id":  state.CurrentSong.ID,
			},
		}
		sm.eventBus.PublishAsync(context.Background(), event)
	}
}

// 事件处理器

// handlePlayStartEvent 处理播放开始事件
func (sm *StatusMonitor) handlePlayStartEvent(ctx context.Context, event event.Event) error {
	sm.logger.Info("Play started", "event_data", event.GetData())
	return nil
}

// handlePlayPauseEvent 处理播放暂停事件
func (sm *StatusMonitor) handlePlayPauseEvent(ctx context.Context, event event.Event) error {
	sm.logger.Info("Play paused", "event_data", event.GetData())
	return nil
}

// handlePlayStopEvent 处理播放停止事件
func (sm *StatusMonitor) handlePlayStopEvent(ctx context.Context, event event.Event) error {
	sm.logger.Info("Play stopped", "event_data", event.GetData())
	return nil
}

// handlePlayResumeEvent 处理播放恢复事件
func (sm *StatusMonitor) handlePlayResumeEvent(ctx context.Context, event event.Event) error {
	sm.logger.Info("Play resumed", "event_data", event.GetData())
	return nil
}

// handleVolumeChangeEvent 处理音量变化事件
func (sm *StatusMonitor) handleVolumeChangeEvent(ctx context.Context, event event.Event) error {
	data := event.GetData()
	if dataMap, ok := data.(map[string]interface{}); ok {
		if volume, ok := dataMap["volume"]; ok {
			sm.logger.Info("Volume changed", "volume", volume)
		}
	}
	return nil
}

// handlePlaylistEvent 处理播放列表事件
func (sm *StatusMonitor) handlePlaylistEvent(ctx context.Context, event event.Event) error {
	data := event.GetData()
	if dataMap, ok := data.(map[string]interface{}); ok {
		if playlist, ok := dataMap["playlist"]; ok {
			sm.logger.Info("Playlist event received", "playlist", playlist)
		}
	}
	return nil
}

// handlePlayErrorEvent 处理播放错误事件
func (sm *StatusMonitor) handlePlayErrorEvent(ctx context.Context, event event.Event) error {
	data := event.GetData()
	if dataMap, ok := data.(map[string]interface{}); ok {
		if errorMsg, ok := dataMap["error"]; ok {
			sm.logger.Error("Play error occurred", "error", errorMsg)
		}
	}
	return nil
}

// handlePlaylistCreateEvent 处理播放列表创建事件
func (sm *StatusMonitor) handlePlaylistCreateEvent(ctx context.Context, event event.Event) error {
	data := event.GetData()
	if dataMap, ok := data.(map[string]interface{}); ok {
		if playlistName, ok := dataMap["playlist_name"]; ok {
			sm.logger.Info("Playlist created", "name", playlistName)
		}
	}
	return nil
}

// handlePlaylistUpdateEvent 处理播放列表更新事件
func (sm *StatusMonitor) handlePlaylistUpdateEvent(ctx context.Context, event event.Event) error {
	data := event.GetData()
	if dataMap, ok := data.(map[string]interface{}); ok {
		if playlistID, ok := dataMap["playlist_id"]; ok {
			sm.logger.Info("Playlist updated", "id", playlistID)
		}
	}
	return nil
}

// handleQueueChangeEvent 处理队列变化事件
func (sm *StatusMonitor) handleQueueChangeEvent(ctx context.Context, event event.Event) error {
	data := event.GetData()
	if dataMap, ok := data.(map[string]interface{}); ok {
		if queueInfo, ok := dataMap["queue_info"]; ok {
			sm.logger.Info("Queue changed", "queue_info", queueInfo)
		}
	}
	return nil
}

// handlePlayModeChangeEvent 处理播放模式变化事件
func (sm *StatusMonitor) handlePlayModeChangeEvent(ctx context.Context, event event.Event) error {
	data := event.GetData()
	if dataMap, ok := data.(map[string]interface{}); ok {
		if playMode, ok := dataMap["play_mode"]; ok {
			sm.logger.Info("Play mode changed", "play_mode", playMode)
		}
	}
	return nil
}

// GetStatus 获取当前状态
func (sm *StatusMonitor) GetStatus() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	status := map[string]interface{}{
		"running":      sm.running,
		"subscriptions": len(sm.subscriptions),
	}

	// 获取音频状态
	if sm.audioPlugin != nil {
		audioState := sm.audioPlugin.GetState()
		if audioState != nil {
			status["audio"] = map[string]interface{}{
				"status":       audioState.Status,
				"current_song": audioState.CurrentSong,
				"position":     audioState.Position,
				"duration":     audioState.Duration,
				"volume":       sm.audioPlugin.GetVolume(),
			}
		}
	}

	// 获取播放列表状态
	if sm.playlistPlugin != nil {
		queue, _ := sm.playlistPlugin.GetCurrentQueue(context.Background())
		playMode := sm.playlistPlugin.GetPlayMode(context.Background())
		status["playlist"] = map[string]interface{}{
			"queue_size": len(queue),
			"play_mode":  playMode,
		}
	}

	return status
}

// IsRunning 检查监控器是否运行中
func (sm *StatusMonitor) IsRunning() bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.running
}