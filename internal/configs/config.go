package configs

import (
	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
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

	opts.AppName = types.AppName
	opts.WhetherDisplayTitle = cfg.Theme.ShowTitle
	opts.LoadingText = cfg.Theme.LoadingText
	opts.PrimaryColor = cfg.Theme.PrimaryColor
	opts.DualColumn = cfg.Theme.DoubleColumn
	opts.AltScreen = cfg.Main.AltScreen

	opts.MouseMode = tea.MouseModeCellMotion
	if !cfg.Main.EnableMouseEvent {
		opts.MouseMode = tea.MouseModeNone
	}
}
