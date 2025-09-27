package playlist

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
)

// SetCurrentQueue 设置当前播放队列
func (p *PlaylistPluginImpl) SetCurrentQueue(ctx context.Context, songs []*model.Song) error {
	if songs == nil {
		return fmt.Errorf("songs cannot be nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 设置新的队列
	p.currentQueue = make([]*model.Song, len(songs))
	copy(p.currentQueue, songs)
	p.currentIndex = -1
	p.shuffleIndex = nil

	// 如果是随机模式，生成随机索引
	if p.playMode == model.PlayModeShuffle {
		p.generateShuffleIndex()
	}

	// 发送队列更新事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("queue_updated_%d", time.Now().UnixNano()),
			Type:      "queue.updated",
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"queue_length": len(p.currentQueue),
				"play_mode":    p.playMode.String(),
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// GetCurrentQueue 获取当前播放队列
func (p *PlaylistPluginImpl) GetCurrentQueue(ctx context.Context) ([]*model.Song, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 返回队列的副本
	queue := make([]*model.Song, len(p.currentQueue))
	copy(queue, p.currentQueue)

	return queue, nil
}

// AddToQueue 添加歌曲到播放队列
func (p *PlaylistPluginImpl) AddToQueue(ctx context.Context, song *model.Song) error {
	if song == nil {
		return fmt.Errorf("song cannot be nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查歌曲是否已在队列中
	for _, existingSong := range p.currentQueue {
		if existingSong.ID == song.ID {
			return fmt.Errorf("song already in queue: %s", song.ID)
		}
	}

	// 添加歌曲到队列
	p.currentQueue = append(p.currentQueue, song)

	// 如果是随机模式，重新生成随机索引
	if p.playMode == model.PlayModeShuffle {
		p.generateShuffleIndex()
	}

	// 发送添加到队列事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("queue_song_added_%d", time.Now().UnixNano()),
			Type:      "queue.song_added",
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"song_id":      song.ID,
				"song_title":   song.Title,
				"song_artist":  song.Artist,
				"queue_length": len(p.currentQueue),
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// RemoveFromQueue 从播放队列移除歌曲
func (p *PlaylistPluginImpl) RemoveFromQueue(ctx context.Context, songID string) error {
	if songID == "" {
		return fmt.Errorf("song ID cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 找到歌曲在队列中的位置
	index := -1
	for i, song := range p.currentQueue {
		if song.ID == songID {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("song not found in queue: %s", songID)
	}

	// 移除歌曲
	p.currentQueue = append(p.currentQueue[:index], p.currentQueue[index+1:]...)

	// 调整当前索引
	if p.currentIndex > index {
		p.currentIndex--
	} else if p.currentIndex == index {
		// 如果移除的是当前播放的歌曲，重置索引
		p.currentIndex = -1
	}

	// 如果是随机模式，重新生成随机索引
	if p.playMode == model.PlayModeShuffle {
		p.generateShuffleIndex()
	}

	// 发送从队列移除事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("queue_song_removed_%d", time.Now().UnixNano()),
			Type:      "queue.song_removed",
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"song_id":      songID,
				"queue_length": len(p.currentQueue),
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// ClearQueue 清空播放队列
func (p *PlaylistPluginImpl) ClearQueue(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 清空队列
	p.currentQueue = make([]*model.Song, 0)
	p.currentIndex = -1
	p.shuffleIndex = nil

	// 发送队列清空事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("queue_cleared_%d", time.Now().UnixNano()),
			Type:      "queue.cleared",
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"queue_length": 0,
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// ShuffleQueue 随机打乱播放队列
func (p *PlaylistPluginImpl) ShuffleQueue(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.currentQueue) <= 1 {
		return nil // 队列太短，无需打乱
	}

	// 使用Fisher-Yates算法打乱队列
	rand.Seed(time.Now().UnixNano())
	for i := len(p.currentQueue) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		p.currentQueue[i], p.currentQueue[j] = p.currentQueue[j], p.currentQueue[i]
	}

	// 重置当前索引
	p.currentIndex = -1
	p.shuffleIndex = nil

	// 发送队列打乱事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("queue_shuffled_%d", time.Now().UnixNano()),
			Type:      "queue.shuffled",
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"queue_length": len(p.currentQueue),
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// generateShuffleIndex 生成随机播放索引
func (p *PlaylistPluginImpl) generateShuffleIndex() {
	if len(p.currentQueue) == 0 {
		p.shuffleIndex = nil
		return
	}

	// 创建索引数组
	p.shuffleIndex = make([]int, len(p.currentQueue))
	for i := range p.shuffleIndex {
		p.shuffleIndex[i] = i
	}

	// 打乱索引
	rand.Seed(time.Now().UnixNano())
	for i := len(p.shuffleIndex) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		p.shuffleIndex[i], p.shuffleIndex[j] = p.shuffleIndex[j], p.shuffleIndex[i]
	}
}

// getCurrentSongIndex 获取当前歌曲在队列中的实际索引
func (p *PlaylistPluginImpl) getCurrentSongIndex(currentSong *model.Song) int {
	if currentSong == nil {
		return -1
	}

	for i, song := range p.currentQueue {
		if song.ID == currentSong.ID {
			return i
		}
	}

	return -1
}

// getShuffleIndex 获取随机播放模式下的索引
func (p *PlaylistPluginImpl) getShuffleIndex(actualIndex int) int {
	if p.shuffleIndex == nil || actualIndex < 0 || actualIndex >= len(p.shuffleIndex) {
		return -1
	}

	for i, idx := range p.shuffleIndex {
		if idx == actualIndex {
			return i
		}
	}

	return -1
}

// getActualIndex 从随机索引获取实际索引
func (p *PlaylistPluginImpl) getActualIndex(shuffleIndex int) int {
	if p.shuffleIndex == nil || shuffleIndex < 0 || shuffleIndex >= len(p.shuffleIndex) {
		return -1
	}

	return p.shuffleIndex[shuffleIndex]
}