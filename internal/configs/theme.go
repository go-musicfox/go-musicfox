package configs

import (
	"strconv"
	"strings"

	"github.com/muesli/termenv"
)

// ThemeOptions 主题配置选项
type ThemeOptions struct {
	// 基础颜色配置
	PrimaryColor    string // 主题色
	MenuTitleColor  string // 菜单标题颜色
	
	// 进度条颜色配置
	ProgressColor ProgressColorOptions
	
	// 歌词颜色配置
	LyricColor LyricColorOptions
}

// ProgressColorOptions 进度条颜色配置
type ProgressColorOptions struct {
	ExcludeRanges string // 颜色排除区间，格式如："45-210"
	Saturation    string // 饱和度范围，格式如："50-80,40-80"
	Brightness    string // 亮度范围，格式如："70-90,60-80"
}

// LyricColorOptions 歌词颜色配置
type LyricColorOptions struct {
	CurrentLine string // 当前行颜色，默认为亮青色
	OtherLines  string // 其他行颜色，默认为灰色
}

// GetPrimaryColor 获取主题色
func (t *ThemeOptions) GetPrimaryColor() termenv.Color {
	if t.PrimaryColor == "" {
		return termenv.ANSIRed // 默认红色
	}
	return termenv.ColorProfile().Color(t.PrimaryColor)
}

// GetMenuTitleColor 获取菜单标题颜色
func (t *ThemeOptions) GetMenuTitleColor() termenv.Color {
	if t.MenuTitleColor == "" {
		return t.GetPrimaryColor() // 默认使用主题色
	}
	return termenv.ColorProfile().Color(t.MenuTitleColor)
}

// GetLyricCurrentLineColor 获取歌词当前行颜色
func (t *ThemeOptions) GetLyricCurrentLineColor() termenv.Color {
	if t.LyricColor.CurrentLine == "" {
		return termenv.ANSIBrightCyan // 默认亮青色
	}
	return termenv.ColorProfile().Color(t.LyricColor.CurrentLine)
}

// GetLyricOtherLinesColor 获取歌词其他行颜色
func (t *ThemeOptions) GetLyricOtherLinesColor() termenv.Color {
	if t.LyricColor.OtherLines == "" {
		return termenv.ANSIBrightBlack // 默认灰色
	}
	return termenv.ColorProfile().Color(t.LyricColor.OtherLines)
}

// ParseProgressColorExcludeRanges 解析进度条颜色排除区间
func (p *ProgressColorOptions) ParseExcludeRanges() [][]int {
	if p.ExcludeRanges == "" {
		return [][]int{{45, 210}} // 默认排除区间
	}
	
	ranges := strings.Split(p.ExcludeRanges, ",")
	var result [][]int
	
	for _, r := range ranges {
		parts := strings.Split(strings.TrimSpace(r), "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				result = append(result, []int{start, end})
			}
		}
	}
	
	if len(result) == 0 {
		return [][]int{{45, 210}} // 默认排除区间
	}
	return result
}

// ParseProgressColorSaturation 解析进度条颜色饱和度范围
func (p *ProgressColorOptions) ParseSaturation() [][]int {
	if p.Saturation == "" {
		return [][]int{{50, 80}, {40, 80}} // 默认饱和度范围
	}
	
	ranges := strings.Split(p.Saturation, ",")
	var result [][]int
	
	for _, r := range ranges {
		parts := strings.Split(strings.TrimSpace(r), "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				result = append(result, []int{start, end})
			}
		}
	}
	
	if len(result) == 0 {
		return [][]int{{50, 80}, {40, 80}} // 默认饱和度范围
	}
	return result
}

// ParseProgressColorBrightness 解析进度条颜色亮度范围
func (p *ProgressColorOptions) ParseBrightness() [][]int {
	if p.Brightness == "" {
		return [][]int{{70, 90}, {60, 80}} // 默认亮度范围
	}
	
	ranges := strings.Split(p.Brightness, ",")
	var result [][]int
	
	for _, r := range ranges {
		parts := strings.Split(strings.TrimSpace(r), "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				result = append(result, []int{start, end})
			}
		}
	}
	
	if len(result) == 0 {
		return [][]int{{70, 90}, {60, 80}} // 默认亮度范围
	}
	return result
}