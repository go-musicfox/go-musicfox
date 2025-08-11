package playlist

import (
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// SingleLoopPlayMode 单曲循环播放模式实现
// 自动播放时重复当前歌曲，手动切换时正常切换
type SingleLoopPlayMode struct{}

// NewSingleLoopPlayMode 创建新的单曲循环播放模式实例
func NewSingleLoopPlayMode() PlayMode {
	return &SingleLoopPlayMode{}
}

// NextSong 获取下一首歌曲的索引
// 单曲循环模式下，自动播放时返回当前索引，手动切换时正常切换
func (s *SingleLoopPlayMode) NextSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}
	
	// 如果当前索引无效，返回第一首
	if currentIndex < 0 || currentIndex >= len(playlist) {
		return 0, nil
	}
	
	// 如果是手动切换，正常切换到下一首（循环到开头）
	if manual {
		nextIndex := (currentIndex + 1) % len(playlist)
		return nextIndex, nil
	}
	
	// 自动播放时，重复当前歌曲
	return currentIndex, nil
}

// PreviousSong 获取上一首歌曲的索引
// 单曲循环模式下，自动播放时返回当前索引，手动切换时正常切换
func (s *SingleLoopPlayMode) PreviousSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}
	
	// 如果当前索引无效，返回最后一首
	if currentIndex < 0 || currentIndex >= len(playlist) {
		return len(playlist) - 1, nil
	}
	
	// 如果是手动切换，正常切换到上一首（循环到末尾）
	if manual {
		prevIndex := currentIndex - 1
		if prevIndex < 0 {
			prevIndex = len(playlist) - 1
		}
		return prevIndex, nil
	}
	
	// 自动播放时，重复当前歌曲
	return currentIndex, nil
}

// Initialize 初始化播放模式
// 单曲循环播放模式无需特殊初始化
func (s *SingleLoopPlayMode) Initialize(currentIndex int, playlist []structs.Song) error {
	// 单曲循环播放模式无需特殊初始化逻辑
	return nil
}

// GetMode 获取播放模式类型
func (s *SingleLoopPlayMode) GetMode() types.Mode {
	return types.PmSingleLoop
}

// GetModeName 获取播放模式名称
func (s *SingleLoopPlayMode) GetModeName() string {
	return "单曲循环"
}

// OnPlaylistChanged 当播放列表发生变化时的处理
// 单曲循环播放模式无需特殊处理
func (s *SingleLoopPlayMode) OnPlaylistChanged(currentIndex int, playlist []structs.Song) error {
	// 单曲循环播放模式无需特殊处理播放列表变化
	return nil
}