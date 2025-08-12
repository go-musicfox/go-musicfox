package playlist

import (
	"math/rand"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// InfiniteRandomPlayMode 无限随机播放模式实现
// 真正随机播放，避免重复，维护播放历史
type InfiniteRandomPlayMode struct {
	history    []int     // 播放历史记录
	currentPos int       // 当前在历史中的位置
	maxHistory int       // 最大历史记录数量
	rng        *rand.Rand
}

// NewInfiniteRandomPlayMode 创建新的无限随机播放模式实例
func NewInfiniteRandomPlayMode() PlayMode {
	return &InfiniteRandomPlayMode{
		maxHistory: 100, // 默认保留100首歌的历史
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NextSong 获取下一首歌曲的索引
// 无限随机播放模式下，真正随机选择，避免重复
func (i *InfiniteRandomPlayMode) NextSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}

	// 如果只有一首歌，直接返回
	if len(playlist) == 1 {
		return 0, nil
	}

	// 如果当前位置不在历史末尾，说明用户在历史中导航
	if i.currentPos < len(i.history)-1 {
		i.currentPos++
		return i.history[i.currentPos], nil
	}

	// 生成新的随机索引
	nextIndex := i.generateRandomIndex(playlist, currentIndex)
	
	// 添加到历史记录
	i.addToHistory(nextIndex)

	return nextIndex, nil
}

// PreviousSong 获取上一首歌曲的索引
// 无限随机播放模式下，基于历史记录返回上一首
func (i *InfiniteRandomPlayMode) PreviousSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}

	// 如果没有历史记录或已经在第一首，无法返回上一首
	if len(i.history) == 0 || i.currentPos <= 0 {
		return -1, ErrNoPreviousSong
	}

	// 移动到历史中的上一个位置
	i.currentPos--
	return i.history[i.currentPos], nil
}

// Initialize 初始化播放模式
func (i *InfiniteRandomPlayMode) Initialize(currentIndex int, playlist []structs.Song) error {
	if len(playlist) == 0 {
		return ErrEmptyPlaylist
	}

	// 初始化历史记录
	if currentIndex >= 0 && currentIndex < len(playlist) {
		i.history = []int{currentIndex}
		i.currentPos = 0
	} else {
		i.history = nil
		i.currentPos = -1
	}

	return nil
}

// GetMode 获取播放模式类型
func (i *InfiniteRandomPlayMode) GetMode() types.Mode {
	return types.PmInfRandom
}

// GetModeName 获取播放模式名称
func (i *InfiniteRandomPlayMode) GetModeName() string {
	return "无限随机"
}

// OnPlaylistChanged 当播放列表发生变化时调用
func (i *InfiniteRandomPlayMode) OnPlaylistChanged(currentIndex int, playlist []structs.Song) error {
	if len(playlist) == 0 {
		i.history = nil
		i.currentPos = -1
		return nil
	}

	// 清理无效的历史记录（索引超出新播放列表范围的）
	validHistory := make([]int, 0, len(i.history))
	for _, index := range i.history {
		if index >= 0 && index < len(playlist) {
			validHistory = append(validHistory, index)
		}
	}

	i.history = validHistory

	// 调整当前位置
	if i.currentPos >= len(i.history) {
		i.currentPos = len(i.history) - 1
	}
	if i.currentPos < 0 && len(i.history) > 0 {
		i.currentPos = 0
	}

	// 如果当前索引有效且不在历史中，添加到历史
	// 但只有在历史记录为空或者当前索引不在历史记录中时才添加
	if currentIndex >= 0 && currentIndex < len(playlist) {
		found := false
		for _, historyIndex := range i.history {
			if historyIndex == currentIndex {
				found = true
				break
			}
		}
		if !found {
			i.addToHistory(currentIndex)
		}
	}

	return nil
}

// generateRandomIndex 生成随机索引，避免与当前索引相同
func (i *InfiniteRandomPlayMode) generateRandomIndex(playlist []structs.Song, currentIndex int) int {
	playlistLen := len(playlist)
	
	// 如果只有一首歌，返回该索引
	if playlistLen == 1 {
		return 0
	}

	// 计算最近播放的歌曲数量（用于避免重复）
	recentCount := min(playlistLen/3, 10) // 避免最近播放的1/3或最多10首歌
	if recentCount < 1 {
		recentCount = 1
	}

	// 获取最近播放的歌曲索引
	recentPlayed := make(map[int]bool)
	historyStart := max(0, len(i.history)-recentCount)
	for j := historyStart; j < len(i.history); j++ {
		recentPlayed[i.history[j]] = true
	}

	// 如果当前索引有效，也加入避免列表
	if currentIndex >= 0 && currentIndex < playlistLen {
		recentPlayed[currentIndex] = true
	}

	// 生成候选索引列表（排除最近播放的）
	candidates := make([]int, 0, playlistLen)
	for idx := 0; idx < playlistLen; idx++ {
		if !recentPlayed[idx] {
			candidates = append(candidates, idx)
		}
	}

	// 如果没有候选项（所有歌曲都在最近播放中），从全部歌曲中选择
	if len(candidates) == 0 {
		candidates = make([]int, 0, playlistLen)
		for idx := 0; idx < playlistLen; idx++ {
			if idx != currentIndex {
				candidates = append(candidates, idx)
			}
		}
	}

	// 如果还是没有候选项（只有一首歌且是当前歌曲），返回当前索引
	if len(candidates) == 0 {
		return currentIndex
	}

	// 随机选择一个候选索引
	return candidates[i.rng.Intn(len(candidates))]
}

// addToHistory 添加索引到历史记录
func (i *InfiniteRandomPlayMode) addToHistory(index int) {
	i.history = append(i.history, index)
	i.currentPos = len(i.history) - 1

	// 限制历史记录长度
	if len(i.history) > i.maxHistory {
		i.history = i.history[1:]
		i.currentPos--
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}