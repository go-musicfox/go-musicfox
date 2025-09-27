package playlist

import (
	"context"
	"fmt"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
)

// SetPlayMode 设置播放模式
func (p *PlaylistPluginImpl) SetPlayMode(ctx context.Context, mode model.PlayMode) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	oldMode := p.playMode
	p.playMode = mode

	// 如果切换到随机模式，生成随机索引
	if mode == model.PlayModeShuffle && oldMode != model.PlayModeShuffle {
		p.generateShuffleIndex()
	} else if mode != model.PlayModeShuffle {
		// 如果切换出随机模式，清除随机索引
		p.shuffleIndex = nil
	}

	// 发送播放模式变化事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("play_mode_changed_%d", time.Now().UnixNano()),
			Type:      event.EventPlayerModeChanged,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"old_mode": oldMode.String(),
				"new_mode": mode.String(),
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return nil
}

// GetPlayMode 获取当前播放模式
func (p *PlaylistPluginImpl) GetPlayMode(ctx context.Context) model.PlayMode {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.playMode
}

// GetNextSong 根据播放模式获取下一首歌曲
func (p *PlaylistPluginImpl) GetNextSong(ctx context.Context, currentSong *model.Song) (*model.Song, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.currentQueue) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	// 如果没有当前歌曲，返回第一首
	if currentSong == nil {
		return p.getFirstSong(), nil
	}

	// 找到当前歌曲在队列中的位置
	currentIndex := p.getCurrentSongIndex(currentSong)
	if currentIndex == -1 {
		// 当前歌曲不在队列中，返回第一首
		return p.getFirstSong(), nil
	}

	p.currentIndex = currentIndex

	switch p.playMode {
	case model.PlayModeSequential:
		return p.getNextSequential()
	case model.PlayModeRepeatOne:
		return p.getRepeatOne()
	case model.PlayModeRepeatAll:
		return p.getNextRepeatAll()
	case model.PlayModeShuffle:
		return p.getNextShuffle()
	default:
		return p.getNextSequential()
	}
}

// GetPreviousSong 根据播放模式获取上一首歌曲
func (p *PlaylistPluginImpl) GetPreviousSong(ctx context.Context, currentSong *model.Song) (*model.Song, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.currentQueue) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	// 如果没有当前歌曲，返回最后一首
	if currentSong == nil {
		return p.getLastSong(), nil
	}

	// 找到当前歌曲在队列中的位置
	currentIndex := p.getCurrentSongIndex(currentSong)
	if currentIndex == -1 {
		// 当前歌曲不在队列中，返回最后一首
		return p.getLastSong(), nil
	}

	p.currentIndex = currentIndex

	switch p.playMode {
	case model.PlayModeSequential:
		return p.getPreviousSequential()
	case model.PlayModeRepeatOne:
		return p.getRepeatOne()
	case model.PlayModeRepeatAll:
		return p.getPreviousRepeatAll()
	case model.PlayModeShuffle:
		return p.getPreviousShuffle()
	default:
		return p.getPreviousSequential()
	}
}

// getFirstSong 获取第一首歌曲
func (p *PlaylistPluginImpl) getFirstSong() *model.Song {
	if p.playMode == model.PlayModeShuffle && len(p.shuffleIndex) > 0 {
		p.currentIndex = 0
		return p.currentQueue[p.shuffleIndex[0]]
	}
	p.currentIndex = 0
	return p.currentQueue[0]
}

// getLastSong 获取最后一首歌曲
func (p *PlaylistPluginImpl) getLastSong() *model.Song {
	if p.playMode == model.PlayModeShuffle && len(p.shuffleIndex) > 0 {
		p.currentIndex = len(p.shuffleIndex) - 1
		return p.currentQueue[p.shuffleIndex[p.currentIndex]]
	}
	p.currentIndex = len(p.currentQueue) - 1
	return p.currentQueue[p.currentIndex]
}

// getNextSequential 顺序播放模式下获取下一首
func (p *PlaylistPluginImpl) getNextSequential() (*model.Song, error) {
	if p.currentIndex >= len(p.currentQueue)-1 {
		// 已经是最后一首，顺序播放模式下停止
		return nil, fmt.Errorf("reached end of queue")
	}

	p.currentIndex++
	return p.currentQueue[p.currentIndex], nil
}

// getPreviousSequential 顺序播放模式下获取上一首
func (p *PlaylistPluginImpl) getPreviousSequential() (*model.Song, error) {
	if p.currentIndex <= 0 {
		// 已经是第一首，顺序播放模式下停止
		return nil, fmt.Errorf("reached beginning of queue")
	}

	p.currentIndex--
	return p.currentQueue[p.currentIndex], nil
}

