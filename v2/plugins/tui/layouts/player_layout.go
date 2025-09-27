package layouts

import (
	"fmt"
	"strings"
	"time"

	ui "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/ui"
)

// PlayerLayout 播放器界面布局
type PlayerLayout struct {
	width  int
	height int
	theme  *ui.Theme
}

// NewPlayerLayout 创建播放器界面布局
func NewPlayerLayout(width, height int, theme *ui.Theme) *PlayerLayout {
	return &PlayerLayout{
		width:  width,
		height: height,
		theme:  theme,
	}
}

// Render 渲染播放器界面布局
func (l *PlayerLayout) Render(state *ui.AppState) ([]string, error) {
	if state == nil {
		return nil, fmt.Errorf("app state is nil")
	}

	lines := make([]string, l.height)
	currentY := 0
	
	// 渲染标题栏
	titleLines := l.renderTitle()
	copy(lines[currentY:currentY+len(titleLines)], titleLines)
	currentY += len(titleLines)
	
	// 渲染歌曲信息
	songInfoLines := l.renderSongInfo(state)
	if currentY+len(songInfoLines) <= l.height {
		copy(lines[currentY:currentY+len(songInfoLines)], songInfoLines)
		currentY += len(songInfoLines)
	}
	
	// 渲染播放控制
	controlLines := l.renderPlaybackControls(state)
	if currentY+len(controlLines) <= l.height {
		copy(lines[currentY:currentY+len(controlLines)], controlLines)
		currentY += len(controlLines)
	}
	
	// 渲染进度条
	progressLines := l.renderProgressBar(state)
	if currentY+len(progressLines) <= l.height {
		copy(lines[currentY:currentY+len(progressLines)], progressLines)
		currentY += len(progressLines)
	}
	
	// 渲染音量控制
	volumeLines := l.renderVolumeControl(state)
	if currentY+len(volumeLines) <= l.height {
		copy(lines[currentY:currentY+len(volumeLines)], volumeLines)
		currentY += len(volumeLines)
	}
	
	// 渲染歌词
	lyricsLines := l.renderLyrics(state)
	lyricsHeight := l.height - currentY - 3 // 减去状态栏高度
	if lyricsHeight > 0 && len(lyricsLines) > 0 {
		if len(lyricsLines) > lyricsHeight {
			lyricsLines = lyricsLines[:lyricsHeight]
		}
		copy(lines[currentY:currentY+len(lyricsLines)], lyricsLines)
		currentY += len(lyricsLines)
	}
	
	// 渲染状态栏
	statusLines := l.renderStatusBar(state)
	statusStartY := l.height - len(statusLines)
	if statusStartY > currentY {
		copy(lines[statusStartY:], statusLines)
	}
	
	// 填充空行
	for i := range lines {
		if lines[i] == "" {
			lines[i] = strings.Repeat(" ", l.width)
		}
	}
	
	return lines, nil
}

// renderTitle 渲染标题栏
func (l *PlayerLayout) renderTitle() []string {
	title := "播放器"
	
	// 居中显示标题
	titlePadding := (l.width - len(title)) / 2
	
	lines := []string{
		l.renderBorder("top"),
		l.padLine(strings.Repeat(" ", titlePadding) + title),
		l.renderBorder("middle"),
	}
	
	return lines
}

// renderSongInfo 渲染歌曲信息
func (l *PlayerLayout) renderSongInfo(state *ui.AppState) []string {
	lines := []string{}
	
	if state.Player == nil || state.Player.CurrentSong == nil {
		lines = append(lines, l.padLine("  暂无播放歌曲"))
		lines = append(lines, l.padLine(""))
		return lines
	}
	
	song := state.Player.CurrentSong
	
	// 歌曲标题
	title := song.Title
	if len(title) > l.width-6 {
		title = title[:l.width-9] + "..."
	}
	lines = append(lines, l.padLine(fmt.Sprintf("  ♪ %s", title)))
	
	// 艺术家
	artist := song.Artist
	if len(artist) > l.width-6 {
		artist = artist[:l.width-9] + "..."
	}
	lines = append(lines, l.padLine(fmt.Sprintf("  ♫ %s", artist)))
	
	// 专辑
	if song.Album != "" {
		album := song.Album
		if len(album) > l.width-6 {
			album = album[:l.width-9] + "..."
		}
		lines = append(lines, l.padLine(fmt.Sprintf("  ♬ %s", album)))
	}
	
	lines = append(lines, l.padLine(""))
	return lines
}

