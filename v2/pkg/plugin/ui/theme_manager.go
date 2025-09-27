package ui

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"
)

// DefaultThemeManager 默认主题管理器
type DefaultThemeManager struct {
	currentTheme *Theme
	themes       map[string]*Theme
	mutex        sync.RWMutex
	logger       *slog.Logger

	// CSS生成器
	cssGenerator *CSSGenerator

	// 主题变量处理器
	variableProcessor *VariableProcessor

	// 主题历史
	themeHistory []ThemeHistoryEntry
	maxHistory   int

	// 自动主题切换
	autoThemeEnabled bool
	lightThemeID     string
	darkThemeID      string
}

// CSSGenerator CSS生成器
type CSSGenerator struct {
	logger *slog.Logger
}

// VariableProcessor 变量处理器
type VariableProcessor struct {
	logger *slog.Logger
}

// ThemeHistoryEntry 主题历史条目
type ThemeHistoryEntry struct {
	Theme     *Theme    `json:"theme"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason"`
}

// ThemeValidationResult 主题验证结果
type ThemeValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// NewDefaultThemeManager 创建默认主题管理器
func NewDefaultThemeManager(logger *slog.Logger) *DefaultThemeManager {
	manager := &DefaultThemeManager{
		themes:            make(map[string]*Theme),
		logger:            logger,
		themeHistory:      make([]ThemeHistoryEntry, 0),
		maxHistory:        50,
		autoThemeEnabled:  false,
		lightThemeID:      "light",
		darkThemeID:       "dark",
	}

	// 初始化组件
	manager.cssGenerator = NewCSSGenerator(logger)
	manager.variableProcessor = NewVariableProcessor(logger)

	// 注册默认主题
	manager.registerDefaultThemes()

	return manager
}