// getRepeatOne 单曲循环模式下获取当前歌曲
func (p *PlaylistPluginImpl) getRepeatOne() (*model.Song, error) {
	if p.currentIndex < 0 || p.currentIndex >= len(p.currentQueue) {
		return nil, fmt.Errorf("invalid current index")
	}
	return p.currentQueue[p.currentIndex], nil
}

// getNextRepeatAll 列表循环模式下获取下一首
func (p *PlaylistPluginImpl) getNextRepeatAll() (*model.Song, error) {
	p.currentIndex++
	if p.currentIndex >= len(p.currentQueue) {
		p.currentIndex = 0 // 循环到第一首
	}
	return p.currentQueue[p.currentIndex], nil
}

// getPreviousRepeatAll 列表循环模式下获取上一首
func (p *PlaylistPluginImpl) getPreviousRepeatAll() (*model.Song, error) {
	p.currentIndex--
	if p.currentIndex < 0 {
		p.currentIndex = len(p.currentQueue) - 1 // 循环到最后一首
	}
	return p.currentQueue[p.currentIndex], nil
}

// getNextShuffle 随机播放模式下获取下一首
func (p *PlaylistPluginImpl) getNextShuffle() (*model.Song, error) {
	if p.shuffleIndex == nil {
		p.generateShuffleIndex()
	}

	if len(p.shuffleIndex) == 0 {
		return nil, fmt.Errorf("shuffle index is empty")
	}

	// 找到当前歌曲在随机索引中的位置
	currentShuffleIndex := p.getShuffleIndex(p.currentIndex)
	if currentShuffleIndex == -1 {
		// 当前歌曲不在随机索引中，从第一首开始
		currentShuffleIndex = 0
	} else {
		currentShuffleIndex++
		if currentShuffleIndex >= len(p.shuffleIndex) {
			// 随机播放完毕，重新生成随机索引
			p.generateShuffleIndex()
			currentShuffleIndex = 0
		}
	}

	actualIndex := p.shuffleIndex[currentShuffleIndex]
	p.currentIndex = actualIndex
	return p.currentQueue[actualIndex], nil
}

// getPreviousShuffle 随机播放模式下获取上一首
func (p *PlaylistPluginImpl) getPreviousShuffle() (*model.Song, error) {
	if p.shuffleIndex == nil {
		p.generateShuffleIndex()
	}

	if len(p.shuffleIndex) == 0 {
		return nil, fmt.Errorf("shuffle index is empty")
	}

	// 找到当前歌曲在随机索引中的位置
	currentShuffleIndex := p.getShuffleIndex(p.currentIndex)
	if currentShuffleIndex == -1 {
		// 当前歌曲不在随机索引中，从最后一首开始
		currentShuffleIndex = len(p.shuffleIndex) - 1
	} else {
		currentShuffleIndex--
		if currentShuffleIndex < 0 {
			// 已经是随机列表的第一首，循环到最后一首
			currentShuffleIndex = len(p.shuffleIndex) - 1
		}
	}

	actualIndex := p.shuffleIndex[currentShuffleIndex]
	p.currentIndex = actualIndex
	return p.currentQueue[actualIndex], nil
}

// TogglePlayMode 切换播放模式
func (p *PlaylistPluginImpl) TogglePlayMode(ctx context.Context) (model.PlayMode, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 循环切换播放模式
	switch p.playMode {
	case model.PlayModeSequential:
		p.playMode = model.PlayModeRepeatOne
	case model.PlayModeRepeatOne:
		p.playMode = model.PlayModeRepeatAll
	case model.PlayModeRepeatAll:
		p.playMode = model.PlayModeShuffle
		p.generateShuffleIndex()
	case model.PlayModeShuffle:
		p.playMode = model.PlayModeSequential
		p.shuffleIndex = nil
	default:
		p.playMode = model.PlayModeSequential
	}

	// 发送播放模式变化事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("play_mode_toggled_%d", time.Now().UnixNano()),
			Type:      event.EventPlayerModeChanged,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"new_mode": p.playMode.String(),
			},
		}
		p.eventBus.Publish(ctx, event)
	}

	return p.playMode, nil
}

// GetPlayModeDescription 获取播放模式描述
func (p *PlaylistPluginImpl) GetPlayModeDescription(ctx context.Context) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	switch p.playMode {
	case model.PlayModeSequential:
		return "顺序播放"
	case model.PlayModeRepeatOne:
		return "单曲循环"
	case model.PlayModeRepeatAll:
		return "列表循环"
	case model.PlayModeShuffle:
		return "随机播放"
	default:
		return "未知模式"
	}
}

// IsShuffleMode 检查是否为随机播放模式
func (p *PlaylistPluginImpl) IsShuffleMode() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.playMode == model.PlayModeShuffle
}

// IsRepeatMode 检查是否为循环播放模式
func (p *PlaylistPluginImpl) IsRepeatMode() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.playMode == model.PlayModeRepeatOne || p.playMode == model.PlayModeRepeatAll
}