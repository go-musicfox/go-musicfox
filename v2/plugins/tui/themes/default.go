package themes

import (
	"strings"

	ui "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/ui"
)

// DefaultTheme 默认主题
var DefaultTheme = &ui.Theme{
	ID:          "default",
	Name:        "Default",
	Description: "Default TUI theme",
	Version:     "1.0.0",
	Author:      "go-musicfox",
	Colors: &ui.ColorScheme{
		Primary:       "#1DB954", // Spotify green
		Secondary:     "#191414", // Dark background
		Accent:        "#1ED760", // Bright green
		Background:    "#000000", // Black
		Surface:       "#121212", // Dark gray
		Text:          "#FFFFFF", // White text
		TextSecondary: "#B3B3B3", // Gray text
		Border:        "#404040", // Dark border
		Error:         "#E22134", // Red
		Warning:       "#FFB000", // Orange
		Info:          "#2196F3", // Blue
		Success:       "#4CAF50", // Green
	},
	Fonts: &ui.FontScheme{
		Primary:   &ui.FontConfig{Family: "default", Size: 12},
		Secondary: &ui.FontConfig{Family: "default", Size: 11},
		Monospace: &ui.FontConfig{Family: "monospace", Size: 11},
		Display:   &ui.FontConfig{Family: "default", Size: 14},
	},
	Spacing: &ui.SpacingScheme{
		XS: "1px",
		SM: "2px",
		MD: "4px",
		LG: "8px",
		XL: "16px",
	},
	Borders: &ui.BorderScheme{
		Width:  "1px",
		Style:  "solid",
		Radius: "4px",
	},
	Shadows: &ui.ShadowScheme{
		Small:  "0 1px 3px rgba(0,0,0,0.12)",
		Medium: "0 4px 6px rgba(0,0,0,0.16)",
		Large:  "0 10px 25px rgba(0,0,0,0.19)",
	},
	Animations: &ui.AnimationScheme{
		Duration: "300ms",
		Easing:   "ease",
		Delay:    "0ms",
	},
	Variables: map[string]string{
		"cursor_style":      "▶",
		"selected_style":    "[::r]",
		"progress_char":     "█",
		"progress_bg_char":  "░",
		"volume_char":       "■",
		"volume_bg_char":    "□",
		"border_style":      "rounded",
		"list_separator":    "─",
		"status_separator":  "│",
		"play_icon":         "▶",
		"pause_icon":        "⏸",
		"stop_icon":         "⏹",
		"next_icon":         "⏭",
		"prev_icon":         "⏮",
		"shuffle_icon":      "🔀",
		"repeat_icon":       "🔁",
		"volume_icon":       "🔊",
		"mute_icon":         "🔇",
		"heart_icon":        "♥",
		"star_icon":         "★",
	},
	IsDark:    true,
	IsDefault: true,
}

// DarkTheme 暗色主题
var DarkTheme = &ui.Theme{
	ID:          "dark",
	Name:        "Dark",
	Description: "Dark TUI theme",
	Version:     "1.0.0",
	Author:      "go-musicfox",
	Colors: &ui.ColorScheme{
		Primary:       "#BB86FC", // Purple
		Secondary:     "#03DAC6", // Teal
		Accent:        "#CF6679", // Pink
		Background:    "#121212", // Very dark gray
		Surface:       "#1E1E1E", // Dark gray
		Text:          "#FFFFFF", // White text
		TextSecondary: "#B3B3B3", // Gray text
		Border:        "#404040", // Dark border
		Error:         "#CF6679", // Pink red
		Warning:       "#FFB74D", // Orange
		Info:          "#81C784", // Light green
		Success:       "#A5D6A7", // Pale green
	},
	Fonts:      DefaultTheme.Fonts,
	Spacing:    DefaultTheme.Spacing,
	Borders:    DefaultTheme.Borders,
	Shadows:    DefaultTheme.Shadows,
	Animations: DefaultTheme.Animations,
	Variables: map[string]string{
		"cursor_style":      "►",
		"selected_style":    "[::b]",
		"progress_char":     "▓",
		"progress_bg_char":  "░",
		"volume_char":       "▮",
		"volume_bg_char":    "▯",
		"border_style":      "double",
		"list_separator":    "═",
		"status_separator":  "║",
		"play_icon":         "▶",
		"pause_icon":        "⏸",
		"stop_icon":         "⏹",
		"next_icon":         "⏭",
		"prev_icon":         "⏮",
		"shuffle_icon":      "🔀",
		"repeat_icon":       "🔁",
		"volume_icon":       "🔊",
		"mute_icon":         "🔇",
		"heart_icon":        "♥",
		"star_icon":         "★",
	},
	IsDark:    true,
	IsDefault: false,
}

