package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PlaylistManager 播放列表管理器
type PlaylistManager struct {
	sources    map[string]MusicSourcePlugin
	cache      Cache
	repository Repository
	mu         sync.RWMutex
	defaultTTL time.Duration
}

// NewPlaylistManager 创建播放列表管理器
func NewPlaylistManager() *PlaylistManager {
	return &PlaylistManager{
		sources:    make(map[string]MusicSourcePlugin),
		defaultTTL: 15 * time.Minute,
	}
}

// RegisterSource 注册音乐源
func (m *PlaylistManager) RegisterSource(name string, source MusicSourcePlugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sources[name] = source
}

// SetCache 设置缓存
func (m *PlaylistManager) SetCache(cache Cache) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = cache
}

// SetRepository 设置数据仓库
func (m *PlaylistManager) SetRepository(repo Repository) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.repository = repo
}

// GetPlaylist 获取播放列表
func (m *PlaylistManager) GetPlaylist(ctx context.Context, id string, sourceName string) (*Playlist, error) {
	if id == "" {
		return nil, fmt.Errorf("playlist id cannot be empty")
	}

	// 检查缓存
	if m.cache != nil {
		cacheKey := fmt.Sprintf("playlist:%s:%s", sourceName, id)
		if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
			if playlist, ok := cached.(*Playlist); ok {
				return playlist, nil
			}
		}
	}

	// 从音乐源获取
	m.mu.RLock()
	source, exists := m.sources[sourceName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	if baseSrc, ok := source.(*BaseMusicSourcePlugin); ok {
		if !baseSrc.HasFeature(MusicSourceFeaturePlaylist) {
			return nil, fmt.Errorf("source %s does not support playlist feature", sourceName)
		}
	}

	playlist, err := source.GetPlaylist(ctx, id)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if m.cache != nil {
		cacheKey := fmt.Sprintf("playlist:%s:%s", sourceName, id)
		m.cache.Set(ctx, cacheKey, playlist, m.defaultTTL)
	}

	// 保存到仓库
	if m.repository != nil {
		m.repository.SavePlaylist(ctx, playlist)
	}

	return playlist, nil
}

// GetPlaylistTracks 获取播放列表音轨
func (m *PlaylistManager) GetPlaylistTracks(ctx context.Context, playlistID string, sourceName string, offset, limit int) ([]*Track, error) {
	if playlistID == "" {
		return nil, fmt.Errorf("playlist id cannot be empty")
	}

	// 检查缓存
	if m.cache != nil {
		cacheKey := fmt.Sprintf("playlist_tracks:%s:%s:%d:%d", sourceName, playlistID, offset, limit)
		if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
			if tracks, ok := cached.([]*Track); ok {
				return tracks, nil
			}
		}
	}

	// 从音乐源获取
	m.mu.RLock()
	source, exists := m.sources[sourceName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	tracks, err := source.GetPlaylistTracks(ctx, playlistID, offset, limit)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if m.cache != nil {
		cacheKey := fmt.Sprintf("playlist_tracks:%s:%s:%d:%d", sourceName, playlistID, offset, limit)
		m.cache.Set(ctx, cacheKey, tracks, m.defaultTTL)
	}

	return tracks, nil
}

