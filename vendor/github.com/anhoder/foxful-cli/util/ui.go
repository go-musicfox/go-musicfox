package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

var (
	TermProfile      = termenv.ColorProfile()
	PrimaryColor     string
	_primaryColor    termenv.Color
	_primaryColorStr string
)

// GetPrimaryColor get random color
func GetPrimaryColor() termenv.Color {
	if _primaryColor != nil {
		return _primaryColor
	}
	initPrimaryColor()
	return _primaryColor
}

var MenuTitleColor string = "" // 菜单标题颜色配置
var ProgressColorExcludeRanges string = "45-210" // 进度条颜色排除区间
var ProgressColorSaturation string = "30-70,30-70" // 进度条颜色饱和度范围

// GetMenuTitleColor 获取菜单标题颜色
func GetMenuTitleColor() termenv.Color {
	if MenuTitleColor != "" && MenuTitleColor != PrimaryColor {
		return TermProfile.Color(MenuTitleColor)
	}
	return GetPrimaryColor()
}

// parseExcludeRanges 解析排除区间字符串，返回排除的色相范围
func parseExcludeRanges(rangeStr string) [][]float64 {
	if rangeStr == "" {
		return [][]float64{{45, 210}} // 默认排除绿色和青色
	}
	
	var excludeRanges [][]float64
	parts := strings.Split(rangeStr, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		rangeParts := strings.Split(part, "-")
		if len(rangeParts) == 2 {
			start, err1 := strconv.ParseFloat(strings.TrimSpace(rangeParts[0]), 64)
			end, err2 := strconv.ParseFloat(strings.TrimSpace(rangeParts[1]), 64)
			
			if err1 == nil && err2 == nil && start >= 0 && end <= 360 && start < end {
				excludeRanges = append(excludeRanges, []float64{start, end})
			}
		}
	}
	
	if len(excludeRanges) == 0 {
		return [][]float64{{45, 210}} // 默认排除绿色和青色
	}
	
	return excludeRanges
}

// generateSafeColorRanges 根据排除区间生成安全的色相范围
func generateSafeColorRanges() [][]float64 {
	excludeRanges := parseExcludeRanges(ProgressColorExcludeRanges)
	
	// 将排除区间按起始位置排序
	for i := 0; i < len(excludeRanges)-1; i++ {
		for j := i + 1; j < len(excludeRanges); j++ {
			if excludeRanges[i][0] > excludeRanges[j][0] {
				excludeRanges[i], excludeRanges[j] = excludeRanges[j], excludeRanges[i]
			}
		}
	}
	
	var safeRanges [][]float64
	lastEnd := 0.0
	
	for _, exclude := range excludeRanges {
		start, end := exclude[0], exclude[1]
		
		// 如果当前排除区间之前有安全区间，添加它
		if start > lastEnd {
			safeRanges = append(safeRanges, []float64{lastEnd, start})
		}
		lastEnd = end
	}
	
	// 添加最后一个安全区间（如果存在）
	if lastEnd < 360 {
		safeRanges = append(safeRanges, []float64{lastEnd, 360})
	}
	
	// 如果没有安全区间，使用默认的安全区间
	if len(safeRanges) == 0 {
		safeRanges = [][]float64{
			{0, 50},     // 红色到橙色
			{200, 240},  // 蓝色
			{240, 280},  // 蓝紫色
			{280, 360},  // 紫色到品红
		}
	}
	
	return safeRanges
}

