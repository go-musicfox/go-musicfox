package configs

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// Config 是所有应用配置的根结构体
type Config struct {
	Startup     StartupConfig     `koanf:"startup"`
	Main        MainConfig        `koanf:"main"`
	Theme       ThemeConfig       `koanf:"theme"`
	Storage     StorageConfig     `koanf:"storage"`
	Player      PlayerConfig      `koanf:"player"`
	Autoplay    AutoplayConfig    `koanf:"autoplay"`
	UNM         UNMConfig         `koanf:"unm"`
	Reporter    ReporterConfig    `koanf:"reporter"`
	Keybindings KeybindingsConfig `koanf:"keybindings"`
	Share       map[string]string `koanf:"share"`
}

func (cfg *Config) FillToModelOpts(opts *model.Options) {
	opts.StartupOptions = cfg.Startup.ToModel()
	opts.ProgressOptions = cfg.Theme.Progress.ToModel()

	util.PrimaryColor = cfg.Theme.PrimaryColor
	opts.AppName = types.AppName
	opts.WhetherDisplayTitle = cfg.Theme.ShowTitle
	opts.LoadingText = cfg.Theme.LoadingText
	opts.PrimaryColor = cfg.Theme.PrimaryColor

	// 无障碍模式：启用高对比度主题 + 强调样式。
	// 未显式开启时，foxful-cli 仍会依据 NO_COLOR/ACCESSIBLE 环境变量自动探测。
	if cfg.Theme.AccessibleMode {
		style.SetAccessibleMode(true)
	}
	if style.AccessibleMode() {
		opts.DarkTheme = style.HighContrastDarkTheme()
		opts.LightTheme = style.HighContrastLightTheme()
		opts.DarkTheme.Primary = util.GetPrimaryColor()
		opts.LightTheme.Primary = util.GetPrimaryColor()
	} else {
		opts.DarkTheme, opts.LightTheme = cfg.Theme.modelThemes(util.GetPrimaryColor())
	}
	opts.DualColumn = cfg.Theme.DoubleColumn
	opts.MaxMenuStartRow = cfg.Theme.MaxTitleStartRow
	opts.AltScreen = cfg.Main.AltScreen

	// 状态栏（面包屑路径 + 时间）
	if cfg.Theme.StatusBar {
		opts.StatusBar = &model.DefaultStatusBar{}
	}
	// 状态栏位置：bottom 时置底部，其余（含空）默认顶部替换标题栏
	if strings.EqualFold(cfg.Theme.StatusBarPosition, "bottom") {
		opts.StatusBarPosition = model.StatusBarBottom
	} else {
		opts.StatusBarPosition = model.StatusBarTop
	}

	// TUI 内 toast 通知的自动消失时长
	if cfg.Main.Notification.InAppTimeout > 0 {
		opts.NotificationOptions.DefaultTimeout = time.Duration(cfg.Main.Notification.InAppTimeout) * time.Second
	}

	// hover 效果与指针形状依赖“无按钮移动”事件，必须用 AllMotion，
	// CellMotion 仅在按住按钮拖动时才发送 motion 事件，会导致 hover 失效。
	opts.MouseMode = tea.MouseModeAllMotion
	if !cfg.Main.EnableMouseEvent {
		opts.MouseMode = tea.MouseModeNone
	}
}