// CreatePlaylist 创建播放列表
func (m *PlaylistManager) CreatePlaylist(ctx context.Context, sourceName string, name, description string) (*Playlist, error) {
	if name == "" {
		return nil, fmt.Errorf("playlist name cannot be empty")
	}

	m.mu.RLock()
	source, exists := m.sources[sourceName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	if baseSrc, ok := source.(*BaseMusicSourcePlugin); ok {
		if !baseSrc.HasFeature(MusicSourceFeaturePlaylist) {
			return nil, fmt.Errorf("source %s does not support playlist feature", sourceName)
		}
	}

	playlist, err := source.CreatePlaylist(ctx, name, description)
	if err != nil {
		return nil, err
	}

	// 保存到仓库
	if m.repository != nil {
		m.repository.SavePlaylist(ctx, playlist)
	}

	// 清除相关缓存
	if m.cache != nil {
		m.invalidatePlaylistCache(ctx, sourceName)
	}

	return playlist, nil
}

// UpdatePlaylist 更新播放列表
func (m *PlaylistManager) UpdatePlaylist(ctx context.Context, sourceName string, playlistID string, updates map[string]interface{}) error {
	if playlistID == "" {
		return fmt.Errorf("playlist id cannot be empty")
	}

	m.mu.RLock()
	source, exists := m.sources[sourceName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("source %s not found", sourceName)
	}

	err := source.UpdatePlaylist(ctx, playlistID, updates)
	if err != nil {
		return err
	}

	// 清除缓存
	if m.cache != nil {
		cacheKey := fmt.Sprintf("playlist:%s:%s", sourceName, playlistID)
		m.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// DeletePlaylist 删除播放列表
func (m *PlaylistManager) DeletePlaylist(ctx context.Context, sourceName string, playlistID string) error {
	if playlistID == "" {
		return fmt.Errorf("playlist id cannot be empty")
	}

	m.mu.RLock()
	source, exists := m.sources[sourceName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("source %s not found", sourceName)
	}

	err := source.DeletePlaylist(ctx, playlistID)
	if err != nil {
		return err
	}

	// 从仓库删除
	if m.repository != nil {
		m.repository.DeletePlaylist(ctx, playlistID)
	}

	// 清除缓存
	if m.cache != nil {
		cacheKey := fmt.Sprintf("playlist:%s:%s", sourceName, playlistID)
		m.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// GetUserPlaylists 获取用户播放列表
func (m *PlaylistManager) GetUserPlaylists(ctx context.Context, sourceName string, userID string) ([]*Playlist, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id cannot be empty")
	}

	// 检查缓存
	if m.cache != nil {
		cacheKey := fmt.Sprintf("user_playlists:%s:%s", sourceName, userID)
		if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
			if playlists, ok := cached.([]*Playlist); ok {
				return playlists, nil
			}
		}
	}

	m.mu.RLock()
	source, exists := m.sources[sourceName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	if baseSrc, ok := source.(*BaseMusicSourcePlugin); ok {
		if !baseSrc.HasFeature(MusicSourceFeatureUser) {
			return nil, fmt.Errorf("source %s does not support user feature", sourceName)
		}
	}

	playlists, err := source.GetUserPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if m.cache != nil {
		cacheKey := fmt.Sprintf("user_playlists:%s:%s", sourceName, userID)
		m.cache.Set(ctx, cacheKey, playlists, m.defaultTTL)
	}

	return playlists, nil
}

// AddTrackToPlaylist 添加音轨到播放列表
func (m *PlaylistManager) AddTrackToPlaylist(ctx context.Context, sourceName string, playlistID string, trackID string) error {
	if playlistID == "" || trackID == "" {
		return fmt.Errorf("playlist id and track id cannot be empty")
	}

	// 这里可以实现添加音轨到播放列表的逻辑
	// 由于接口中没有定义这个方法，我们可以通过更新播放列表来实现
	updates := map[string]interface{}{
		"add_track": trackID,
	}

	return m.UpdatePlaylist(ctx, sourceName, playlistID, updates)
}

// RemoveTrackFromPlaylist 从播放列表移除音轨
func (m *PlaylistManager) RemoveTrackFromPlaylist(ctx context.Context, sourceName string, playlistID string, trackID string) error {
	if playlistID == "" || trackID == "" {
		return fmt.Errorf("playlist id and track id cannot be empty")
	}

	updates := map[string]interface{}{
		"remove_track": trackID,
	}

	return m.UpdatePlaylist(ctx, sourceName, playlistID, updates)
}

// SearchPlaylists 搜索播放列表
func (m *PlaylistManager) SearchPlaylists(ctx context.Context, query string, limit, offset int) ([]*Playlist, error) {
	if m.repository == nil {
		return nil, fmt.Errorf("repository not configured")
	}

	playlists, err := m.repository.SearchPlaylists(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	// 转换为指针切片
	result := make([]*Playlist, len(playlists))
	for i := range playlists {
		result[i] = &playlists[i]
	}

	return result, nil
}

// GetPlaylistsByGenre 按流派获取播放列表
func (m *PlaylistManager) GetPlaylistsByGenre(ctx context.Context, genre string, limit int) ([]*Playlist, error) {
	// 这里可以实现按流派获取播放列表的逻辑
	// 可以从多个音乐源聚合结果
	results := make([]*Playlist, 0)

	m.mu.RLock()
	for _, source := range m.sources {
		if baseSrc, ok := source.(*BaseMusicSourcePlugin); ok && baseSrc.HasFeature(MusicSourceFeatureSearch) {
			// 使用搜索功能查找特定流派的播放列表
			options := SearchOptions{
				Query:  genre,
				Type:   SearchTypePlaylist,
				Limit:  limit,
				Offset: 0,
				Filters: map[string]interface{}{
					"genre": genre,
				},
			}

			result, err := source.Search(context.Background(), genre, options)
			if err == nil && result != nil {
				for _, playlist := range result.Playlists {
					results = append(results, &playlist)
				}
			}
		}
	}
	m.mu.RUnlock()

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// invalidatePlaylistCache 清除播放列表相关缓存
func (m *PlaylistManager) invalidatePlaylistCache(ctx context.Context, sourceName string) {
	if m.cache == nil {
		return
	}

	// 获取所有相关的缓存键并删除
	pattern := fmt.Sprintf("*playlist*:%s:*", sourceName)
	keys, err := m.cache.Keys(ctx, pattern)
	if err != nil {
		return
	}

	for _, key := range keys {
		m.cache.Delete(ctx, key)
	}
}

// GetPlaylistStatistics 获取播放列表统计信息
func (m *PlaylistManager) GetPlaylistStatistics(ctx context.Context, playlistID string, sourceName string) (map[string]interface{}, error) {
	playlist, err := m.GetPlaylist(ctx, playlistID, sourceName)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"track_count":    playlist.TrackCount,
		"total_duration": playlist.Duration.String(),
		"created_at":     playlist.CreatedAt,
		"updated_at":     playlist.UpdatedAt,
		"is_public":      playlist.Public,
		"owner":          playlist.Owner,
	}

	// 计算额外统计信息
	if len(playlist.Tracks) > 0 {
		genres := make(map[string]int)
		for _, track := range playlist.Tracks {
			if track.Genre != "" {
				genres[track.Genre]++
			}
		}
		stats["genres"] = genres
	}

	return stats, nil
}