// parseSaturationRanges 解析饱和度范围字符串，返回起始色和结束色的饱和度范围
func parseSaturationRanges(saturationStr string) (startMin, startMax, endMin, endMax float64) {
	// 默认值
	startMin, startMax, endMin, endMax = 0.5, 0.8, 0.4, 0.8
	
	if saturationStr == "" {
		return
	}
	
	parts := strings.Split(saturationStr, ",")
	if len(parts) != 2 {
		return
	}
	
	// 解析起始色饱和度范围
	startParts := strings.Split(strings.TrimSpace(parts[0]), "-")
	if len(startParts) == 2 {
		if min, err := strconv.ParseFloat(strings.TrimSpace(startParts[0]), 64); err == nil && min >= 0 && min <= 100 {
			startMin = min / 100.0
		}
		if max, err := strconv.ParseFloat(strings.TrimSpace(startParts[1]), 64); err == nil && max >= 0 && max <= 100 {
			startMax = max / 100.0
		}
	}
	
	// 解析结束色饱和度范围
	endParts := strings.Split(strings.TrimSpace(parts[1]), "-")
	if len(endParts) == 2 {
		if min, err := strconv.ParseFloat(strings.TrimSpace(endParts[0]), 64); err == nil && min >= 0 && min <= 100 {
			endMin = min / 100.0
		}
		if max, err := strconv.ParseFloat(strings.TrimSpace(endParts[1]), 64); err == nil && max >= 0 && max <= 100 {
			endMax = max / 100.0
		}
	}
	
	// 确保最小值不大于最大值
	if startMin > startMax {
		startMin, startMax = startMax, startMin
	}
	if endMin > endMax {
		endMin, endMax = endMax, endMin
	}
	
	return
}

func GetPrimaryColorString() string {
	if _primaryColorStr != "" {
		return _primaryColorStr
	}
	initPrimaryColor()
	return _primaryColorStr
}

func initPrimaryColor() {
	if _primaryColorStr != "" && _primaryColor != nil {
		return
	}
	if PrimaryColor == "" || PrimaryColor == RandomColor {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		_primaryColorStr = strconv.Itoa(rand.Intn(228-17) + 17)
	} else {
		_primaryColorStr = PrimaryColor
	}
	_primaryColor = TermProfile.Color(GetPrimaryColorString())
}

// GetRandomRgbColor get random rgb color
func GetRandomRgbColor(isRange bool) (string, string) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// 使用HSV色彩空间，根据配置生成安全的色相范围
	ranges := generateSafeColorRanges()
	
	// 随机选择一个安全区间
	selectedRange := ranges[rand.Intn(len(ranges))]
	hue := selectedRange[0] + rand.Float64()*(selectedRange[1]-selectedRange[0])
	
	// 解析饱和度配置
	startMin, startMax, endMin, endMax := parseSaturationRanges(ProgressColorSaturation)
	
	saturation := startMin + rand.Float64() * (startMax - startMin)  // 使用配置的饱和度范围
	value := 0.3 + rand.Float64() * 0.65       // 30-95% 亮度
	
	r, g, b := hsvToRgb(hue, saturation, value)
	startColor := fmt.Sprintf("#%02x%02x%02x", r, g, b)
	
	if !isRange {
		return startColor, ""
	}

	// 为结束颜色生成另一个安全色相
	rand.New(rand.NewSource(time.Now().UnixNano() / 5))
	var hueEnd float64
	
	// 50%概率在同一区间内生成相近色，50%概率跳到另一个安全区间
	if rand.Float64() < 0.5 {
		// 在同一区间内生成相近色
		rangeWidth := selectedRange[1] - selectedRange[0]
		maxOffset := rangeWidth * 0.6  // 最多偏移区间宽度的60%
		offset := (rand.Float64() - 0.5) * maxOffset
		hueEnd = hue + offset
		
		// 确保不超出区间
		if hueEnd < selectedRange[0] {
			hueEnd = selectedRange[0]
		} else if hueEnd > selectedRange[1] {
			hueEnd = selectedRange[1]
		}
	} else {
		// 跳到另一个安全区间，实现更丰富的色彩组合
		otherRanges := make([][]float64, 0)
		for _, r := range ranges {
			if r[0] != selectedRange[0] || r[1] != selectedRange[1] {  // 排除当前区间（精确比较）
				otherRanges = append(otherRanges, r)
			}
		}
		
		if len(otherRanges) > 0 {
			endRange := otherRanges[rand.Intn(len(otherRanges))]
			hueEnd = endRange[0] + rand.Float64()*(endRange[1]-endRange[0])
		} else {
			// 如果没有其他区间，在当前区间内生成
			rangeWidth := selectedRange[1] - selectedRange[0]
			maxOffset := rangeWidth * 0.6
			offset := (rand.Float64() - 0.5) * maxOffset
			hueEnd = hue + offset
			
			if hueEnd < selectedRange[0] {
				hueEnd = selectedRange[0]
			} else if hueEnd > selectedRange[1] {
				hueEnd = selectedRange[1]
			}
		}
	}
	
	saturationEnd := endMin + rand.Float64() * (endMax - endMin)  // 使用配置的结束色饱和度范围
	valueEnd := 0.3 + rand.Float64() * 0.55       // 30-85% 亮度
	
	rEnd, gEnd, bEnd := hsvToRgb(hueEnd, saturationEnd, valueEnd)
	endColor := fmt.Sprintf("#%02x%02x%02x", rEnd, gEnd, bEnd)

	return startColor, endColor
}

