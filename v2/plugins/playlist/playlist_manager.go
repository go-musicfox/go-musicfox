package playlist

import (
	"context"
	"fmt"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
)

// CreatePlaylist 创建播放列表
func (p *PlaylistPluginImpl) CreatePlaylist(ctx context.Context, name, description string) (*model.Playlist, error) {
	if name == "" {
		return nil, fmt.Errorf("playlist name cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 生成唯一ID
	playlistID := fmt.Sprintf("playlist_%d", time.Now().UnixNano())

	// 创建播放列表
	playlist := model.NewPlaylist(playlistID, name, "local", "user")
	playlist.Description = description

	// 存储播放列表
	p.playlists[playlistID] = playlist

	// 发送创建事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playlist_created_%d", time.Now().UnixNano()),
			Type:      event.EventPlaylistCreated,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"playlist_id":   playlistID,
				"playlist_name": name,
				"description":   description,
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return playlist, nil
}

// DeletePlaylist 删除播放列表
func (p *PlaylistPluginImpl) DeletePlaylist(ctx context.Context, playlistID string) error {
	if playlistID == "" {
		return fmt.Errorf("playlist ID cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查播放列表是否存在
	playlist, exists := p.playlists[playlistID]
	if !exists {
		return fmt.Errorf("playlist not found: %s", playlistID)
	}

	// 删除播放列表
	delete(p.playlists, playlistID)

	// 发送删除事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playlist_deleted_%d", time.Now().UnixNano()),
			Type:      event.EventPlaylistDeleted,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"playlist_id":   playlistID,
				"playlist_name": playlist.Name,
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// UpdatePlaylist 更新播放列表
func (p *PlaylistPluginImpl) UpdatePlaylist(ctx context.Context, playlistID string, updates map[string]interface{}) error {
	if playlistID == "" {
		return fmt.Errorf("playlist ID cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查播放列表是否存在
	playlist, exists := p.playlists[playlistID]
	if !exists {
		return fmt.Errorf("playlist not found: %s", playlistID)
	}

	// 更新字段
	if name, ok := updates["name"].(string); ok && name != "" {
		playlist.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		playlist.Description = description
	}
	if isPublic, ok := updates["is_public"].(bool); ok {
		playlist.IsPublic = isPublic
	}

	playlist.UpdatedAt = time.Now()

	// 发送更新事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playlist_updated_%d", time.Now().UnixNano()),
			Type:      event.EventPlaylistUpdated,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"playlist_id": playlistID,
				"updates":     updates,
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// GetPlaylist 获取播放列表
func (p *PlaylistPluginImpl) GetPlaylist(ctx context.Context, playlistID string) (*model.Playlist, error) {
	if playlistID == "" {
		return nil, fmt.Errorf("playlist ID cannot be empty")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	playlist, exists := p.playlists[playlistID]
	if !exists {
		return nil, fmt.Errorf("playlist not found: %s", playlistID)
	}

	// 返回播放列表的副本
	playlistCopy := *playlist
	playlistCopy.Songs = make([]*model.Song, len(playlist.Songs))
	copy(playlistCopy.Songs, playlist.Songs)

	return &playlistCopy, nil
}

// ListPlaylists 列出所有播放列表
func (p *PlaylistPluginImpl) ListPlaylists(ctx context.Context) ([]*model.Playlist, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	playlists := make([]*model.Playlist, 0, len(p.playlists))
	for _, playlist := range p.playlists {
		// 创建播放列表副本
		playlistCopy := *playlist
		playlistCopy.Songs = make([]*model.Song, len(playlist.Songs))
		copy(playlistCopy.Songs, playlist.Songs)
		playlists = append(playlists, &playlistCopy)
	}

	return playlists, nil
}

// AddSong 向播放列表添加歌曲
func (p *PlaylistPluginImpl) AddSong(ctx context.Context, playlistID string, song *model.Song) error {
	if playlistID == "" {
		return fmt.Errorf("playlist ID cannot be empty")
	}
	if song == nil {
		return fmt.Errorf("song cannot be nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查播放列表是否存在
	playlist, exists := p.playlists[playlistID]
	if !exists {
		return fmt.Errorf("playlist not found: %s", playlistID)
	}

	// 检查歌曲是否已存在
	for _, existingSong := range playlist.Songs {
		if existingSong.ID == song.ID {
			return fmt.Errorf("song already exists in playlist: %s", song.ID)
		}
	}

	// 添加歌曲
	playlist.AddSong(song)

	// 发送添加事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playlist_song_added_%d", time.Now().UnixNano()),
			Type:      event.EventPlaylistSongAdded,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"playlist_id": playlistID,
				"song_id":     song.ID,
				"song_title":  song.Title,
				"song_artist": song.Artist,
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// RemoveSong 从播放列表移除歌曲
func (p *PlaylistPluginImpl) RemoveSong(ctx context.Context, playlistID string, songID string) error {
	if playlistID == "" {
		return fmt.Errorf("playlist ID cannot be empty")
	}
	if songID == "" {
		return fmt.Errorf("song ID cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查播放列表是否存在
	playlist, exists := p.playlists[playlistID]
	if !exists {
		return fmt.Errorf("playlist not found: %s", playlistID)
	}

	// 移除歌曲
	if !playlist.RemoveSong(songID) {
		return fmt.Errorf("song not found in playlist: %s", songID)
	}

	// 发送移除事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playlist_song_removed_%d", time.Now().UnixNano()),
			Type:      event.EventPlaylistSongRemoved,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"playlist_id": playlistID,
				"song_id":     songID,
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// MoveSong 移动播放列表中的歌曲位置
func (p *PlaylistPluginImpl) MoveSong(ctx context.Context, playlistID string, songID string, newIndex int) error {
	if playlistID == "" {
		return fmt.Errorf("playlist ID cannot be empty")
	}
	if songID == "" {
		return fmt.Errorf("song ID cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查播放列表是否存在
	playlist, exists := p.playlists[playlistID]
	if !exists {
		return fmt.Errorf("playlist not found: %s", playlistID)
	}

	// 找到歌曲的当前位置
	oldIndex := -1
	for i, song := range playlist.Songs {
		if song.ID == songID {
			oldIndex = i
			break
		}
	}

	if oldIndex == -1 {
		return fmt.Errorf("song not found in playlist: %s", songID)
	}

	// 检查新位置是否有效
	if newIndex < 0 || newIndex >= len(playlist.Songs) {
		return fmt.Errorf("invalid new index: %d", newIndex)
	}

	// 如果位置没有变化，直接返回
	if oldIndex == newIndex {
		return nil
	}

	// 移动歌曲
	song := playlist.Songs[oldIndex]
	
	// 创建新的歌曲切片
	newSongs := make([]*model.Song, 0, len(playlist.Songs))
	
	// 根据移动方向处理
	if oldIndex < newIndex {
		// 向后移动：复制oldIndex之前的元素
		newSongs = append(newSongs, playlist.Songs[:oldIndex]...)
		// 复制oldIndex+1到newIndex的元素
		newSongs = append(newSongs, playlist.Songs[oldIndex+1:newIndex+1]...)
		// 插入移动的歌曲
		newSongs = append(newSongs, song)
		// 复制newIndex+1之后的元素
		newSongs = append(newSongs, playlist.Songs[newIndex+1:]...)
	} else {
		// 向前移动：复制newIndex之前的元素
		newSongs = append(newSongs, playlist.Songs[:newIndex]...)
		// 插入移动的歌曲
		newSongs = append(newSongs, song)
		// 复制newIndex到oldIndex-1的元素
		newSongs = append(newSongs, playlist.Songs[newIndex:oldIndex]...)
		// 复制oldIndex+1之后的元素
		newSongs = append(newSongs, playlist.Songs[oldIndex+1:]...)
	}
	
	playlist.Songs = newSongs
	playlist.UpdatedAt = time.Now()

	// 发送移动事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playlist_song_moved_%d", time.Now().UnixNano()),
			Type:      event.EventPlaylistSongMoved,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"playlist_id": playlistID,
				"song_id":     songID,
				"old_index":   oldIndex,
				"new_index":   newIndex,
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// ClearPlaylist 清空播放列表
func (p *PlaylistPluginImpl) ClearPlaylist(ctx context.Context, playlistID string) error {
	if playlistID == "" {
		return fmt.Errorf("playlist ID cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查播放列表是否存在
	playlist, exists := p.playlists[playlistID]
	if !exists {
		return fmt.Errorf("playlist not found: %s", playlistID)
	}

	// 清空歌曲列表
	playlist.Songs = make([]*model.Song, 0)
	playlist.UpdatedAt = time.Now()

	// 发送清空事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playlist_cleared_%d", time.Now().UnixNano()),
			Type:      event.EventPlaylistCleared,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"playlist_id": playlistID,
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}