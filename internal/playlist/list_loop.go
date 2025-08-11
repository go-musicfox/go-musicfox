package playlist

import (
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// ListLoopPlayMode 列表循环播放模式实现
// 播放列表循环播放，到达末尾时回到开头，到达开头时跳到末尾
type ListLoopPlayMode struct{}

// NewListLoopPlayMode 创建新的列表循环播放模式实例
func NewListLoopPlayMode() PlayMode {
	return &ListLoopPlayMode{}
}

// NextSong 获取下一首歌曲的索引
// 列表循环模式下，到达末尾时回到开头
func (l *ListLoopPlayMode) NextSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}
	
	// 如果当前索引无效，返回第一首
	if currentIndex < 0 || currentIndex >= len(playlist) {
		return 0, nil
	}
	
	// 计算下一首的索引，循环到开头
	nextIndex := (currentIndex + 1) % len(playlist)
	
	return nextIndex, nil
}

// PreviousSong 获取上一首歌曲的索引
// 列表循环模式下，到达开头时跳到末尾
func (l *ListLoopPlayMode) PreviousSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}
	
	// 如果当前索引无效，返回最后一首
	if currentIndex < 0 || currentIndex >= len(playlist) {
		return len(playlist) - 1, nil
	}
	
	// 计算上一首的索引，循环到末尾
	prevIndex := currentIndex - 1
	if prevIndex < 0 {
		prevIndex = len(playlist) - 1
	}
	
	return prevIndex, nil
}

// Initialize 初始化播放模式
// 列表循环播放模式无需特殊初始化
func (l *ListLoopPlayMode) Initialize(currentIndex int, playlist []structs.Song) error {
	// 列表循环播放模式无需特殊初始化逻辑
	return nil
}

// GetMode 获取播放模式类型
func (l *ListLoopPlayMode) GetMode() types.Mode {
	return types.PmListLoop
}

// GetModeName 获取播放模式名称
func (l *ListLoopPlayMode) GetModeName() string {
	return "列表循环"
}

// OnPlaylistChanged 当播放列表发生变化时的处理
// 列表循环播放模式无需特殊处理
func (l *ListLoopPlayMode) OnPlaylistChanged(currentIndex int, playlist []structs.Song) error {
	// 列表循环播放模式无需特殊处理播放列表变化
	return nil
}