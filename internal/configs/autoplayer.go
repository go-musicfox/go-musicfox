package configs

import (
	"strings"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

// AutoPlayerPlaylist 自动播放列表
type AutoPlayerPlaylist string

const (
	AutoPlayerPlaylistDailyReco AutoPlayerPlaylist = "dailyReco" // 每日推荐
	AutoPlayerPlaylistLike      AutoPlayerPlaylist = "like"      // 我喜欢的音乐
	AutoPlayerPlaylistNo        AutoPlayerPlaylist = "no"        // 保持上次退出时的设置，无视offset
	AutoPlayerPlaylistName      AutoPlayerPlaylist = "name:%s"   // 指定歌单名 name:[歌单名]
)

func (p AutoPlayerPlaylist) SpecialPlaylist() string {
	if p == AutoPlayerPlaylistName {
		return strings.TrimPrefix(string(p), "name:")
	}
	return ""
}

var autoPlayerModeMap = map[string]types.Mode{
	"listLoop":    types.PmListLoop,
	"order":       types.PmOrdered,
	"singleLoop":  types.PmSingleLoop,
	"random":      types.PmInfRandom,
	"infRandom":   types.PmInfRandom,
	"listRandom":  types.PmListRandom,
	"intelligent": types.PmIntelligent,
}

func PlayerModeFromAutoPlayModeString(mode string) types.Mode {
	if m, ok := autoPlayerModeMap[mode]; ok {
		return m
	}
	return types.PmUnknown
}

type AutoPlayerOptions struct {
	Enable   bool // 是否自动开始播放
	Playlist AutoPlayerPlaylist
	Offset   int // 播放偏移：0为歌单第一项，-1为歌单最后一项
	Mode     types.Mode
}

func AutoPlayerPlaylistFromString(playlist string) AutoPlayerPlaylist {
	if strings.HasPrefix(playlist, "name:") {
		return AutoPlayerPlaylistName
	}
	switch playlist {
	case "dailyReco":
		return AutoPlayerPlaylistDailyReco
	case "like":
		return AutoPlayerPlaylistLike
	case "no":
		return AutoPlayerPlaylistNo
	default:
		return AutoPlayerPlaylistDailyReco
	}
}

// AutoplayConfig 启动时自动播放相关配置
type AutoplayConfig struct {
	// 是否自动开始播放
	Enable bool `koanf:"enable"`
	// 自动播放列表
	Playlist AutoPlayerPlaylist `koanf:"playlist"`
	// 播放偏移：0为歌单第一项，-1为歌单最后一项
	Offset int `koanf:"offset"`
	// 播放模式
	Mode types.Mode `koanf:"mode"`
}
