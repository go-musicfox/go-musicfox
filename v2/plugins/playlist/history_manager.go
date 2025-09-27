package playlist

import (
	"context"
	"fmt"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
)

// AddToHistory 添加歌曲到播放历史
func (p *PlaylistPluginImpl) AddToHistory(ctx context.Context, song *model.Song) error {
	if song == nil {
		return fmt.Errorf("song cannot be nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查是否已经是最近播放的歌曲
	if len(p.history) > 0 && p.history[len(p.history)-1].ID == song.ID {
		return nil // 避免重复添加相同歌曲
	}

	// 创建歌曲副本，添加播放时间戳
	historySong := *song
	if historySong.Metadata == nil {
		historySong.Metadata = make(map[string]string)
	}
	historySong.Metadata["played_at"] = time.Now().Format(time.RFC3339)

	// 添加到历史记录
	p.history = append(p.history, &historySong)

	// 如果历史记录超过最大限制，移除最旧的记录
	if len(p.history) > p.maxHistory {
		p.history = p.history[len(p.history)-p.maxHistory:]
	}

	// 发送添加到历史事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("history_song_added_%d", time.Now().UnixNano()),
			Type:      "history.song_added",
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"song_id":        song.ID,
				"song_title":     song.Title,
				"song_artist":    song.Artist,
				"history_length": len(p.history),
				"played_at":      historySong.Metadata["played_at"],
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// GetHistory 获取播放历史
func (p *PlaylistPluginImpl) GetHistory(ctx context.Context, limit int) ([]*model.Song, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 如果没有指定限制或限制无效，返回所有历史
	if limit <= 0 {
		limit = len(p.history)
	}

	// 计算起始位置（返回最近的记录）
	start := len(p.history) - limit
	if start < 0 {
		start = 0
	}

	// 创建历史记录副本
	history := make([]*model.Song, len(p.history)-start)
	for i, song := range p.history[start:] {
		// 创建歌曲副本
		songCopy := *song
		songCopy.Metadata = make(map[string]string)
		for k, v := range song.Metadata {
			songCopy.Metadata[k] = v
		}
		history[i] = &songCopy
	}

	// 反转数组，使最新的记录在前面
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

// ClearHistory 清空播放历史
func (p *PlaylistPluginImpl) ClearHistory(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 清空历史记录
	p.history = make([]*model.Song, 0)

	// 发送历史清空事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("history_cleared_%d", time.Now().UnixNano()),
			Type:      "history.cleared",
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"history_length": 0,
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// GetRecentlyPlayed 获取最近播放的歌曲（去重）
func (p *PlaylistPluginImpl) GetRecentlyPlayed(ctx context.Context, limit int) ([]*model.Song, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if limit <= 0 {
		limit = 10 // 默认返回10首
	}

	// 使用map去重，保持最新的播放记录
	seenSongs := make(map[string]*model.Song)
	recentSongs := make([]*model.Song, 0, limit)

	// 从最新的历史记录开始遍历
	for i := len(p.history) - 1; i >= 0 && len(recentSongs) < limit; i-- {
		song := p.history[i]
		if _, exists := seenSongs[song.ID]; !exists {
			seenSongs[song.ID] = song
			// 创建歌曲副本
			songCopy := *song
			songCopy.Metadata = make(map[string]string)
			for k, v := range song.Metadata {
				songCopy.Metadata[k] = v
			}
			recentSongs = append(recentSongs, &songCopy)
		}
	}

	return recentSongs, nil
}

// GetMostPlayed 获取播放次数最多的歌曲
func (p *PlaylistPluginImpl) GetMostPlayed(ctx context.Context, limit int) ([]*model.Song, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if limit <= 0 {
		limit = 10 // 默认返回10首
	}

	// 统计播放次数
	playCount := make(map[string]int)
	songMap := make(map[string]*model.Song)

	for _, song := range p.history {
		playCount[song.ID]++
		songMap[song.ID] = song
	}

	// 创建结果切片
	type songWithCount struct {
		song  *model.Song
		count int
	}

	songs := make([]songWithCount, 0, len(songMap))
	for songID, song := range songMap {
		songs = append(songs, songWithCount{
			song:  song,
			count: playCount[songID],
		})
	}

	// 按播放次数排序（冒泡排序，简单实现）
	for i := 0; i < len(songs)-1; i++ {
		for j := 0; j < len(songs)-1-i; j++ {
			if songs[j].count < songs[j+1].count {
				songs[j], songs[j+1] = songs[j+1], songs[j]
			}
		}
	}

	// 取前limit首歌曲
	result := make([]*model.Song, 0, limit)
	for i := 0; i < len(songs) && i < limit; i++ {
		// 创建歌曲副本，添加播放次数信息
		songCopy := *songs[i].song
		songCopy.Metadata = make(map[string]string)
		for k, v := range songs[i].song.Metadata {
			songCopy.Metadata[k] = v
		}
		songCopy.Metadata["play_count"] = fmt.Sprintf("%d", songs[i].count)
		result = append(result, &songCopy)
	}

	return result, nil
}

// GetHistoryStats 获取播放历史统计信息
func (p *PlaylistPluginImpl) GetHistoryStats(ctx context.Context) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]interface{})

	// 基本统计
	stats["total_plays"] = len(p.history)
	stats["unique_songs"] = len(p.getUniqueSongs())

	// 艺术家统计
	artistCount := make(map[string]int)
	for _, song := range p.history {
		artistCount[song.Artist]++
	}
	stats["unique_artists"] = len(artistCount)

	// 专辑统计
	albumCount := make(map[string]int)
	for _, song := range p.history {
		if song.Album != "" {
			albumCount[song.Album]++
		}
	}
	stats["unique_albums"] = len(albumCount)

	// 最近播放时间
	if len(p.history) > 0 {
		lastSong := p.history[len(p.history)-1]
		if playedAt, exists := lastSong.Metadata["played_at"]; exists {
			stats["last_played_at"] = playedAt
		}
	}

	// 最常播放的艺术家
	mostPlayedArtist := ""
	maxArtistPlays := 0
	for artist, count := range artistCount {
		if count > maxArtistPlays {
			maxArtistPlays = count
			mostPlayedArtist = artist
		}
	}
	if mostPlayedArtist != "" {
		stats["most_played_artist"] = map[string]interface{}{
			"name":  mostPlayedArtist,
			"plays": maxArtistPlays,
		}
	}

	return stats, nil
}

// getUniqueSongs 获取历史中的唯一歌曲
func (p *PlaylistPluginImpl) getUniqueSongs() map[string]*model.Song {
	uniqueSongs := make(map[string]*model.Song)
	for _, song := range p.history {
		uniqueSongs[song.ID] = song
	}
	return uniqueSongs
}