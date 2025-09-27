package playlist

import (
	"context"
	"fmt"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
)

// EventIntegration 事件集成管理器
type EventIntegration struct {
	plugin   *PlaylistPluginImpl
	eventBus event.EventBus
}

// NewEventIntegration 创建事件集成管理器
func NewEventIntegration(plugin *PlaylistPluginImpl, eventBus event.EventBus) *EventIntegration {
	return &EventIntegration{
		plugin:   plugin,
		eventBus: eventBus,
	}
}

// RegisterEventHandlers 注册事件处理器
func (ei *EventIntegration) RegisterEventHandlers() error {
	if ei.eventBus == nil {
		return fmt.Errorf("event bus is not available")
	}

	// 注册播放器事件处理器
	if err := ei.registerPlayerEventHandlers(); err != nil {
		return fmt.Errorf("failed to register player event handlers: %w", err)
	}

	// 注册系统事件处理器
	if err := ei.registerSystemEventHandlers(); err != nil {
		return fmt.Errorf("failed to register system event handlers: %w", err)
	}

	return nil
}

// registerPlayerEventHandlers 注册播放器事件处理器
func (ei *EventIntegration) registerPlayerEventHandlers() error {
	// 监听歌曲变化事件
	_, err := ei.eventBus.Subscribe(event.EventPlayerSongChanged, ei.handlePlayerSongChanged)
	if err != nil {
		return fmt.Errorf("failed to subscribe to song changed event: %w", err)
	}

	// 监听播放状态变化事件
	_, err = ei.eventBus.Subscribe(event.EventPlayerStateChanged, ei.handlePlayerStateChanged)
	if err != nil {
		return fmt.Errorf("failed to subscribe to state changed event: %w", err)
	}

	// 监听播放器错误事件
	_, err = ei.eventBus.Subscribe(event.EventPlayerError, ei.handlePlayerError)
	if err != nil {
		return fmt.Errorf("failed to subscribe to player error event: %w", err)
	}

	// 监听下一首歌曲请求事件
	_, err = ei.eventBus.Subscribe(event.EventPlayerNext, ei.handlePlayerNext)
	if err != nil {
		return fmt.Errorf("failed to subscribe to player next event: %w", err)
	}

	// 监听上一首歌曲请求事件
	_, err = ei.eventBus.Subscribe(event.EventPlayerPrevious, ei.handlePlayerPrevious)
	if err != nil {
		return fmt.Errorf("failed to subscribe to player previous event: %w", err)
	}

	return nil
}

// registerSystemEventHandlers 注册系统事件处理器
func (ei *EventIntegration) registerSystemEventHandlers() error {
	// 监听配置变化事件
	_, err := ei.eventBus.Subscribe(event.EventConfigChanged, ei.handleConfigChanged)
	if err != nil {
		return fmt.Errorf("failed to subscribe to config changed event: %w", err)
	}

	// 监听系统关闭事件
	_, err = ei.eventBus.Subscribe(event.EventSystemShutdown, ei.handleSystemShutdown)
	if err != nil {
		return fmt.Errorf("failed to subscribe to system shutdown event: %w", err)
	}

	return nil
}

// handlePlayerSongChanged 处理歌曲变化事件
func (ei *EventIntegration) handlePlayerSongChanged(ctx context.Context, e event.Event) error {
	data := e.GetData()
	if data == nil {
		return nil
	}

	// 解析事件数据
	eventData, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event data format")
	}

	// 获取歌曲信息
	songID, ok := eventData["song_id"].(string)
	if !ok || songID == "" {
		return fmt.Errorf("missing or invalid song_id")
	}

	// 从当前队列中找到歌曲
	ei.plugin.mu.RLock()
	var song *model.Song
	for _, queueSong := range ei.plugin.currentQueue {
		if queueSong.ID == songID {
			song = queueSong
			break
		}
	}
	ei.plugin.mu.RUnlock()

	if song != nil {
		// 添加到播放历史
		if err := ei.plugin.AddToHistory(ctx, song); err != nil {
			return fmt.Errorf("failed to add song to history: %w", err)
		}
	}

	return nil
}

