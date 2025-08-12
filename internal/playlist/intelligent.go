package playlist

import (
	"strconv"

	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// IntelligentPlayMode 心动模式播放实现
// 基于当前歌曲智能推荐相似歌曲，实现智能播放
type IntelligentPlayMode struct {
	recommendedSongs []structs.Song // 推荐歌曲列表
	currentIndex     int            // 当前播放索引
	originalSong     structs.Song   // 原始触发歌曲
	playlistId       int64          // 播放列表ID（用于API调用）
}

// NewIntelligentPlayMode 创建新的心动模式实例
func NewIntelligentPlayMode() PlayMode {
	return &IntelligentPlayMode{
		currentIndex: -1,
	}
}

// NextSong 获取下一首歌曲的索引
// 心动模式下，当播放到列表末尾时，会基于当前歌曲获取更多推荐
func (i *IntelligentPlayMode) NextSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}

	// 如果当前索引无效，返回第一首
	if currentIndex < 0 || currentIndex >= len(playlist) {
		return 0, nil
	}

	// 计算下一首的索引
	nextIndex := currentIndex + 1

	// 如果超出范围，心动模式需要获取更多推荐歌曲
	if nextIndex >= len(playlist) {
		// 心动模式的特殊处理：当到达列表末尾时，返回错误
		// 实际的智能推荐逻辑由上层Player.NextSong()处理
		return -1, ErrNoNextSong
	}

	return nextIndex, nil
}

// PreviousSong 获取上一首歌曲的索引
// 心动模式下，简单递减索引
func (i *IntelligentPlayMode) PreviousSong(currentIndex int, playlist []structs.Song, manual bool) (int, error) {
	if len(playlist) == 0 {
		return -1, ErrEmptyPlaylist
	}

	// 如果当前索引无效，返回最后一首
	if currentIndex < 0 || currentIndex >= len(playlist) {
		return len(playlist) - 1, nil
	}

	// 计算上一首的索引
	prevIndex := currentIndex - 1

	// 如果小于0，心动模式下停止播放
	if prevIndex < 0 {
		return -1, ErrNoPreviousSong
	}

	return prevIndex, nil
}

// Initialize 初始化心动模式
// 心动模式需要记录原始歌曲信息，用于后续的智能推荐
func (i *IntelligentPlayMode) Initialize(currentIndex int, playlist []structs.Song) error {
	if len(playlist) == 0 {
		return ErrEmptyPlaylist
	}

	if currentIndex < 0 || currentIndex >= len(playlist) {
		return ErrInvalidIndex
	}

	// 记录当前播放的歌曲作为推荐基础
	i.originalSong = playlist[currentIndex]
	i.currentIndex = currentIndex

	return nil
}

// GetMode 获取播放模式类型
func (i *IntelligentPlayMode) GetMode() types.Mode {
	return types.PmIntelligent
}

// GetModeName 获取播放模式名称
func (i *IntelligentPlayMode) GetModeName() string {
	return "心动模式"
}

// OnPlaylistChanged 当播放列表发生变化时的回调
// 心动模式下，播放列表变化时需要更新推荐基础
func (i *IntelligentPlayMode) OnPlaylistChanged(currentIndex int, playlist []structs.Song) error {
	if len(playlist) == 0 {
		i.recommendedSongs = nil
		i.currentIndex = -1
		return nil
	}

	// 更新当前索引和原始歌曲
	if currentIndex >= 0 && currentIndex < len(playlist) {
		i.currentIndex = currentIndex
		i.originalSong = playlist[currentIndex]
	}

	return nil
}

// GetIntelligentRecommendations 获取智能推荐歌曲列表
// 这是心动模式的核心功能，基于指定歌曲获取推荐
func (i *IntelligentPlayMode) GetIntelligentRecommendations(songId, playlistId int64) ([]structs.Song, error) {
	intelligenceService := service.PlaymodeIntelligenceListService{
		SongId:       strconv.FormatInt(songId, 10),
		PlaylistId:   strconv.FormatInt(playlistId, 10),
		StartMusicId: strconv.FormatInt(songId, 10),
	}

	code, response := intelligenceService.PlaymodeIntelligenceList()
	codeType := _struct.CheckCode(code)
	if codeType != _struct.Success {
		return nil, ErrInvalidPlayMode // 使用现有错误类型
	}

	songs := _struct.GetIntelligenceSongs(response)
	return songs, nil
}

