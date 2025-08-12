package playlist

import (
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// PlaylistManager 播放列表管理器接口
// 提供播放列表的核心管理功能，包括播放控制、模式切换等
type PlaylistManager interface {
	// Initialize 初始化播放列表和当前播放索引
	Initialize(index int, playlist []structs.Song) error
	
	// GetPlaylist 获取当前播放列表
	GetPlaylist() []structs.Song
	
	// GetCurrentIndex 获取当前播放歌曲的索引
	GetCurrentIndex() int
	
	// GetCurrentSong 获取当前播放的歌曲
	GetCurrentSong() (structs.Song, error)
	
	// NextSong 切换到下一首歌曲
	// manual 参数表示是否为手动切换
	NextSong(manual bool) (structs.Song, error)
	
	// PreviousSong 切换到上一首歌曲
	// manual 参数表示是否为手动切换
	PreviousSong(manual bool) (structs.Song, error)
	
	// RemoveSong 从播放列表中移除指定索引的歌曲
	// 返回移除后应该播放的歌曲（如果有的话）
	RemoveSong(index int) (structs.Song, error)
	
	// SetPlayMode 设置播放模式
	SetPlayMode(mode types.Mode) error
	
	// GetPlayMode 获取当前播放模式
	GetPlayMode() types.Mode
	
	// GetPlayModeName 获取当前播放模式的名称
	GetPlayModeName() string
	
	// SaveState 保存播放列表状态到存储
	SaveState() error
	
	// LoadState 从存储加载播放列表状态
	LoadState() error
}

// PlayMode 播放模式策略接口
// 定义不同播放模式的行为策略
type PlayMode interface {
	// NextSong 根据当前播放模式获取下一首歌曲
	// manual 参数表示是否为手动切换
	NextSong(currentIndex int, playlist []structs.Song, manual bool) (int, error)
	
	// PreviousSong 根据当前播放模式获取上一首歌曲
	// manual 参数表示是否为手动切换
	PreviousSong(currentIndex int, playlist []structs.Song, manual bool) (int, error)
	
	// Initialize 初始化播放模式（用于需要特殊初始化的模式，如随机播放）
	Initialize(currentIndex int, playlist []structs.Song) error
	
	// GetMode 获取播放模式类型
	GetMode() types.Mode
	
	// GetModeName 获取播放模式名称
	GetModeName() string
	
	// OnPlaylistChanged 当播放列表发生变化时的回调
	// 用于处理播放列表变化对播放模式的影响（如随机播放需要重新洗牌）
	OnPlaylistChanged(currentIndex int, playlist []structs.Song) error
}