// registerDefaultThemes 注册默认主题
func (m *DefaultThemeManager) registerDefaultThemes() {
	// 浅色主题
	lightTheme := &Theme{
		ID:          "light",
		Name:        "Light Theme",
		Version:     "1.0.0",
		Description: "Default light theme",
		Author:      "go-musicfox",
		Colors: &ColorScheme{
			Primary:       "#1976d2",
			Secondary:     "#757575",
			Accent:        "#ff4081",
			Background:    "#ffffff",
			Surface:       "#f5f5f5",
			Text:          "#212121",
			TextSecondary: "#757575",
			Border:        "#e0e0e0",
			Error:         "#f44336",
			Warning:       "#ff9800",
			Success:       "#4caf50",
			Info:          "#2196f3",
		},
		Fonts: &FontScheme{
			Primary: &FontConfig{
				Family: "Inter, -apple-system, BlinkMacSystemFont, sans-serif",
				Size:   14,
				Weight: "400",
				Style:  "normal",
			},
			Secondary: &FontConfig{
				Family: "Inter, -apple-system, BlinkMacSystemFont, sans-serif",
				Size:   12,
				Weight: "400",
				Style:  "normal",
			},
			Monospace: &FontConfig{
				Family: "'JetBrains Mono', 'Fira Code', monospace",
				Size:   13,
				Weight: "400",
				Style:  "normal",
			},
		},
		Spacing: &SpacingScheme{
			XS: "4px",
			SM: "8px",
			MD: "16px",
			LG: "24px",
			XL: "32px",
		},
		Borders: &BorderScheme{
			Width:  "1px",
			Style:  "solid",
			Radius: "4px",
		},
		Shadows: &ShadowScheme{
			Small:  "0 1px 3px rgba(0,0,0,0.12), 0 1px 2px rgba(0,0,0,0.24)",
			Medium: "0 3px 6px rgba(0,0,0,0.16), 0 3px 6px rgba(0,0,0,0.23)",
			Large:  "0 10px 20px rgba(0,0,0,0.19), 0 6px 6px rgba(0,0,0,0.23)",
		},
		Animations: &AnimationScheme{
			Duration: "0.3s",
			Easing:   "ease-in-out",
			Delay:    "0s",
		},
		Variables: map[string]string{
			"--header-height": "60px",
			"--sidebar-width": "250px",
			"--footer-height": "40px",
		},
		IsDark:    false,
		IsDefault: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.themes[lightTheme.ID] = lightTheme
	m.currentTheme = lightTheme

	// 深色主题
	darkTheme := &Theme{
		ID:          "dark",
		Name:        "Dark Theme",
		Version:     "1.0.0",
		Description: "Default dark theme",
		Author:      "go-musicfox",
		Colors: &ColorScheme{
			Primary:       "#1976d2",
			Secondary:     "#424242",
			Accent:        "#ff4081",
			Background:    "#121212",
			Surface:       "#1e1e1e",
			Text:          "#ffffff",
			TextSecondary: "#b3b3b3",
			Border:        "#333333",
			Error:         "#f44336",
			Warning:       "#ff9800",
			Success:       "#4caf50",
			Info:          "#2196f3",
		},
		Fonts: lightTheme.Fonts, // 复用字体配置
		Spacing: lightTheme.Spacing, // 复用间距配置
		Borders: lightTheme.Borders, // 复用边框配置
		Shadows: &ShadowScheme{
			Small:  "0 1px 3px rgba(0,0,0,0.5), 0 1px 2px rgba(0,0,0,0.6)",
			Medium: "0 3px 6px rgba(0,0,0,0.6), 0 3px 6px rgba(0,0,0,0.7)",
			Large:  "0 10px 20px rgba(0,0,0,0.8), 0 6px 6px rgba(0,0,0,0.9)",
		},
		Animations: lightTheme.Animations, // 复用动画配置
		Variables:  lightTheme.Variables,  // 复用变量配置
		IsDark:     true,
		IsDefault:  false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	m.themes[darkTheme.ID] = darkTheme
}

// ApplyTheme 应用主题
func (m *DefaultThemeManager) ApplyTheme(ctx context.Context, theme *Theme) error {
	if theme == nil {
		return fmt.Errorf("theme cannot be nil")
	}

	// 验证主题
	if err := m.ValidateTheme(theme); err != nil {
		return fmt.Errorf("theme validation failed: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 记录主题历史
	if m.currentTheme != nil {
		m.addToHistory(m.currentTheme, "theme_change")
	}

	// 应用新主题
	m.currentTheme = theme
	m.themes[theme.ID] = theme

	m.logger.Info("Theme applied", "theme", theme.Name, "version", theme.Version)
	return nil
}

// GetCurrentTheme 获取当前主题
func (m *DefaultThemeManager) GetCurrentTheme() *Theme {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.currentTheme
}

// ValidateTheme 验证主题
func (m *DefaultThemeManager) ValidateTheme(theme *Theme) error {
	result := &ThemeValidationResult{
		Valid:    true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// 基本字段验证
	if theme.ID == "" {
		result.Errors = append(result.Errors, "theme ID cannot be empty")
		result.Valid = false
	}

	if theme.Name == "" {
		result.Errors = append(result.Errors, "theme name cannot be empty")
		result.Valid = false
	}

	if theme.Version == "" {
		result.Warnings = append(result.Warnings, "theme version is empty")
	}

	// 颜色方案验证
	if theme.Colors != nil {
		m.validateColorScheme(theme.Colors, result)
	} else {
		result.Errors = append(result.Errors, "color scheme is required")
		result.Valid = false
	}

	// 字体方案验证
	if theme.Fonts != nil {
		m.validateFontScheme(theme.Fonts, result)
	}

	// CSS验证
	if theme.CustomCSS != "" {
		m.validateCSS(theme.CustomCSS, result)
	}

	// 变量验证
	if len(theme.Variables) > 0 {
		m.validateVariables(theme.Variables, result)
	}

	// 如果验证失败，返回错误
	if !result.Valid {
		return fmt.Errorf("theme validation failed: %v", result.Errors)
	}

	// 记录警告
	if len(result.Warnings) > 0 {
		m.logger.Warn("Theme validation warnings", "warnings", result.Warnings)
	}

	return nil
}

// validateColorScheme 验证颜色方案
func (m *DefaultThemeManager) validateColorScheme(colors *ColorScheme, result *ThemeValidationResult) {
	colorFields := map[string]string{
		"primary":        colors.Primary,
		"secondary":      colors.Secondary,
		"accent":         colors.Accent,
		"background":     colors.Background,
		"surface":        colors.Surface,
		"text":           colors.Text,
		"text_secondary": colors.TextSecondary,
		"border":         colors.Border,
		"error":          colors.Error,
		"warning":        colors.Warning,
		"success":        colors.Success,
		"info":           colors.Info,
	}

	colorRegex := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$|^rgb\(|^rgba\(|^hsl\(|^hsla\(`)

	for field, color := range colorFields {
		if color == "" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("color field '%s' is empty", field))
			continue
		}

		if !colorRegex.MatchString(color) {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid color format for field '%s': %s", field, color))
			result.Valid = false
		}
	}
}

// validateFontScheme 验证字体方案
func (m *DefaultThemeManager) validateFontScheme(fonts *FontScheme, result *ThemeValidationResult) {
	if fonts.Primary != nil {
		m.validateFontConfig("primary", fonts.Primary, result)
	}
	if fonts.Secondary != nil {
		m.validateFontConfig("secondary", fonts.Secondary, result)
	}
	if fonts.Monospace != nil {
		m.validateFontConfig("monospace", fonts.Monospace, result)
	}
	if fonts.Display != nil {
		m.validateFontConfig("display", fonts.Display, result)
	}
}

// validateFontConfig 验证字体配置
func (m *DefaultThemeManager) validateFontConfig(name string, font *FontConfig, result *ThemeValidationResult) {
	if font.Family == "" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("font family for '%s' is empty", name))
	}

	if font.Size <= 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("font size for '%s' must be positive", name))
		result.Valid = false
	}

	validWeights := []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "normal", "bold", "lighter", "bolder"}
	validWeight := false
	for _, weight := range validWeights {
		if font.Weight == weight {
			validWeight = true
			break
		}
	}
	if !validWeight {
		result.Warnings = append(result.Warnings, fmt.Sprintf("unusual font weight for '%s': %s", name, font.Weight))
	}
}

// validateCSS 验证CSS
func (m *DefaultThemeManager) validateCSS(css string, result *ThemeValidationResult) {
	// 简单的CSS语法检查
	if strings.Contains(css, "<script") || strings.Contains(css, "javascript:") {
		result.Errors = append(result.Errors, "CSS contains potentially dangerous content")
		result.Valid = false
	}

	// 检查CSS语法基本结构
	openBraces := strings.Count(css, "{")
	closeBraces := strings.Count(css, "}")
	if openBraces != closeBraces {
		result.Errors = append(result.Errors, "CSS has unmatched braces")
		result.Valid = false
	}
}

// validateVariables 验证变量
func (m *DefaultThemeManager) validateVariables(variables map[string]string, result *ThemeValidationResult) {
	for key, value := range variables {
		if !strings.HasPrefix(key, "--") {
			result.Warnings = append(result.Warnings, fmt.Sprintf("CSS variable '%s' should start with '--'", key))
		}

		if value == "" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("CSS variable '%s' has empty value", key))
		}
	}
}

// GenerateCSS 生成CSS
func (m *DefaultThemeManager) GenerateCSS(theme *Theme) (string, error) {
	if theme == nil {
		return "", fmt.Errorf("theme cannot be nil")
	}

	return m.cssGenerator.Generate(theme)
}

// RegisterTheme 注册主题
func (m *DefaultThemeManager) RegisterTheme(theme *Theme) error {
	if err := m.ValidateTheme(theme); err != nil {
		return fmt.Errorf("theme validation failed: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.themes[theme.ID] = theme
	m.logger.Info("Theme registered", "id", theme.ID, "name", theme.Name)
	return nil
}

// GetTheme 获取主题
func (m *DefaultThemeManager) GetTheme(themeID string) (*Theme, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	theme, exists := m.themes[themeID]
	if !exists {
		return nil, fmt.Errorf("theme not found: %s", themeID)
	}

	return theme, nil
}

// ListThemes 列出所有主题
func (m *DefaultThemeManager) ListThemes() []*Theme {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	themes := make([]*Theme, 0, len(m.themes))
	for _, theme := range m.themes {
		themes = append(themes, theme)
	}

	return themes
}

// EnableAutoTheme 启用自动主题切换
func (m *DefaultThemeManager) EnableAutoTheme(lightThemeID, darkThemeID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查主题是否存在
	if _, exists := m.themes[lightThemeID]; !exists {
		return fmt.Errorf("light theme not found: %s", lightThemeID)
	}

	if _, exists := m.themes[darkThemeID]; !exists {
		return fmt.Errorf("dark theme not found: %s", darkThemeID)
	}

	m.autoThemeEnabled = true
	m.lightThemeID = lightThemeID
	m.darkThemeID = darkThemeID

	m.logger.Info("Auto theme enabled", "light", lightThemeID, "dark", darkThemeID)
	return nil
}

// DisableAutoTheme 禁用自动主题切换
func (m *DefaultThemeManager) DisableAutoTheme() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.autoThemeEnabled = false
	m.logger.Info("Auto theme disabled")
}

// SwitchToSystemTheme 切换到系统主题
func (m *DefaultThemeManager) SwitchToSystemTheme(isDark bool) error {
	if !m.autoThemeEnabled {
		return fmt.Errorf("auto theme is not enabled")
	}

	themeID := m.lightThemeID
	if isDark {
		themeID = m.darkThemeID
	}

	theme, exists := m.themes[themeID]
	if !exists {
		return fmt.Errorf("theme not found: %s", themeID)
	}

	return m.ApplyTheme(context.Background(), theme)
}

// addToHistory 添加到主题历史
func (m *DefaultThemeManager) addToHistory(theme *Theme, reason string) {
	entry := ThemeHistoryEntry{
		Theme:     theme,
		Timestamp: time.Now(),
		Reason:    reason,
	}

	m.themeHistory = append(m.themeHistory, entry)

	// 限制历史长度
	if len(m.themeHistory) > m.maxHistory {
		m.themeHistory = m.themeHistory[1:]
	}
}

// GetThemeHistory 获取主题历史
func (m *DefaultThemeManager) GetThemeHistory() []ThemeHistoryEntry {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	history := make([]ThemeHistoryEntry, len(m.themeHistory))
	copy(history, m.themeHistory)
	return history
}

// NewCSSGenerator 创建CSS生成器
func NewCSSGenerator(logger *slog.Logger) *CSSGenerator {
	return &CSSGenerator{
		logger: logger,
	}
}

// Generate 生成CSS
func (g *CSSGenerator) Generate(theme *Theme) (string, error) {
	var css strings.Builder

	// 生成CSS变量
	css.WriteString(":root {\n")

	// 颜色变量
	if theme.Colors != nil {
		css.WriteString(fmt.Sprintf("  --color-primary: %s;\n", theme.Colors.Primary))
		css.WriteString(fmt.Sprintf("  --color-secondary: %s;\n", theme.Colors.Secondary))
		css.WriteString(fmt.Sprintf("  --color-accent: %s;\n", theme.Colors.Accent))
		css.WriteString(fmt.Sprintf("  --color-background: %s;\n", theme.Colors.Background))
		css.WriteString(fmt.Sprintf("  --color-surface: %s;\n", theme.Colors.Surface))
		css.WriteString(fmt.Sprintf("  --color-text: %s;\n", theme.Colors.Text))
		css.WriteString(fmt.Sprintf("  --color-text-secondary: %s;\n", theme.Colors.TextSecondary))
		css.WriteString(fmt.Sprintf("  --color-border: %s;\n", theme.Colors.Border))
		css.WriteString(fmt.Sprintf("  --color-error: %s;\n", theme.Colors.Error))
		css.WriteString(fmt.Sprintf("  --color-warning: %s;\n", theme.Colors.Warning))
		css.WriteString(fmt.Sprintf("  --color-success: %s;\n", theme.Colors.Success))
		css.WriteString(fmt.Sprintf("  --color-info: %s;\n", theme.Colors.Info))
	}

	// 字体变量
	if theme.Fonts != nil {
		if theme.Fonts.Primary != nil {
			css.WriteString(fmt.Sprintf("  --font-family-primary: %s;\n", theme.Fonts.Primary.Family))
			css.WriteString(fmt.Sprintf("  --font-size-primary: %dpx;\n", theme.Fonts.Primary.Size))
			css.WriteString(fmt.Sprintf("  --font-weight-primary: %s;\n", theme.Fonts.Primary.Weight))
		}
		if theme.Fonts.Monospace != nil {
			css.WriteString(fmt.Sprintf("  --font-family-monospace: %s;\n", theme.Fonts.Monospace.Family))
		}
	}

	// 间距变量
	if theme.Spacing != nil {
		css.WriteString(fmt.Sprintf("  --spacing-xs: %s;\n", theme.Spacing.XS))
		css.WriteString(fmt.Sprintf("  --spacing-sm: %s;\n", theme.Spacing.SM))
		css.WriteString(fmt.Sprintf("  --spacing-md: %s;\n", theme.Spacing.MD))
		css.WriteString(fmt.Sprintf("  --spacing-lg: %s;\n", theme.Spacing.LG))
		css.WriteString(fmt.Sprintf("  --spacing-xl: %s;\n", theme.Spacing.XL))
	}

	// 边框变量
	if theme.Borders != nil {
		css.WriteString(fmt.Sprintf("  --border-width: %s;\n", theme.Borders.Width))
		css.WriteString(fmt.Sprintf("  --border-style: %s;\n", theme.Borders.Style))
		css.WriteString(fmt.Sprintf("  --border-radius: %s;\n", theme.Borders.Radius))
	}

	// 阴影变量
	if theme.Shadows != nil {
		css.WriteString(fmt.Sprintf("  --shadow-small: %s;\n", theme.Shadows.Small))
		css.WriteString(fmt.Sprintf("  --shadow-medium: %s;\n", theme.Shadows.Medium))
		css.WriteString(fmt.Sprintf("  --shadow-large: %s;\n", theme.Shadows.Large))
	}

	// 动画变量
	if theme.Animations != nil {
		css.WriteString(fmt.Sprintf("  --animation-duration: %s;\n", theme.Animations.Duration))
		css.WriteString(fmt.Sprintf("  --animation-easing: %s;\n", theme.Animations.Easing))
		css.WriteString(fmt.Sprintf("  --animation-delay: %s;\n", theme.Animations.Delay))
	}

	// 自定义变量
	for key, value := range theme.Variables {
		css.WriteString(fmt.Sprintf("  %s: %s;\n", key, value))
	}

	css.WriteString("}\n\n")

	// 基础样式
	css.WriteString("body {\n")
	css.WriteString("  background-color: var(--color-background);\n")
	css.WriteString("  color: var(--color-text);\n")
	if theme.Fonts != nil && theme.Fonts.Primary != nil {
		css.WriteString("  font-family: var(--font-family-primary);\n")
		css.WriteString("  font-size: var(--font-size-primary);\n")
		css.WriteString("  font-weight: var(--font-weight-primary);\n")
	}
	css.WriteString("}\n\n")

	// 添加自定义CSS
	if theme.CustomCSS != "" {
		css.WriteString(theme.CustomCSS)
		css.WriteString("\n")
	}

	return css.String(), nil
}

// NewVariableProcessor 创建变量处理器
func NewVariableProcessor(logger *slog.Logger) *VariableProcessor {
	return &VariableProcessor{
		logger: logger,
	}
}

// ProcessVariables 处理变量
func (p *VariableProcessor) ProcessVariables(css string, variables map[string]string) string {
	result := css

	// 替换变量
	for key, value := range variables {
		placeholder := fmt.Sprintf("var(%s)", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// ExtractVariables 提取变量
func (p *VariableProcessor) ExtractVariables(css string) map[string]string {
	variables := make(map[string]string)

	// 使用正则表达式提取CSS变量
	varRegex := regexp.MustCompile(`(--[\w-]+):\s*([^;]+);`)
	matches := varRegex.FindAllStringSubmatch(css, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			variables[key] = value
		}
	}

	return variables
}