// handlePlayerStateChanged 处理播放状态变化事件
func (ei *EventIntegration) handlePlayerStateChanged(ctx context.Context, e event.Event) error {
	data := e.GetData()
	if data == nil {
		return nil
	}

	// 解析事件数据
	eventData, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event data format")
	}

	// 获取播放状态
	state, ok := eventData["state"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid state")
	}

	// 根据状态执行相应操作
	switch state {
	case "stopped":
		// 播放停止时，可以触发自动播放下一首
		if ei.shouldAutoPlayNext() {
			if err := ei.triggerNextSong(ctx); err != nil {
				return fmt.Errorf("failed to trigger next song: %w", err)
			}
		}
	case "error":
		// 播放错误时，尝试播放下一首
		if err := ei.handlePlaybackError(ctx, eventData); err != nil {
			return fmt.Errorf("failed to handle playback error: %w", err)
		}
	}

	return nil
}

// handlePlayerError 处理播放器错误事件
func (ei *EventIntegration) handlePlayerError(ctx context.Context, e event.Event) error {
	data := e.GetData()
	if data == nil {
		return nil
	}

	// 解析事件数据
	eventData, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event data format")
	}

	// 处理播放错误
	return ei.handlePlaybackError(ctx, eventData)
}

// handlePlayerNext 处理下一首歌曲请求事件
func (ei *EventIntegration) handlePlayerNext(ctx context.Context, e event.Event) error {
	return ei.triggerNextSong(ctx)
}

// handlePlayerPrevious 处理上一首歌曲请求事件
func (ei *EventIntegration) handlePlayerPrevious(ctx context.Context, e event.Event) error {
	return ei.triggerPreviousSong(ctx)
}

// handleConfigChanged 处理配置变化事件
func (ei *EventIntegration) handleConfigChanged(ctx context.Context, e event.Event) error {
	data := e.GetData()
	if data == nil {
		return nil
	}

	// 解析事件数据
	eventData, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event data format")
	}

	// 检查是否有播放列表相关的配置变化
	if configSection, ok := eventData["section"].(string); ok && configSection == "playlist" {
		if configData, ok := eventData["config"].(map[string]interface{}); ok {
			return ei.plugin.UpdateConfig(configData)
		}
	}

	return nil
}

// handleSystemShutdown 处理系统关闭事件
func (ei *EventIntegration) handleSystemShutdown(ctx context.Context, e event.Event) error {
	// 保存当前状态（如果需要持久化）
	return ei.saveCurrentState(ctx)
}

// shouldAutoPlayNext 检查是否应该自动播放下一首
func (ei *EventIntegration) shouldAutoPlayNext() bool {
	ei.plugin.mu.RLock()
	defer ei.plugin.mu.RUnlock()

	// 根据播放模式决定是否自动播放下一首
	switch ei.plugin.playMode {
	case model.PlayModeSequential:
		// 顺序播放模式下，如果不是最后一首则自动播放
		return ei.plugin.currentIndex < len(ei.plugin.currentQueue)-1
	case model.PlayModeRepeatOne:
		// 单曲循环模式下，总是自动播放（重复当前歌曲）
		return true
	case model.PlayModeRepeatAll:
		// 列表循环模式下，总是自动播放
		return len(ei.plugin.currentQueue) > 0
	case model.PlayModeShuffle:
		// 随机播放模式下，总是自动播放
		return len(ei.plugin.currentQueue) > 0
	default:
		return false
	}
}

// triggerNextSong 触发播放下一首歌曲
func (ei *EventIntegration) triggerNextSong(ctx context.Context) error {
	// 获取当前播放的歌曲
	var currentSong *model.Song
	ei.plugin.mu.RLock()
	if ei.plugin.currentIndex >= 0 && ei.plugin.currentIndex < len(ei.plugin.currentQueue) {
		currentSong = ei.plugin.currentQueue[ei.plugin.currentIndex]
	}
	ei.plugin.mu.RUnlock()

	// 获取下一首歌曲
	nextSong, err := ei.plugin.GetNextSong(ctx, currentSong)
	if err != nil {
		return fmt.Errorf("failed to get next song: %w", err)
	}

	// 发送播放下一首歌曲的事件
	event := &event.BaseEvent{
		ID:        fmt.Sprintf("auto_next_song_%d", time.Now().UnixNano()),
		Type:      "playlist.auto_next",
		Source:    ei.plugin.info.ID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"song_id":     nextSong.ID,
			"song_title":  nextSong.Title,
			"song_artist": nextSong.Artist,
			"song_url":    nextSong.URL,
		},
	}

	return ei.eventBus.Publish(ctx, event)
}

