package layouts

import (
	"fmt"
	"strconv"
	"strings"

	ui "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/ui"
)

// MainLayout 主界面布局
type MainLayout struct {
	width  int
	height int
	theme  *ui.Theme
}

// NewMainLayout 创建主界面布局
func NewMainLayout(width, height int, theme *ui.Theme) *MainLayout {
	return &MainLayout{
		width:  width,
		height: height,
		theme:  theme,
	}
}

// Render 渲染主界面布局
func (l *MainLayout) Render(state *ui.AppState) ([]string, error) {
	if state == nil {
		return nil, fmt.Errorf("app state is nil")
	}

	lines := make([]string, l.height)
	
	// 渲染标题栏
	titleLines := l.renderTitle()
	copy(lines[0:len(titleLines)], titleLines)
	
	// 渲染主菜单
	menuStartY := len(titleLines) + 1
	menuLines := l.renderMainMenu(state)
	if menuStartY+len(menuLines) <= l.height {
		copy(lines[menuStartY:menuStartY+len(menuLines)], menuLines)
	}
	
	// 渲染状态栏
	statusLines := l.renderStatusBar(state)
	statusStartY := l.height - len(statusLines)
	if statusStartY > 0 {
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
func (l *MainLayout) renderTitle() []string {
	title := "go-musicfox v2.0"
	subtitle := "Terminal Music Player"
	
	// 居中显示标题
	titlePadding := (l.width - len(title)) / 2
	subtitlePadding := (l.width - len(subtitle)) / 2
	
	lines := []string{
		l.renderBorder("top"),
		l.padLine(strings.Repeat(" ", titlePadding) + title),
		l.padLine(strings.Repeat(" ", subtitlePadding) + subtitle),
		l.renderBorder("middle"),
	}
	
	return lines
}

// renderMainMenu 渲染主菜单
func (l *MainLayout) renderMainMenu(state *ui.AppState) []string {
	menuItems := []string{
		"每日推荐歌曲",
		"每日推荐歌单",
		"我的歌单",
		"我的收藏",
		"私人FM",
		"专辑列表",
		"搜索",
		"排行榜",
		"精选歌单",
		"热门歌手",
		"最近播放歌曲",
		"云盘",
		"主播电台",
		"LastFM",
		"帮助",
		"检查更新",
	}
	
	lines := []string{}
	selectedIndex := 0
	// 从state中获取选中索引
	if state.Config != nil {
		if selectedVal, ok := state.Config["selected_index"]; ok {
			// Config是map[string]string类型，需要解析字符串
			if selectedInt, err := strconv.Atoi(selectedVal); err == nil {
				selectedIndex = selectedInt
			}
		}
	}
	
	// 计算可见区域
	maxVisibleItems := l.height - 8 // 减去标题栏和状态栏的高度
	startIndex := 0
	if selectedIndex >= maxVisibleItems {
		startIndex = selectedIndex - maxVisibleItems + 1
	}
	endIndex := startIndex + maxVisibleItems
	if endIndex > len(menuItems) {
		endIndex = len(menuItems)
	}
	
	for i := startIndex; i < endIndex; i++ {
		item := menuItems[i]
		line := ""
		
		if i == selectedIndex {
			// 高亮选中项
			cursor := l.getCursorStyle()
			line = fmt.Sprintf(" %s %s", cursor, item)
			if l.theme != nil && l.theme.Variables != nil {
				if selectedStyle, ok := l.theme.Variables["selected_style"]; ok {
					line = selectedStyle + line + "[::]"
				}
			}
		} else {
			line = fmt.Sprintf("   %s", item)
		}
		
		lines = append(lines, l.padLine(line))
	}
	
	// 如果有更多项目，显示滚动指示器
	if startIndex > 0 {
		lines[0] = l.padLine("   ↑ 更多项目")
	}
	if endIndex < len(menuItems) {
		lastIndex := len(lines) - 1
		lines[lastIndex] = l.padLine("   ↓ 更多项目")
	}
	
	return lines
}

// renderStatusBar 渲染状态栏
func (l *MainLayout) renderStatusBar(state *ui.AppState) []string {
	lines := []string{
		l.renderBorder("middle"),
	}
	
	// 左侧：用户信息
	leftInfo := "未登录"
	if state.User != nil && state.User.Username != "" {
		leftInfo = fmt.Sprintf("用户: %s", state.User.Username)
	}
	
	// 中间：当前播放信息
	centerInfo := "暂无播放"
	if state.Player != nil && state.Player.CurrentSong != nil {
		song := state.Player.CurrentSong
		status := "⏸"
		if state.Player.Status == ui.PlayStatusPlaying {
			status = "▶"
		}
		centerInfo = fmt.Sprintf("%s %s - %s", status, song.Title, song.Artist)
	}
	
	// 右侧：快捷键提示
	rightInfo := "q:退出 h:帮助"
	
	// 计算布局
	leftLen := len(leftInfo)
	centerLen := len(centerInfo)
	rightLen := len(rightInfo)
	
	availableWidth := l.width - 4 // 减去边框
	rightPadding := availableWidth - rightLen - 2
	centerPadding := (availableWidth - centerLen) / 2
	
	// 调整布局以避免重叠
	if leftLen + centerLen + rightLen + 6 > availableWidth {
		// 如果内容太长，优先显示播放信息
		if centerLen + rightLen + 4 <= availableWidth {
			leftInfo = ""
			centerPadding = 2
			rightPadding = availableWidth - centerLen - rightLen - 4
		} else {
			// 截断中间信息
			maxCenterLen := availableWidth - rightLen - 6
			if maxCenterLen > 10 {
				centerInfo = centerInfo[:maxCenterLen-3] + "..."
			} else {
				centerInfo = ""
			}
			leftInfo = ""
		}
	}
	
	// 构建状态栏内容
	statusContent := ""
	if leftInfo != "" {
		statusContent += leftInfo
	}
	
	if centerInfo != "" {
		currentLen := len(statusContent)
		padding := centerPadding - currentLen
		if padding > 0 {
			statusContent += strings.Repeat(" ", padding)
		}
		statusContent += centerInfo
	}
	
	if rightInfo != "" {
		currentLen := len(statusContent)
		padding := rightPadding - currentLen
		if padding > 0 {
			statusContent += strings.Repeat(" ", padding)
		}
		statusContent += rightInfo
	}
	
	lines = append(lines, l.padLine(" "+statusContent))
	lines = append(lines, l.renderBorder("bottom"))
	
	return lines
}

// renderBorder 渲染边框
func (l *MainLayout) renderBorder(position string) string {
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
func (l *MainLayout) padLine(content string) string {
	if len(content) >= l.width {
		return content[:l.width]
	}
	padding := l.width - len(content)
	return content + strings.Repeat(" ", padding)
}

// getCursorStyle 获取光标样式
func (l *MainLayout) getCursorStyle() string {
	if l.theme != nil && l.theme.Variables != nil {
		if cursor, ok := l.theme.Variables["cursor_style"]; ok {
			return cursor
		}
	}
	return "▶"
}

// SetSize 设置布局大小
func (l *MainLayout) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// GetSize 获取布局大小
func (l *MainLayout) GetSize() (int, int) {
	return l.width, l.height
}

// SetTheme 设置主题
func (l *MainLayout) SetTheme(theme *ui.Theme) {
	l.theme = theme
}

// GetTheme 获取主题
func (l *MainLayout) GetTheme() *ui.Theme {
	return l.theme
}