package playlist

import (
	"errors"
	"fmt"
)

// PlaylistError 播放列表相关错误的自定义类型
type PlaylistError struct {
	Op  string // 操作名称
	Err error  // 底层错误
}

func (e *PlaylistError) Error() string {
	if e.Op == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("playlist %s: %v", e.Op, e.Err)
}

func (e *PlaylistError) Unwrap() error {
	return e.Err
}

// 标准错误变量
var (
	// ErrEmptyPlaylist 播放列表为空
	ErrEmptyPlaylist = errors.New("playlist is empty")
	
	// ErrInvalidIndex 无效的索引
	ErrInvalidIndex = errors.New("invalid index")
	
	// ErrInvalidPlayMode 无效的播放模式
	ErrInvalidPlayMode = errors.New("invalid play mode")
	
	// ErrNoCurrentSong 没有当前播放的歌曲
	ErrNoCurrentSong = errors.New("no current song")
	
	// ErrNoNextSong 没有下一首歌曲
	ErrNoNextSong = errors.New("no next song")
	
	// ErrNoPreviousSong 没有上一首歌曲
	ErrNoPreviousSong = errors.New("no previous song")
	
	// ErrIndexOutOfRange 索引超出范围
	ErrIndexOutOfRange = errors.New("index out of range")
)

// newPlaylistError 创建一个新的播放列表错误
func newPlaylistError(op string, err error) error {
	return &PlaylistError{
		Op:  op,
		Err: err,
	}
}