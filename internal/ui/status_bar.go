package ui

import (
	"fmt"
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	"github.com/go-musicfox/netease-music/service"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

const musicfoxStatusBarLabel = "musicfox"

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

	statusTextStyle := style.CurrentStyleSet().StatusBarText
	return statusTextStyle.Foreground(util.GetPrimaryColor()).Render(musicfoxStatusBarLabel) +
		statusTextStyle.Render(fmt.Sprintf(" · %s · %s", position, qualityName))
}

type queueQualityStatusBarComponent struct {
	player  *Player
	openURL func(string) error
}

func (c *queueQualityStatusBarComponent) View(*model.App, *model.Main) string {
	return formatQueueAndQuality(c.player)
}

func (c *queueQualityStatusBarComponent) HandleMouse(mouse tea.Mouse, x, _ int) (bool, tea.Cmd) {
	if mouse.Button != tea.MouseLeft || !c.IsMouseOver(x, 0) {
		return false, nil
	}
	return true, func() tea.Msg {
		if err := c.openURL(types.AppGithubUrl); err != nil {
			slog.Error("Failed to open musicfox GitHub repository", "url", types.AppGithubUrl, "err", err)
		}
		return nil
	}
}

func (c *queueQualityStatusBarComponent) IsMouseOver(x, _ int) bool {
	return x >= 0 && x < len(musicfoxStatusBarLabel)
}

// NewQueueQualityStatusBar 创建带队列位置与音质中间文本的 DefaultStatusBar。
func NewQueueQualityStatusBar(player *Player) *model.DefaultStatusBar {
	return newQueueQualityStatusBar(player, open.Start)
}

func newQueueQualityStatusBar(player *Player, openURL func(string) error) *model.DefaultStatusBar {
	return &model.DefaultStatusBar{
		Components: []model.StatusBarComponent{
			&queueQualityStatusBarComponent{player: player, openURL: openURL},
		},
	}
}