// renderPlaybackControls 渲染播放控制
func (l *PlayerLayout) renderPlaybackControls(state *ui.AppState) []string {
	lines := []string{}
	
	if state.Player == nil {
		lines = append(lines, l.padLine("  播放控制不可用"))
		return lines
	}
	
	// 播放状态图标
	playIcon := l.getIcon("pause_icon")
	if state.Player != nil && state.Player.Status == ui.PlayStatusPlaying {
		playIcon = l.getIcon("play_icon")
	}
	
	// 随机播放图标
	shuffleIcon := ""
	if state.Player != nil && state.Player.PlayMode == ui.PlayModeShuffle {
		shuffleIcon = l.getIcon("shuffle_icon")
	}
	
	// 重复播放图标
	repeatIcon := ""
	if state.Player != nil {
		switch state.Player.PlayMode {
		case ui.PlayModeRepeatOne:
			repeatIcon = "🔂"
		case ui.PlayModeRepeatAll:
			repeatIcon = l.getIcon("repeat_icon")
		}
	}
	
	// 构建控制栏
	controls := fmt.Sprintf("  %s %s %s %s %s",
		l.getIcon("prev_icon"),
		playIcon,
		l.getIcon("next_icon"),
		shuffleIcon,
		repeatIcon,
	)
	
	lines = append(lines, l.padLine(controls))
	lines = append(lines, l.padLine(""))
	return lines
}

// renderProgressBar 渲染进度条
func (l *PlayerLayout) renderProgressBar(state *ui.AppState) []string {
	lines := []string{}
	
	if state.Player == nil {
		lines = append(lines, l.padLine("  进度: --:-- / --:--"))
		return lines
	}
	
	position := float64(state.Player.Position.Seconds())
	duration := float64(state.Player.Duration.Seconds())
	
	// 格式化时间
	positionStr := l.formatDuration(position)
	durationStr := l.formatDuration(duration)
	
	// 计算进度条宽度
	timeInfoLen := len(positionStr) + len(durationStr) + 3 // " / "
	progressBarWidth := l.width - timeInfoLen - 6 // 减去边框和空格
	if progressBarWidth < 10 {
		progressBarWidth = 10
	}
	
	// 计算进度
	progress := 0.0
	if duration > 0 {
		progress = position / duration
	}
	if progress > 1.0 {
		progress = 1.0
	}
	
	// 渲染进度条
	filledWidth := int(float64(progressBarWidth) * progress)
	emptyWidth := progressBarWidth - filledWidth
	
	progressChar := l.getProgressChar("progress_char")
	progressBgChar := l.getProgressChar("progress_bg_char")
	
	progressBar := strings.Repeat(progressChar, filledWidth) + strings.Repeat(progressBgChar, emptyWidth)
	
	progressLine := fmt.Sprintf("  %s [%s] %s", positionStr, progressBar, durationStr)
	lines = append(lines, l.padLine(progressLine))
	lines = append(lines, l.padLine(""))
	return lines
}

// renderVolumeControl 渲染音量控制
func (l *PlayerLayout) renderVolumeControl(state *ui.AppState) []string {
	lines := []string{}
	
	if state.Player == nil {
		lines = append(lines, l.padLine("  音量: --%"))
		return lines
	}
	
	volume := int(state.Player.Volume * 100)
	volumeIcon := l.getIcon("volume_icon")
	if state.Player.IsMuted || volume == 0 {
		volumeIcon = l.getIcon("mute_icon")
	}
	
	// 计算音量条宽度
	volumeBarWidth := 20
	filledWidth := int(float64(volumeBarWidth) * float64(volume) / 100.0)
	emptyWidth := volumeBarWidth - filledWidth
	
	volumeChar := l.getProgressChar("volume_char")
	volumeBgChar := l.getProgressChar("volume_bg_char")
	
	volumeBar := strings.Repeat(volumeChar, filledWidth) + strings.Repeat(volumeBgChar, emptyWidth)
	
	volumeLine := fmt.Sprintf("  %s [%s] %d%%", volumeIcon, volumeBar, volume)
	lines = append(lines, l.padLine(volumeLine))
	lines = append(lines, l.padLine(""))
	return lines
}

