package configs

import (
	"image/color"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func firstCharOrDefault(s, defaultStr string) rune {
	if len(s) > 0 {
		return []rune(s)[0]
	}
	return []rune(defaultStr)[0]
}

type ProgressOptions struct {
	model.ProgressOptions
}

// ThemeConfig 主题设置
type ThemeConfig struct {
	// 活跃主题名称（对应主题文件名中的 name 字段，默认 "NetEase Red"）
	ActiveTheme string `koanf:"activeTheme"`
	// Deprecated: migrated to theme files. Kept for backward compatibility.
	PrimaryColor string `koanf:"primaryColor"`

	// 主界面是否显示标题
	ShowTitle bool `koanf:"showTitle"`
	// 主页面加载中提示
	LoadingText string `koanf:"loadingText"`
	// 是否双列显示
	DoubleColumn bool `koanf:"doubleColumn"`
	// 菜单行数动态变更
	DynamicMenuRows bool `koanf:"dynamicMenuRows"`
	// 菜单内容起始行上限，限制菜单上方标题区域向下的最大偏移量（全局生效，0 不限制）
	MaxTitleStartRow int `koanf:"maxTitleStartRow"`
	// 界面全部居中
	CenterEverything bool `koanf:"centerEverything"`
	// 是否显示状态栏（面包屑导航路径 + 时间）
	StatusBar bool `koanf:"statusBar"`
	// 状态栏位置
	StatusBarPosition string `koanf:"statusBarPosition"`
	// 无障碍模式
	AccessibleMode bool `koanf:"accessibleMode"`

	Progress ProgressConfig `koanf:"progress"`
}

// ThemeRegistry holds all loaded themes and provides runtime theme switching support.
type ThemeRegistry struct {
	mu          sync.RWMutex
	themes      map[string]*ThemeFile
	allNames    []string        // sorted list of all theme names
	darkNames   []string        // themes with dark variant configured
	lightNames  []string        // themes with light variant configured
	themeIndex  int             // current index in the active brightness-category theme list
}

var globalThemeRegistry = &ThemeRegistry{
	themes: make(map[string]*ThemeFile),
}

// LoadThemeRegistry loads themes and populates the global registry.
func LoadThemeRegistry(userConfigDir string) {
	globalThemeRegistry.mu.Lock()
	defer globalThemeRegistry.mu.Unlock()

	globalThemeRegistry.themes = LoadAllThemes(userConfigDir)
	globalThemeRegistry.rebuildIndex()
}

func (r *ThemeRegistry) rebuildIndex() {
	r.allNames = make([]string, 0, len(r.themes))
	r.darkNames = nil
	r.lightNames = nil
	for name, tf := range r.themes {
		r.allNames = append(r.allNames, name)
		if tf.Dark.isConfigured() {
			r.darkNames = append(r.darkNames, name)
		}
		if tf.Light.isConfigured() {
			r.lightNames = append(r.lightNames, name)
		}
	}
}

// Get returns a theme by name.
func (r *ThemeRegistry) Get(name string) (*ThemeFile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tf, ok := r.themes[name]
	return tf, ok
}

// First returns the first available theme.
func (r *ThemeRegistry) First() (*ThemeFile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, tf := range r.themes {
		return tf, true
	}
	return nil, false
}

// Names returns a copy of sorted theme names (all, dark, light).
func (r *ThemeRegistry) Names(darkBackground bool) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if darkBackground {
		cp := make([]string, len(r.darkNames))
		copy(cp, r.darkNames)
		return cp
	}
	cp := make([]string, len(r.lightNames))
	copy(cp, r.lightNames)
	return cp
}

// ActiveThemeOrDefault returns the active theme file, falling back to any available.
func (r *ThemeRegistry) ActiveThemeOrDefault(name string) (*ThemeFile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if tf, ok := r.themes[name]; ok {
		return tf, ok
	}
	return r.First()
}

// Current returns the global theme registry.
func CurrentThemeRegistry() *ThemeRegistry {
	return globalThemeRegistry
}

