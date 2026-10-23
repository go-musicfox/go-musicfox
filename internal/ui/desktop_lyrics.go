package ui

import (
	"fmt"
	"log/slog"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/utils/app"
)

func (n *Netease) toggleDesktopLyricsHandle() {
	spec := model.NotificationSpec{Title: "桌面歌词", Level: model.NotificationInfo}
	if !desktop_lyrics.DesktopLyricsAvailable {
		spec.Message = "桌面歌词仅支持 macOS"
	} else if enabled, err := n.toggleDesktopLyrics(app.ConfigFilePath(), desktop_lyrics.NewController); err != nil {
		slog.Error("Failed to toggle desktop lyrics", "err", err)
		spec.Level = model.NotificationError
		spec.Message = err.Error()
	} else if enabled {
		n.player.updateDesktopLyrics()
		spec.Message = "已开启桌面歌词"
	} else {
		spec.Message = "已关闭桌面歌词"
	}
	n.Notify(spec)
}

// toggleDesktopLyrics retains the hidden window so toggling preserves its position.
// Window creation is supplied by the platform; playback updates share this lock.
func (n *Netease) toggleDesktopLyrics(path string, createController func(configs.DesktopLyricsConfig) desktop_lyrics.Controller) (bool, error) {
	n.desktopLyricsMu.Lock()
	defer n.desktopLyricsMu.Unlock()

	cfg := &configs.AppConfig.Main.Lyric.DesktopLyrics
	enabled := !cfg.Enable
	if enabled && n.desktopLyrics == nil {
		windowConfig := *cfg
		windowConfig.Enable = true
		n.desktopLyrics = createController(windowConfig)
		if n.desktopLyrics == nil {
			return cfg.Enable, fmt.Errorf("无法创建桌面歌词窗口，状态未改变")
		}
	}
	if err := configs.SetTOMLValue(path, []string{"main", "lyric", "desktopLyrics", "enable"}, enabled); err != nil {
		return cfg.Enable, fmt.Errorf("无法保存桌面歌词配置，状态未改变: %w", err)
	}
	cfg.Enable = enabled
	if !enabled && n.desktopLyrics != nil {
		n.desktopLyrics.Hide()
	}
	return enabled, nil
}