// renderLyrics 渲染歌词
func (l *PlayerLayout) renderLyrics(state *ui.AppState) []string {
	lines := []string{}
	
	if state.Player == nil || state.Player.CurrentSong == nil {
		return lines
	}
	
	// 添加歌词标题
	lines = append(lines, l.renderBorder("middle"))
	lines = append(lines, l.padLine("  歌词"))
	lines = append(lines, l.padLine(""))
	
	// 实现歌词显示功能
	// Config是map[string]string类型，暂时不支持复杂的歌词数据
	lyrics := []LyricLine{}
	// TODO: 实现歌词功能，需要从其他地方获取歌词数据
	
	// 如果没有歌词数据，显示提示信息
	if len(lyrics) == 0 {
		lines = append(lines, l.padLine("  暂无歌词"))
		return lines
	}
	
	// 获取当前播放时间
	currentTime := state.Player.Position.Seconds()
	
	// 找到当前歌词索引
	currentIndex := l.findCurrentLyricIndex(lyrics, currentTime)
	
	// 计算显示范围
	maxLyricsLines := 8 // 最多显示8行歌词
	startIndex := currentIndex - maxLyricsLines/2
	if startIndex < 0 {
		startIndex = 0
	}
	endIndex := startIndex + maxLyricsLines
	if endIndex > len(lyrics) {
		endIndex = len(lyrics)
		startIndex = endIndex - maxLyricsLines
		if startIndex < 0 {
			startIndex = 0
		}
	}
	
	// 渲染歌词行
	for i := startIndex; i < endIndex; i++ {
		lyric := lyrics[i]
		text := lyric.Text
		
		// 截断过长的歌词
		if len(text) > l.width-6 {
			text = text[:l.width-9] + "..."
		}
		
		// 高亮当前歌词
		if i == currentIndex {
			// 当前歌词行，添加高亮
			lines = append(lines, l.padLine(fmt.Sprintf("▶ %s", text)))
		} else {
			// 普通歌词行
			lines = append(lines, l.padLine(fmt.Sprintf("  %s", text)))
		}
	}
	
	return lines
}

// renderStatusBar 渲染状态栏
func (l *PlayerLayout) renderStatusBar(state *ui.AppState) []string {
	lines := []string{
		l.renderBorder("middle"),
	}
	
	// 左侧：播放模式信息
	leftInfo := "单曲播放"
	if state.Player != nil {
		switch state.Player.PlayMode {
		case ui.PlayModeRepeatOne:
			leftInfo = "单曲循环"
		case ui.PlayModeRepeatAll:
			leftInfo = "列表循环"
		case ui.PlayModeShuffle:
			leftInfo = "随机播放"
		}
	}
	
	// 右侧：快捷键提示
	rightInfo := "空格:播放/暂停 q:返回"
	
	// 计算布局
	leftLen := len(leftInfo)
	rightLen := len(rightInfo)
	availableWidth := l.width - 4 // 减去边框
	
	statusContent := leftInfo
	if leftLen + rightLen + 4 <= availableWidth {
		padding := availableWidth - leftLen - rightLen
		statusContent += strings.Repeat(" ", padding) + rightInfo
	} else {
		// 如果内容太长，只显示右侧信息
		if rightLen + 2 <= availableWidth {
			padding := availableWidth - rightLen
			statusContent = strings.Repeat(" ", padding) + rightInfo
		} else {
			statusContent = rightInfo[:availableWidth-2]
		}
	}
	
	lines = append(lines, l.padLine(" "+statusContent))
	lines = append(lines, l.renderBorder("bottom"))
	
	return lines
}

// findCurrentLyricIndex 查找当前时间对应的歌词索引
// LyricLine 歌词行结构
type LyricLine struct {
	Time float64 `json:"time"`
	Text string  `json:"text"`
}

func (l *PlayerLayout) findCurrentLyricIndex(lyrics []LyricLine, currentTime float64) int {
	for i, lyric := range lyrics {
		if currentTime < lyric.Time {
			if i > 0 {
				return i - 1
			}
			return 0
		}
	}
	if len(lyrics) > 0 {
		return len(lyrics) - 1
	}
	return 0
}