// triggerPreviousSong 触发播放上一首歌曲
func (ei *EventIntegration) triggerPreviousSong(ctx context.Context) error {
	// 获取当前播放的歌曲
	var currentSong *model.Song
	ei.plugin.mu.RLock()
	if ei.plugin.currentIndex >= 0 && ei.plugin.currentIndex < len(ei.plugin.currentQueue) {
		currentSong = ei.plugin.currentQueue[ei.plugin.currentIndex]
	}
	ei.plugin.mu.RUnlock()

	// 获取上一首歌曲
	previousSong, err := ei.plugin.GetPreviousSong(ctx, currentSong)
	if err != nil {
		return fmt.Errorf("failed to get previous song: %w", err)
	}

	// 发送播放上一首歌曲的事件
	event := &event.BaseEvent{
		ID:        fmt.Sprintf("auto_previous_song_%d", time.Now().UnixNano()),
		Type:      "playlist.auto_previous",
		Source:    ei.plugin.info.ID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"song_id":     previousSong.ID,
			"song_title":  previousSong.Title,
			"song_artist": previousSong.Artist,
			"song_url":    previousSong.URL,
		},
	}

	return ei.eventBus.Publish(ctx, event)
}

// handlePlaybackError 处理播放错误
func (ei *EventIntegration) handlePlaybackError(ctx context.Context, eventData map[string]interface{}) error {
	// 获取错误信息
	errorMsg, _ := eventData["error"].(string)
	songID, _ := eventData["song_id"].(string)

	// 记录错误
	ei.logPlaybackError(songID, errorMsg)

	// 根据错误类型决定是否跳过当前歌曲
	if ei.shouldSkipOnError(errorMsg) {
		return ei.triggerNextSong(ctx)
	}

	return nil
}

// shouldSkipOnError 检查是否应该因错误跳过当前歌曲
func (ei *EventIntegration) shouldSkipOnError(errorMsg string) bool {
	// 定义需要跳过的错误类型
	skipErrors := []string{
		"file not found",
		"network timeout",
		"invalid format",
		"codec error",
	}

	for _, skipError := range skipErrors {
		if len(errorMsg) > 0 && len(skipError) > 0 {
			// 简单的字符串包含检查
			for i := 0; i <= len(errorMsg)-len(skipError); i++ {
				if errorMsg[i:i+len(skipError)] == skipError {
					return true
				}
			}
		}
	}

	return false
}

// logPlaybackError 记录播放错误
func (ei *EventIntegration) logPlaybackError(songID, errorMsg string) {
	// 发送错误事件
	if ei.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playback_error_%d", time.Now().UnixNano()),
			Type:      "playlist.playback_error",
			Source:    ei.plugin.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"song_id": songID,
				"error":   errorMsg,
			},
		}
		ei.eventBus.Publish(context.Background(), event)
	}
}

// saveCurrentState 保存当前状态
func (ei *EventIntegration) saveCurrentState(ctx context.Context) error {
	// 这里可以实现状态持久化逻辑
	// 例如保存当前播放列表、队列、播放模式等到文件或数据库

	// 发送状态保存事件
	if ei.eventBus != nil {
		ei.plugin.mu.RLock()
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("state_saved_%d", time.Now().UnixNano()),
			Type:      "playlist.state_saved",
			Source:    ei.plugin.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"playlists_count": len(ei.plugin.playlists),
				"queue_length":    len(ei.plugin.currentQueue),
				"history_length":  len(ei.plugin.history),
				"play_mode":       ei.plugin.playMode.String(),
			},
		}
		ei.plugin.mu.RUnlock()
		ei.eventBus.Publish(ctx, event)
	}

	return nil
}