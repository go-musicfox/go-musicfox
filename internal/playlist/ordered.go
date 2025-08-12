package playlist

import (
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// OrderedPlayMode 顺序播放模式实现
// 按照播放列表的顺序依次播放，不循环
type OrderedPlayMode struct{}

// NewOrderedPlayMode 创建新的顺序播放模式实例
func NewOrderedPlayMode() PlayMode {
	return &OrderedPlayMode{}
}

// NextSong 获取下一首歌曲的索引
// 顺序播放模式下，简单递增索引，到达末尾时返回错误
func (o *OrderedPlayMode) NextSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}
	
	// 如果当前索引无效，返回第一首
	if currentIndex < 0 || currentIndex >= len(playlist) {
		return 0, nil
	}
	
	// 计算下一首的索引
	nextIndex := currentIndex + 1
	
	// 如果超出范围，顺序播放模式下停止播放
	if nextIndex >= len(playlist) {
		return -1, ErrNoNextSong
	}
	
	return nextIndex, nil
}

// PreviousSong 获取上一首歌曲的索引
// 顺序播放模式下，简单递减索引，到达开头时返回错误
func (o *OrderedPlayMode) PreviousSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}
	
	// 如果当前索引无效，返回最后一首
	if currentIndex < 0 || currentIndex >= len(playlist) {
		return len(playlist) - 1, nil
	}
	
	// 计算上一首的索引
	prevIndex := currentIndex - 1
	
	// 如果小于0，顺序播放模式下停止播放
	if prevIndex < 0 {
		return -1, ErrNoPreviousSong
	}
	
	return prevIndex, nil
}

// Initialize 初始化播放模式
// 顺序播放模式无需特殊初始化
func (o *OrderedPlayMode) Initialize(currentIndex int, playlist []structs.Song) error {
	// 顺序播放模式无需特殊初始化逻辑
	return nil
}

// GetMode 获取播放模式类型
func (o *OrderedPlayMode) GetMode() types.Mode {
	return types.PmOrdered
}

// GetModeName 获取播放模式名称
func (o *OrderedPlayMode) GetModeName() string {
	return "顺序播放"
}

// OnPlaylistChanged 当播放列表发生变化时的处理
// 顺序播放模式无需特殊处理
func (o *OrderedPlayMode) OnPlaylistChanged(currentIndex int, playlist []structs.Song) error {
	// 顺序播放模式无需特殊处理播放列表变化
	return nil
}