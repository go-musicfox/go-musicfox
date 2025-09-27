package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	ui "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/ui"
	"github.com/go-musicfox/go-musicfox/v2/plugins/tui/config"
)

// TUIRenderer TUI渲染器
type TUIRenderer struct {
	config *config.TUIConfig
	theme  *TUITheme
}

// TUITheme TUI主题
type TUITheme struct {
	Primary     string
	Secondary   string
	Accent      string
	Background  string
	Text        string
	Border      string
	Highlight   string
	Error       string
	Warning     string
	Success     string
}

// NewTUIRenderer 创建TUI渲染器
func NewTUIRenderer(config *config.TUIConfig) *TUIRenderer {
	return &TUIRenderer{
		config: config,
		theme:  getDefaultTheme(),
	}
}

// Render 渲染组件
func (r *TUIRenderer) Render(ctx context.Context, component *ui.UIComponent, state *ui.AppState) ([]byte, error) {
	if component == nil {
		return nil, fmt.Errorf("component cannot be nil")
	}

	switch component.Type {
	case ui.ComponentTypePlayer:
		return r.renderPlayer(ctx, component, state)
	case ui.ComponentTypePlaylist:
		return r.renderPlaylist(ctx, component, state)
	case ui.ComponentTypeList:
		return r.renderList(ctx, component, state)
	case ui.ComponentTypeLyrics:
		return r.renderLyrics(ctx, component, state)
	case ui.ComponentTypeCustom:
		return r.renderCustom(ctx, component, state)
	default:
		return r.renderDefault(ctx, component, state)
	}
}

// SupportsType 检查是否支持指定的UI类型
func (r *TUIRenderer) SupportsType(uiType ui.UIType) bool {
	return uiType == ui.UITypeTerminal
}

// GetSupportedTypes 获取支持的UI类型
func (r *TUIRenderer) GetSupportedTypes() []ui.UIType {
	return []ui.UIType{ui.UITypeTerminal}
}

// renderPlayer 渲染播放器组件
func (r *TUIRenderer) renderPlayer(ctx context.Context, component *ui.UIComponent, state *ui.AppState) ([]byte, error) {
	if state.Player == nil {
		return []byte("No player state available"), nil
	}

	var output strings.Builder

	// 渲染当前歌曲信息
	if state.Player.CurrentSong != nil {
		song := state.Player.CurrentSong
		output.WriteString(fmt.Sprintf("♪ %s - %s\n", song.Title, song.Artist))
		if song.Album != "" {
			output.WriteString(fmt.Sprintf("  专辑: %s\n", song.Album))
		}
	} else {
		output.WriteString("♪ 暂无播放歌曲\n")
	}

	// 渲染播放状态
	status := r.getPlayStatusText(state.Player.Status)
	output.WriteString(fmt.Sprintf("状态: %s\n", status))

	// 渲染进度条
	progressBar := r.renderProgressBar(state.Player.Position, state.Player.Duration)
	output.WriteString(progressBar)
	output.WriteString("\n")

	// 渲染音量
	volumeBar := r.renderVolumeBar(state.Player.Volume)
	output.WriteString(volumeBar)
	output.WriteString("\n")

	// 渲染播放模式
	playMode := r.getPlayModeText(state.Player.PlayMode)
	output.WriteString(fmt.Sprintf("播放模式: %s\n", playMode))

	return []byte(output.String()), nil
}

// renderPlaylist 渲染播放列表组件
func (r *TUIRenderer) renderPlaylist(ctx context.Context, component *ui.UIComponent, state *ui.AppState) ([]byte, error) {
	if state.Player == nil || len(state.Player.Queue) == 0 {
		return []byte("播放列表为空"), nil
	}

	var output strings.Builder
	output.WriteString("=== 播放列表 ===\n")

	for i, song := range state.Player.Queue {
		prefix := "  "
		if state.Player.CurrentSong != nil && song.ID == state.Player.CurrentSong.ID {
			prefix = "► "
		}
		output.WriteString(fmt.Sprintf("%s%d. %s - %s\n", prefix, i+1, song.Title, song.Artist))
	}

	return []byte(output.String()), nil
}

// renderList 渲染列表组件
func (r *TUIRenderer) renderList(ctx context.Context, component *ui.UIComponent, state *ui.AppState) ([]byte, error) {
	var output strings.Builder

	// 从组件属性中获取列表数据
	if items, ok := component.Props["items"].([]interface{}); ok {
		for i, item := range items {
			if itemStr, ok := item.(string); ok {
				output.WriteString(fmt.Sprintf("%d. %s\n", i+1, itemStr))
			}
		}
	} else {
		output.WriteString("列表为空")
	}

	return []byte(output.String()), nil
}

