package layouts

import (
	"fmt"
	"strconv"
	"strings"

	ui "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/ui"
)

// SearchLayout 搜索界面布局
type SearchLayout struct {
	width  int
	height int
	theme  *ui.Theme
}

// NewSearchLayout 创建搜索界面布局
func NewSearchLayout(width, height int, theme *ui.Theme) *SearchLayout {
	return &SearchLayout{
		width:  width,
		height: height,
		theme:  theme,
	}
}

// Render 渲染搜索界面布局
func (l *SearchLayout) Render(state *ui.AppState) ([]string, error) {
	if state == nil {
		return nil, fmt.Errorf("app state is nil")
	}

	lines := make([]string, l.height)
	currentY := 0
	
	// 渲染标题栏
	titleLines := l.renderTitle()
	copy(lines[currentY:currentY+len(titleLines)], titleLines)
	currentY += len(titleLines)
	
	// 渲染搜索类型标签
	typeTabLines := l.renderSearchTypeTabs(state)
	if currentY+len(typeTabLines) <= l.height {
		copy(lines[currentY:currentY+len(typeTabLines)], typeTabLines)
		currentY += len(typeTabLines)
	}
	
	// 渲染搜索输入框
	inputLines := l.renderSearchInput(state)
	if currentY+len(inputLines) <= l.height {
		copy(lines[currentY:currentY+len(inputLines)], inputLines)
		currentY += len(inputLines)
	}
	
	// 渲染搜索结果
	resultLines := l.renderSearchResults(state)
	resultHeight := l.height - currentY - 3 // 减去状态栏高度
	if resultHeight > 0 && len(resultLines) > 0 {
		if len(resultLines) > resultHeight {
			resultLines = resultLines[:resultHeight]
		}
		copy(lines[currentY:currentY+len(resultLines)], resultLines)
		currentY += len(resultLines)
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
func (l *SearchLayout) renderTitle() []string {
	title := "搜索"
	
	// 居中显示标题
	titlePadding := (l.width - len(title)) / 2
	
	lines := []string{
		l.renderBorder("top"),
		l.padLine(strings.Repeat(" ", titlePadding) + title),
		l.renderBorder("middle"),
	}
	
	return lines
}

// renderSearchTypeTabs 渲染搜索类型标签
func (l *SearchLayout) renderSearchTypeTabs(state *ui.AppState) []string {
	lines := []string{}
	
	searchTypes := []struct {
		key   string
		label string
	}{
		{"song", "歌曲"},
		{"album", "专辑"},
		{"artist", "歌手"},
		{"playlist", "歌单"},
	}
	
	currentType := "song"
	// 从state.Config中获取搜索类型
	if searchType, ok := state.Config["search_type"]; ok {
		// Config是map[string]string类型，直接使用
		currentType = searchType
	}
	
	// 构建标签栏
	tabsContent := "  "
	for i, searchType := range searchTypes {
		if i > 0 {
			tabsContent += " | "
		}
		
		if searchType.key == currentType {
			// 高亮当前选中的类型
			tabsContent += fmt.Sprintf("[%s]", searchType.label)
		} else {
			tabsContent += searchType.label
		}
	}
	
	lines = append(lines, l.padLine(tabsContent))
	lines = append(lines, l.padLine(""))
	return lines
}

// renderSearchInput 渲染搜索输入框
func (l *SearchLayout) renderSearchInput(state *ui.AppState) []string {
	lines := []string{}
	
	query := ""
	// 从state.Config中获取搜索查询
	if searchQueryVal, ok := state.Config["search_query"]; ok {
		// Config是map[string]string类型，直接使用
		query = searchQueryVal
	}
	
	// 搜索框标签
	lines = append(lines, l.padLine("  搜索关键词:"))
	
	// 搜索输入框
	inputBoxWidth := l.width - 6
	inputContent := query
	if len(inputContent) > inputBoxWidth-4 {
		inputContent = inputContent[:inputBoxWidth-7] + "..."
	}
	
	// 添加光标（如果正在输入）
	isInputting := false
	// 从state.Config中获取搜索状态
	if searchingVal2, ok := state.Config["is_searching"]; ok {
		// Config是map[string]string类型，检查字符串值
		if searchingVal2 == "true" {
			isInputting = true
		}
	}
	
	if isInputting {
		inputContent += "|"
	}
	
	// 填充输入框
	padding := inputBoxWidth - len(inputContent) - 2
	if padding < 0 {
		padding = 0
	}
	inputBox := fmt.Sprintf("  [%s%s]", inputContent, strings.Repeat(" ", padding))
	lines = append(lines, l.padLine(inputBox))
	
	// 搜索提示
	if query == "" {
		lines = append(lines, l.padLine("  输入关键词后按回车搜索"))
	} else if searchingVal, ok := state.Config["is_searching"]; ok {
		// Config是map[string]string类型，检查字符串值
		if searchingVal == "true" {
			lines = append(lines, l.padLine("  搜索中..."))
		} else {
			// 显示搜索结果数量
			resultCount := 0
			if countVal, ok := state.Config["search_result_count"]; ok {
				// Config是map[string]string类型，需要解析字符串
				if count, err := strconv.Atoi(countVal); err == nil {
					resultCount = count
				}
			}
			lines = append(lines, l.padLine(fmt.Sprintf("  找到 %d 个结果", resultCount)))
		}
	} else {
		// 显示搜索结果数量
		resultCount := 0
		if countVal2, ok := state.Config["search_result_count"]; ok {
			// Config是map[string]string类型，需要解析字符串
			if count, err := strconv.Atoi(countVal2); err == nil {
				resultCount = count
			}
		}
		lines = append(lines, l.padLine(fmt.Sprintf("  找到 %d 个结果", resultCount)))
	}
	
	lines = append(lines, l.padLine(""))
	return lines
}

// renderSearchResults 渲染搜索结果
func (l *SearchLayout) renderSearchResults(state *ui.AppState) []string {
	lines := []string{}
	
	// 实现搜索结果显示
	query := ""
	if searchQuery, ok := state.Config["search_query"]; ok {
		// Config是map[string]string类型，直接使用
		query = searchQuery
	}
	
	if query == "" {
		lines = append(lines, l.padLine("  输入关键词开始搜索"))
		return lines
	}
	
	// 获取搜索结果 - Config是map[string]string类型，暂时不支持复杂数据
	results := []interface{}{}
	// TODO: 搜索结果需要从其他地方获取，Config只能存储字符串

	// 获取选中索引
	selectedIndex := 0
	if selectedVal3, ok := state.Config["selected_index"]; ok {
		// Config是map[string]string类型，需要解析字符串
		if index, err := strconv.Atoi(selectedVal3); err == nil {
			selectedIndex = index
		}
	}
	
	// 显示搜索结果
	if len(results) == 0 {
		// 检查是否正在搜索
		if searchingVal, ok := state.Config["is_searching"]; ok {
			// Config是map[string]string类型，检查字符串值
			if searchingVal == "true" {
				lines = append(lines, l.padLine("  搜索中，请稍候..."))
				return lines
			}
		}
		lines = append(lines, l.padLine("  未找到相关结果"))
		return lines
	}
	
	// 渲染搜索结果列表
	for i, result := range results {
		formattedResult := l.formatSearchResult(result, i, selectedIndex)
		lines = append(lines, l.padLine(formattedResult))
		
		// 限制显示的结果数量，避免界面过长
		if i >= 20 {
			lines = append(lines, l.padLine("  ... 更多结果"))
			break
		}
	}
	
	return lines
}

// formatSearchResult 格式化搜索结果
func (l *SearchLayout) formatSearchResult(result interface{}, index, selectedIndex int) string {
	cursor := "  "
	if index == selectedIndex {
		cursor = l.getCursorStyle() + " "
	}
	
	// 尝试解析结果为map
	if resultMap, ok := result.(map[string]interface{}); ok {
		// 获取结果类型
		resultType := "unknown"
		if typeVal, exists := resultMap["type"]; exists {
			if typeStr, isString := typeVal.(string); isString {
				resultType = typeStr
			}
		}
		
		// 根据类型格式化
		switch resultType {
		case "song":
			title := "未知歌曲"
			artist := "未知歌手"
			if titleVal, exists := resultMap["title"]; exists {
				if titleStr, isString := titleVal.(string); isString {
					title = titleStr
				}
			}
			if artistVal, exists := resultMap["artist"]; exists {
				if artistStr, isString := artistVal.(string); isString {
					artist = artistStr
				}
			}
			return cursor + fmt.Sprintf("🎵 %s - %s", title, artist)
			
		case "album":
			name := "未知专辑"
			artist := "未知歌手"
			if nameVal, exists := resultMap["name"]; exists {
				if nameStr, isString := nameVal.(string); isString {
					name = nameStr
				}
			}
			if artistVal, exists := resultMap["artist"]; exists {
				if artistStr, isString := artistVal.(string); isString {
					artist = artistStr
				}
			}
			return cursor + fmt.Sprintf("💿 %s - %s", name, artist)
			
		case "artist":
			name := "未知歌手"
			if nameVal, exists := resultMap["name"]; exists {
				if nameStr, isString := nameVal.(string); isString {
					name = nameStr
				}
			}
			return cursor + fmt.Sprintf("🎤 %s", name)
			
		case "playlist":
			name := "未知歌单"
			if nameVal, exists := resultMap["name"]; exists {
				if nameStr, isString := nameVal.(string); isString {
					name = nameStr
				}
			}
			return cursor + fmt.Sprintf("📋 %s", name)
			
		default:
			// 尝试获取通用的名称或标题
			name := "未知项目"
			if nameVal, exists := resultMap["name"]; exists {
				if nameStr, isString := nameVal.(string); isString {
					name = nameStr
				}
			} else if titleVal, exists := resultMap["title"]; exists {
				if titleStr, isString := titleVal.(string); isString {
					name = titleStr
				}
			}
			return cursor + fmt.Sprintf("• %s", name)
		}
	}
	
	// 如果无法解析为map，对于字符串类型显示为搜索结果项
	if _, ok := result.(string); ok {
		return cursor + "♪ 搜索结果项"
	}
	// 其他类型直接转换为字符串显示
	return cursor + fmt.Sprintf("%d. %v", index+1, result)
}

// renderStatusBar 渲染状态栏
func (l *SearchLayout) renderStatusBar(state *ui.AppState) []string {
	lines := []string{
		l.renderBorder("middle"),
	}
	
	// 左侧：搜索统计信息
	leftInfo := ""
	if searching, ok := state.Config["is_searching"]; ok && searching == "true" {
		leftInfo = "搜索中..."
	} else if query, ok := state.Config["search_query"]; ok && query != "" {
		leftInfo = "搜索完成"
	}
	
	// 右侧：快捷键提示
	rightInfo := "回车:搜索 Tab:切换类型 q:返回"
	
	// 计算布局
	leftLen := len(leftInfo)
	rightLen := len(rightInfo)
	availableWidth := l.width - 4 // 减去边框
	
	statusContent := leftInfo
	if leftLen + rightLen + 4 <= availableWidth {
		padding := availableWidth - leftLen - rightLen
		statusContent += strings.Repeat(" ", padding) + rightInfo
	} else {
		// 如果内容太长，优先显示右侧快捷键
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

// renderBorder 渲染边框
func (l *SearchLayout) renderBorder(position string) string {
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
func (l *SearchLayout) padLine(content string) string {
	if len(content) >= l.width {
		return content[:l.width]
	}
	padding := l.width - len(content)
	return content + strings.Repeat(" ", padding)
}

// getCursorStyle 获取光标样式
func (l *SearchLayout) getCursorStyle() string {
	if l.theme != nil && l.theme.Variables != nil {
		if cursor, ok := l.theme.Variables["cursor_style"]; ok {
			return cursor
		}
	}
	return "▶"
}

// SetSize 设置布局大小
func (l *SearchLayout) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// GetSize 获取布局大小
func (l *SearchLayout) GetSize() (int, int) {
	return l.width, l.height
}

// SetTheme 设置主题
func (l *SearchLayout) SetTheme(theme *ui.Theme) {
	l.theme = theme
}

// GetTheme 获取主题
func (l *SearchLayout) GetTheme() *ui.Theme {
	return l.theme
}