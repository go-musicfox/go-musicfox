package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

// qualityDisplayName 返回音质的中文显示名称。
func qualityDisplayName(level service.SongQualityLevel) string {
	switch level {
	case service.Standard:
		return "标准"
	case service.Higher:
		return "较高"
	case service.Exhigh:
		return "极高"
	case service.Lossless:
		return "无损"
	case service.Hires:
		return "Hi-Res"
	case service.JYEffect:
		return "高清环绕"
	case service.Sky:
		return "沉浸环绕"
	case service.JYMaster:
		return "超清母带"
	default:
		return string(level)
	}
}

// formatQueueAndQuality 格式化状态栏中间文本。
// 格式：musicfox · [当前索引/总数] · 音质
// 若无播放歌曲或播放列表为空，返回空字符串。
func formatQueueAndQuality(player *Player) string {
	song := player.CurSong()
	if song.Id == 0 {
		return ""
	}

	playlist := player.Playlist()
	if len(playlist) == 0 {
		return ""
	}

	// 队列位置（1-indexed 显示）
	curIndex := player.CurSongIndex() + 1
	total := len(playlist)
	position := fmt.Sprintf("[%d/%d]", curIndex, total)

	// 音质
	quality := configs.AppConfig.Player.SongLevel
	qualityName := qualityDisplayName(quality)

	return fmt.Sprintf("%s · %s · %s", util.SetFgStyle("musicfox", util.GetPrimaryColor()), position, qualityName)
}

// NewQueueQualityStatusBar 创建带队列位置与音质中间文本的 DefaultStatusBar。
func NewQueueQualityStatusBar(player *Player) *model.DefaultStatusBar {
	return &model.DefaultStatusBar{
		Center: func(a *model.App, m *model.Main) string {
			return formatQueueAndQuality(player)
		},
	}
}