// NextStyleSet cycles to the next theme in the list for the given brightness and returns a new StyleSet.
// If no themes are available or the list is empty, returns nil.
func (r *ThemeRegistry) NextStyleSet(darkBackground bool) *style.StyleSet {
	r.mu.Lock()
	defer r.mu.Unlock()

	names := r.darkNames
	if !darkBackground {
		names = r.lightNames
	}
	if len(names) == 0 {
		return nil
	}

	r.themeIndex = (r.themeIndex + 1) % len(names)
	name := names[r.themeIndex]
	tf := r.themes[name]
	var t style.Theme
	if darkBackground {
		t = tf.Dark.toTheme()
	} else {
		t = tf.Light.toTheme()
	}
	ss := style.NewStyleSet(t)
	return &ss
}

// CurrentTheme returns the current active theme and StyleSet based on brightness.
func (r *ThemeRegistry) CurrentStyleSet(darkBackground bool) *style.StyleSet {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := r.darkNames
	if !darkBackground {
		names = r.lightNames
	}
	if len(names) == 0 {
		return nil
	}
	if r.themeIndex >= len(names) {
		r.themeIndex = 0
	}
	name := names[r.themeIndex]
	tf := r.themes[name]
	var t style.Theme
	if darkBackground {
		t = tf.Dark.toTheme()
	} else {
		t = tf.Light.toTheme()
	}
	ss := style.NewStyleSet(t)
	return &ss
}

// modelThemesFromFiles builds foxful-cli Theme pair from the active theme file.
// Falls back to legacy behavior using primaryColor if no theme files are loaded.
func (tc ThemeConfig) modelThemesFromFiles(themes *ThemeRegistry, primary color.Color) (style.Theme, style.Theme) {
	tf, ok := themes.ActiveThemeOrDefault(tc.ActiveTheme)
	if ok && tf != nil {
		dark := tf.Dark.toTheme()
		light := tf.Light.toTheme()
		return dark, light
	}
	return tc.modelThemesLegacy(primary)
}

// modelThemesLegacy builds themes from the old primaryColor config for backward compatibility.
func (tc ThemeConfig) modelThemesLegacy(primary color.Color) (style.Theme, style.Theme) {
	dark := style.DefaultDarkTheme()
	dark.Primary = primary
	dark.MenuTitle = style.Highlight{Fg: lipgloss.BrightGreen}
	dark.Popup = style.PopupTheme{
		Surface: lipgloss.Color("#383838"),
		Border:  lipgloss.Color("#5C5C5C"),
	}

	light := style.DefaultLightTheme()
	light.Primary = primary
	light.MenuTitle = style.Highlight{Fg: lipgloss.BrightGreen}
	light.Popup = style.PopupTheme{
		Surface: lipgloss.Color("#E8E8E8"),
		Border:  lipgloss.Color("#A8A8A8"),
	}

	return dark, light
}

// ProgressConfig 进度条字符样式配置
type ProgressConfig struct {
	RenderMode string `koanf:"renderMode"`

	FullChar           string `koanf:"fullChar"`
	FullCharWhenFirst  string `koanf:"fullCharWhenFirst"`
	FullCharWhenLast   string `koanf:"fullCharWhenLast"`
	LastFullChar       string `koanf:"lastFullChar"`
	EmptyChar          string `koanf:"emptyChar"`
	EmptyCharWhenFirst string `koanf:"emptyCharWhenFirst"`
	EmptyCharWhenLast  string `koanf:"emptyCharWhenLast"`
	FirstEmptyChar     string `koanf:"firstEmptyChar"`
}

// ToModel 将 ProgressConfig 转换为 foxful-cli 所需的 model.ProgressOptions。
func (pc ProgressConfig) ToModel() model.ProgressOptions {
	return model.ProgressOptions{
		FullChar:           firstCharOrDefault(pc.FullChar, types.ProgressFullChar),
		FullCharWhenFirst:  firstCharOrDefault(pc.FullCharWhenFirst, types.ProgressFullChar),
		FullCharWhenLast:   firstCharOrDefault(pc.FullCharWhenLast, types.ProgressFullChar),
		LastFullChar:       firstCharOrDefault(pc.LastFullChar, types.ProgressFullChar),
		EmptyChar:          firstCharOrDefault(pc.EmptyChar, types.ProgressEmptyChar),
		EmptyCharWhenFirst: firstCharOrDefault(pc.EmptyCharWhenFirst, types.ProgressEmptyChar),
		EmptyCharWhenLast:  firstCharOrDefault(pc.EmptyCharWhenLast, types.ProgressEmptyChar),
		FirstEmptyChar:     firstCharOrDefault(pc.FirstEmptyChar, types.ProgressEmptyChar),
	}
}