// renderLyrics 渲染歌词组件
func (r *TUIRenderer) renderLyrics(ctx context.Context, component *ui.UIComponent, state *ui.AppState) ([]byte, error) {
	if !r.config.ShowLyrics {
		return []byte(""), nil
	}

	var output strings.Builder
	output.WriteString("=== 歌词 ===\n")

	// 这里应该实现真正的歌词渲染逻辑
	// 暂时显示占位符
	for i := 0; i < r.config.LyricLines; i++ {
		output.WriteString("♪ 歌词行 " + fmt.Sprintf("%d", i+1) + "\n")
	}

	return []byte(output.String()), nil
}

// renderCustom 渲染自定义组件
func (r *TUIRenderer) renderCustom(ctx context.Context, component *ui.UIComponent, state *ui.AppState) ([]byte, error) {
	// 执行自定义模板渲染
	if component.Template != "" {
		return r.renderTemplate(component.Template, state)
	}

	return []byte(fmt.Sprintf("Custom Component: %s", component.Name)), nil
}

// renderDefault 渲染默认组件
func (r *TUIRenderer) renderDefault(ctx context.Context, component *ui.UIComponent, state *ui.AppState) ([]byte, error) {
	return []byte(fmt.Sprintf("Component: %s (Type: %d)", component.Name, component.Type)), nil
}

// renderTemplate 渲染模板
func (r *TUIRenderer) renderTemplate(template string, state *ui.AppState) ([]byte, error) {
	// 这里应该实现模板引擎
	// 暂时返回原始模板
	return []byte(template), nil
}

// renderProgressBar 渲染进度条
func (r *TUIRenderer) renderProgressBar(position, duration time.Duration) string {
	if duration == 0 {
		return "[--:--] ━━━━━━━━━━━━━━━━━━━━ [--:--]"
	}

	progress := float64(position) / float64(duration)
	if progress > 1.0 {
		progress = 1.0
	}

	barWidth := 20
	filledWidth := int(progress * float64(barWidth))

	var bar strings.Builder
	bar.WriteString("[")
	bar.WriteString(formatDuration(position.Seconds()))
	bar.WriteString("] ")

	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar.WriteString("━")
		} else {
			bar.WriteString("─")
		}
	}

	bar.WriteString(" [")
	bar.WriteString(formatDuration(duration.Seconds()))
	bar.WriteString("]")

	return bar.String()
}

// renderVolumeBar 渲染音量条
func (r *TUIRenderer) renderVolumeBar(volume float64) string {
	if volume > 1.0 {
		volume = 1.0
	}

	barWidth := 10
	filledWidth := int(volume * float64(barWidth))

	var bar strings.Builder
	bar.WriteString("音量: [")

	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar.WriteString("■")
		} else {
			bar.WriteString("□")
		}
	}

	bar.WriteString(fmt.Sprintf("] %d%%", int(volume*100)))
	return bar.String()
}

// getPlayStatusText 获取播放状态文本
func (r *TUIRenderer) getPlayStatusText(status ui.PlayStatus) string {
	switch status {
	case ui.PlayStatusPlaying:
		return "播放中 ▶"
	case ui.PlayStatusPaused:
		return "暂停 ⏸"
	case ui.PlayStatusStopped:
		return "停止 ⏹"
	case ui.PlayStatusBuffering:
		return "缓冲中 ⏳"
	case ui.PlayStatusError:
		return "错误 ❌"
	default:
		return "未知"
	}
}

// getPlayModeText 获取播放模式文本
func (r *TUIRenderer) getPlayModeText(mode ui.PlayMode) string {
	switch mode {
	case ui.PlayModeSequential:
		return "顺序播放 →"
	case ui.PlayModeRepeatOne:
		return "单曲循环 🔂"
	case ui.PlayModeRepeatAll:
		return "列表循环 🔁"
	case ui.PlayModeShuffle:
		return "随机播放 🔀"
	default:
		return "未知"
	}
}

// formatDuration 格式化时长
func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "--:--"
	}
	
	minutes := int(seconds) / 60
	secs := int(seconds) % 60
	
	if minutes >= 60 {
		hours := minutes / 60
		minutes = minutes % 60
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

// getDefaultTheme 获取默认主题
func getDefaultTheme() *TUITheme {
	return &TUITheme{
		Primary:    "#007ACC",
		Secondary:  "#6C757D",
		Accent:     "#28A745",
		Background: "#000000",
		Text:       "#FFFFFF",
		Border:     "#6C757D",
		Highlight:  "#FFC107",
		Error:      "#DC3545",
		Warning:    "#FD7E14",
		Success:    "#28A745",
	}
}