// LightTheme 亮色主题
var LightTheme = &ui.Theme{
	ID:          "light",
	Name:        "Light",
	Description: "Light TUI theme",
	Version:     "1.0.0",
	Author:      "go-musicfox",
	Colors: &ui.ColorScheme{
		Primary:       "#6200EE", // Deep purple
		Secondary:     "#018786", // Dark teal
		Accent:        "#B00020", // Dark red
		Background:    "#FFFFFF", // White
		Surface:       "#F5F5F5", // Light gray
		Text:          "#000000", // Black text
		TextSecondary: "#666666", // Gray text
		Border:        "#CCCCCC", // Light border
		Error:         "#B00020", // Dark red
		Warning:       "#FF6F00", // Dark orange
		Info:          "#1976D2", // Dark blue
		Success:       "#388E3C", // Dark green
	},
	Fonts:      DefaultTheme.Fonts,
	Spacing:    DefaultTheme.Spacing,
	Borders:    DefaultTheme.Borders,
	Shadows:    DefaultTheme.Shadows,
	Animations: DefaultTheme.Animations,
	Variables: map[string]string{
		"cursor_style":      "→",
		"selected_style":    "[::u]",
		"progress_char":     "■",
		"progress_bg_char":  "□",
		"volume_char":       "●",
		"volume_bg_char":    "○",
		"border_style":      "single",
		"list_separator":    "─",
		"status_separator":  "│",
		"play_icon":         "▶",
		"pause_icon":        "⏸",
		"stop_icon":         "⏹",
		"next_icon":         "⏭",
		"prev_icon":         "⏮",
		"shuffle_icon":      "🔀",
		"repeat_icon":       "🔁",
		"volume_icon":       "🔊",
		"mute_icon":         "🔇",
		"heart_icon":        "♥",
		"star_icon":         "★",
	},
	IsDark:    false,
	IsDefault: false,
}

// GetAvailableThemes 获取可用主题列表
func GetAvailableThemes() []*ui.Theme {
	return []*ui.Theme{
		DefaultTheme,
		DarkTheme,
		LightTheme,
	}
}

// GetThemeByName 根据名称获取主题
func GetThemeByName(name string) *ui.Theme {
	themes := GetAvailableThemes()
	for _, theme := range themes {
		if strings.EqualFold(theme.Name, name) || strings.EqualFold(theme.ID, name) {
			return theme
		}
	}
	return DefaultTheme
}

// ApplyTheme 应用主题到组件
func ApplyTheme(component *ui.UIComponent, theme *ui.Theme) {
	if component == nil || theme == nil {
		return
	}

	// 应用主题属性到组件
	if component.Props == nil {
		component.Props = make(map[string]interface{})
	}

	// 应用颜色方案
	if theme.Colors != nil {
		component.Props["primary_color"] = theme.Colors.Primary
		component.Props["secondary_color"] = theme.Colors.Secondary
		component.Props["background_color"] = theme.Colors.Background
		component.Props["surface_color"] = theme.Colors.Surface
		component.Props["text_color"] = theme.Colors.Text
	}

	// 应用自定义属性
	for key, value := range theme.Variables {
		component.Props[key] = value
	}
}

// CreateCustomTheme 创建自定义主题
func CreateCustomTheme(name, description string, colorScheme *ui.ColorScheme, customProps map[string]string) *ui.Theme {
	theme := &ui.Theme{
		ID:          strings.ToLower(strings.ReplaceAll(name, " ", "-")),
		Name:        name,
		Description: description,
		Version:     "1.0.0",
		Author:      "user",
		Colors:      colorScheme,
		Fonts:       DefaultTheme.Fonts,
		Spacing:     DefaultTheme.Spacing,
		Borders:     DefaultTheme.Borders,
		Shadows:     DefaultTheme.Shadows,
		Animations:  DefaultTheme.Animations,
		Variables:   make(map[string]string),
		IsDark:      false,
		IsDefault:   false,
	}

	// 复制默认自定义属性
	for key, value := range DefaultTheme.Variables {
		theme.Variables[key] = value
	}

	// 应用用户自定义属性
	for key, value := range customProps {
		theme.Variables[key] = value
	}

	return theme
}