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

	// 无障碍模式
	if cfg.Theme.AccessibleMode {
		style.SetAccessibleMode(true)
	}

	primary := util.GetPrimaryColor()

	if style.AccessibleMode() {
		opts.DarkTheme = style.HighContrastDarkTheme()
		opts.LightTheme = style.HighContrastLightTheme()
		opts.DarkTheme.Primary = primary
		opts.LightTheme.Primary = primary
	} else {
		// Use theme files if loaded; fallback to legacy primaryColor config
		registry := CurrentThemeRegistry()
		opts.DarkTheme, opts.LightTheme = cfg.Theme.modelThemesFromFiles(registry, primary)
	}

	opts.DualColumn = cfg.Theme.DoubleColumn
	opts.MaxMenuStartRow = cfg.Theme.MaxTitleStartRow
	opts.AltScreen = cfg.Main.AltScreen

	// 状态栏
	if cfg.Theme.StatusBar {
		opts.StatusBar = &model.DefaultStatusBar{}
	}
	if strings.EqualFold(cfg.Theme.StatusBarPosition, "bottom") {
		opts.StatusBarPosition = model.StatusBarBottom
	} else {
		opts.StatusBarPosition = model.StatusBarTop
	}

	// Toast 通知超时
	if cfg.Main.Notification.InAppTimeout > 0 {
		opts.NotificationOptions.DefaultTimeout = time.Duration(cfg.Main.Notification.InAppTimeout) * time.Second
	}

	// 鼠标事件
	opts.MouseMode = tea.MouseModeAllMotion
	if !cfg.Main.EnableMouseEvent {
		opts.MouseMode = tea.MouseModeNone
	}
}
