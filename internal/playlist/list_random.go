package playlist

import (
	"math/rand"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// ListRandomPlayMode 列表随机播放模式实现
// 生成随机播放顺序，按照随机序列播放完整个列表
type ListRandomPlayMode struct {
	randomOrder []int // 随机播放顺序
	currentPos  int   // 当前在随机序列中的位置
	rng         *rand.Rand
}

// NewListRandomPlayMode 创建新的列表随机播放模式实例
func NewListRandomPlayMode() PlayMode {
	return &ListRandomPlayMode{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NextSong 获取下一首歌曲的索引
// 列表随机播放模式下，按照随机序列播放
func (l *ListRandomPlayMode) NextSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}

	// 如果随机序列为空或长度不匹配，重新生成
	if len(l.randomOrder) != len(playlist) {
		l.generateRandomOrder(len(playlist))
		l.currentPos = l.findCurrentPosition(currentIndex)
	}

	// 移动到下一个位置
	l.currentPos++

	// 如果超出范围，列表随机播放模式下停止播放
	if l.currentPos >= len(l.randomOrder) {
		return -1, ErrNoNextSong
	}

	return l.randomOrder[l.currentPos], nil
}

// PreviousSong 获取上一首歌曲的索引
// 列表随机播放模式下，按照随机序列反向播放
func (l *ListRandomPlayMode) PreviousSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}

	// 如果随机序列为空或长度不匹配，重新生成
	if len(l.randomOrder) != len(playlist) {
		l.generateRandomOrder(len(playlist))
		l.currentPos = l.findCurrentPosition(currentIndex)
	}

	// 移动到上一个位置
	l.currentPos--

	// 如果超出范围，列表随机播放模式下停止播放
	if l.currentPos < 0 {
		return -1, ErrNoPreviousSong
	}

	return l.randomOrder[l.currentPos], nil
}

// Initialize 初始化播放模式
func (l *ListRandomPlayMode) Initialize(currentIndex int, playlist []structs.Song) error {
	if len(playlist) == 0 {
		return ErrEmptyPlaylist
	}

	// 生成随机播放顺序
	l.generateRandomOrder(len(playlist))
	
	// 找到当前歌曲在随机序列中的位置
	l.currentPos = l.findCurrentPosition(currentIndex)

	return nil
}

// GetMode 获取播放模式类型
func (l *ListRandomPlayMode) GetMode() types.Mode {
	return types.PmListRandom
}

// GetModeName 获取播放模式名称
func (l *ListRandomPlayMode) GetModeName() string {
	return "列表随机"
}

// OnPlaylistChanged 当播放列表发生变化时调用
func (l *ListRandomPlayMode) OnPlaylistChanged(currentIndex int, playlist []structs.Song) error {
	if len(playlist) == 0 {
		l.randomOrder = nil
		l.currentPos = -1
		return nil
	}

	// 重新生成随机播放顺序
	l.generateRandomOrder(len(playlist))
	
	// 找到当前歌曲在新的随机序列中的位置
	l.currentPos = l.findCurrentPosition(currentIndex)

	return nil
}

// generateRandomOrder 生成随机播放顺序
// 使用Fisher-Yates洗牌算法
func (l *ListRandomPlayMode) generateRandomOrder(length int) {
	l.randomOrder = make([]int, length)
	
	// 初始化顺序数组
	for i := 0; i < length; i++ {
		l.randomOrder[i] = i
	}

	// Fisher-Yates洗牌算法
	for i := length - 1; i > 0; i-- {
		j := l.rng.Intn(i + 1)
		l.randomOrder[i], l.randomOrder[j] = l.randomOrder[j], l.randomOrder[i]
	}
}

// findCurrentPosition 在随机序列中找到当前索引的位置
func (l *ListRandomPlayMode) findCurrentPosition(currentIndex int) int {
	for pos, index := range l.randomOrder {
		if index == currentIndex {
			return pos
		}
	}
	// 如果没找到，返回第一个位置
	return 0
}