// hsvToRgb converts HSV to RGB values
func hsvToRgb(h, s, v float64) (r, g, b int) {
	c := v * s
	x := c * (1 - abs(mod(h/60, 2) - 1))
	m := v - c
	
	var r1, g1, b1 float64
	
	switch {
	case h < 60:
		r1, g1, b1 = c, x, 0
	case h < 120:
		r1, g1, b1 = x, c, 0
	case h < 180:
		r1, g1, b1 = 0, c, x
	case h < 240:
		r1, g1, b1 = 0, x, c
	case h < 300:
		r1, g1, b1 = x, 0, c
	default:
		r1, g1, b1 = c, 0, x
	}
	
	r = int((r1 + m) * 255)
	g = int((g1 + m) * 255)
	b = int((b1 + m) * 255)
	
	return
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func mod(x, y float64) float64 {
	return x - y*float64(int(x/y))
}

// SetFgStyle Return a function that will colorize the foreground of a given string.
func SetFgStyle(content string, color termenv.Color) string {
	return termenv.Style{}.Foreground(color).Styled(content)
}

// SetFgBgStyle Color a string's foreground and background with the given value.
func SetFgBgStyle(content string, fg, bg termenv.Color) string {
	return termenv.Style{}.Foreground(fg).Background(bg).Styled(content)
}

// SetNormalStyle don't set any style
func SetNormalStyle(content string) string {
	seq := strings.Join([]string{"0"}, ";")
	return fmt.Sprintf("%s%sm%s%sm", termenv.CSI, seq, content, termenv.CSI+termenv.ResetSeq)
}

func GetPrimaryFontStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(GetPrimaryColorString()))
}

// MakeRamp Generate a blend of colors using HSV to avoid green hues.
func MakeRamp(colorA, colorB string, steps float64) (result []string) {
	cA, _ := colorful.Hex(colorA)
	cB, _ := colorful.Hex(colorB)
	
	// 转换为HSV
	hA, sA, vA := cA.Hsv()
	hB, sB, vB := cB.Hsv()
	
	// 确保色相以最短路径插值，且避开绿色区域
	if abs(hB - hA) > 180 {
		if hA < hB {
			hA += 360
		} else {
			hB += 360
		}
	}
	
	for i := 0.0; i < steps; i++ {
		t := i / steps
		
		// 在HSV空间中插值
		h := hA + (hB - hA) * t
		saturation := sA + (sB - sA) * t
		value := vA + (vB - vA) * t
		
		// 确保色相在0-360范围内
		for h < 0 {
			h += 360
		}
		for h >= 360 {
			h -= 360
		}
		
		// 转回RGB
		c := colorful.Hsv(h, saturation, value)
		result = append(result, colorToHex(c))
	}
	return
}

// Convert a colorful.Color to a hexidecimal format compatible with termenv.
func colorToHex(c colorful.Color) string {
	return fmt.Sprintf("#%s%s%s", colorFloatToHex(c.R), colorFloatToHex(c.G), colorFloatToHex(c.B))
}

// Helper function for converting colors to hex. Assumes a value between 0 and 1.
func colorFloatToHex(f float64) (s string) {
	s = strconv.FormatInt(int64(f*255), 16)
	if len(s) == 1 {
		s = "0" + s
	}
	return
}
