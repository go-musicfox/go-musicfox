package configs

import (
	"image/color"

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
	// 主题色
	PrimaryColor string `koanf:"primaryColor"`
	// 是否显示状态栏（面包屑导航路径 + 时间）
	StatusBar bool `koanf:"statusBar"`
	// 状态栏位置：top（默认，顶部替换标题栏）或 bottom（底部）
	StatusBarPosition string `koanf:"statusBarPosition"`
	// 无障碍模式：高对比度主题 + 强调样式（留空跟随终端 NO_COLOR/ACCESSIBLE 环境变量自动探测）
	AccessibleMode bool `koanf:"accessibleMode"`

	Progress ProgressConfig `koanf:"progress"`
}

func (tc ThemeConfig) modelThemes(primary color.Color) (style.Theme, style.Theme) {
	dark := style.DefaultDarkTheme()
	dark.Primary = primary
	// 保持菜单标题为 BrightGreen，与旧版硬编码行为一致
	dark.MenuTitle = style.Highlight{Fg: lipgloss.BrightGreen}
	// 右键菜单浮层：默认继承 Surface(#242424) 与暗色终端背景几乎无法区分。
	// 提亮到 #383838 让浮层明显“浮起”。边框/分隔符用 #5C5C5C——
	// 比 Surface 明显亮、清晰可见，但仍是低调的中性灰，不抢眼。
	dark.Popup = style.PopupTheme{
		Surface: lipgloss.Color("#383838"),
		Border:  lipgloss.Color("#5C5C5C"),
	}

	light := style.DefaultLightTheme()
	light.Primary = primary
	// 保持菜单标题为 BrightGreen，与旧版硬编码行为一致
	light.MenuTitle = style.Highlight{Fg: lipgloss.BrightGreen}
	// 右键菜单浮层：默认 Surface(#F5F5F5) 与白色终端背景过于接近。
	// 压深到 #E8E8E8 拉开与背景的差异。边框/分隔符用 #A8A8A8——
	// 比 Surface 明显深、清晰可见，但仍是低调的中性灰，不抢眼。
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