// formatDuration 格式化时长
func (l *PlayerLayout) formatDuration(seconds float64) string {
	if seconds < 0 {
		return "--:--"
	}
	
	duration := time.Duration(seconds * float64(time.Second))
	minutes := int(duration.Minutes())
	secs := int(duration.Seconds()) % 60
	
	if minutes >= 60 {
		hours := minutes / 60
		minutes = minutes % 60
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

// getIcon 获取图标
func (l *PlayerLayout) getIcon(iconName string) string {
	if l.theme != nil && l.theme.Variables != nil {
		if icon, ok := l.theme.Variables[iconName]; ok {
			return icon
		}
	}
	
	// 默认图标
	defaultIcons := map[string]string{
		"play_icon":   "▶",
		"pause_icon":  "⏸",
		"stop_icon":   "⏹",
		"next_icon":   "⏭",
		"prev_icon":   "⏮",
		"shuffle_icon": "🔀",
		"repeat_icon": "🔁",
		"volume_icon": "🔊",
		"mute_icon":   "🔇",
	}
	
	if icon, ok := defaultIcons[iconName]; ok {
		return icon
	}
	return "?"
}

// getProgressChar 获取进度条字符
func (l *PlayerLayout) getProgressChar(charName string) string {
	if l.theme != nil && l.theme.Variables != nil {
		if char, ok := l.theme.Variables[charName]; ok {
			return char
		}
	}
	
	// 默认字符
	defaultChars := map[string]string{
		"progress_char":    "█",
		"progress_bg_char": "░",
		"volume_char":      "■",
		"volume_bg_char":   "□",
	}
	
	if char, ok := defaultChars[charName]; ok {
		return char
	}
	return "█"
}

// renderBorder 渲染边框
func (l *PlayerLayout) renderBorder(position string) string {
	borderChars := map[string]map[string]string{
		"single": {
			"horizontal": "─",
			"vertical":   "│",
			"top_left":   "┌",
			"top_right":  "┐",
			"bottom_left": "└",
			"bottom_right": "┘",
			"cross":      "┼",
			"t_down":     "┬",
			"t_up":       "┴",
			"t_left":     "┤",
			"t_right":    "├",
		},
		"double": {
			"horizontal": "═",
			"vertical":   "║",
			"top_left":   "╔",
			"top_right":  "╗",
			"bottom_left": "╚",
			"bottom_right": "╝",
			"cross":      "╬",
			"t_down":     "╦",
			"t_up":       "╩",
			"t_left":     "╣",
			"t_right":    "╠",
		},
		"rounded": {
			"horizontal": "─",
			"vertical":   "│",
			"top_left":   "╭",
			"top_right":  "╮",
			"bottom_left": "╰",
			"bottom_right": "╯",
			"cross":      "┼",
			"t_down":     "┬",
			"t_up":       "┴",
			"t_left":     "┤",
			"t_right":    "├",
		},
	}
	
	borderStyle := "single"
	if l.theme != nil && l.theme.Variables != nil {
		if style, ok := l.theme.Variables["border_style"]; ok {
			borderStyle = style
		}
	}
	
	chars, exists := borderChars[borderStyle]
	if !exists {
		chars = borderChars["single"]
	}
	
	switch position {
	case "top":
		border := chars["top_left"] + strings.Repeat(chars["horizontal"], l.width-2) + chars["top_right"]
		if len(border) > l.width {
			border = border[:l.width]
		} else if len(border) < l.width {
			border += strings.Repeat(" ", l.width-len(border))
		}
		return border
	case "middle":
		border := chars["t_right"] + strings.Repeat(chars["horizontal"], l.width-2) + chars["t_left"]
		if len(border) > l.width {
			border = border[:l.width]
		} else if len(border) < l.width {
			border += strings.Repeat(" ", l.width-len(border))
		}
		return border
	case "bottom":
		border := chars["bottom_left"] + strings.Repeat(chars["horizontal"], l.width-2) + chars["bottom_right"]
		if len(border) > l.width {
			border = border[:l.width]
		} else if len(border) < l.width {
			border += strings.Repeat(" ", l.width-len(border))
		}
		return border
	default:
		border := strings.Repeat(chars["horizontal"], l.width)
		if len(border) > l.width {
			border = border[:l.width]
		} else if len(border) < l.width {
			border += strings.Repeat(" ", l.width-len(border))
		}
		return border
	}
}

// padLine 填充行到指定宽度
func (l *PlayerLayout) padLine(content string) string {
	if len(content) >= l.width {
		return content[:l.width]
	}
	padding := l.width - len(content)
	return content + strings.Repeat(" ", padding)
}

// SetSize 设置布局大小
func (l *PlayerLayout) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// GetSize 获取布局大小
func (l *PlayerLayout) GetSize() (int, int) {
	return l.width, l.height
}

// SetTheme 设置主题
func (l *PlayerLayout) SetTheme(theme *ui.Theme) {
	l.theme = theme
}

// GetTheme 获取主题
func (l *PlayerLayout) GetTheme() *ui.Theme {
	return l.theme
}