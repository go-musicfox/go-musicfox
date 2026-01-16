package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/anhoder/foxful-cli/util"
	neteaseutil "github.com/go-musicfox/netease-music/util"
	"github.com/gookit/gcli/v2"

	"github.com/go-musicfox/go-musicfox/internal/commands"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/runtime"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/filex"
	_ "github.com/go-musicfox/go-musicfox/utils/slogx"
)

func main() {
	runtime.Run(musicfox)
}

func musicfox() {
	app := gcli.NewApp()
	app.Name = types.AppName
	app.Version = types.AppVersion
	if types.BuildTags != "" {
		app.Version += " [" + types.BuildTags + "]"
	}
	app.Description = types.AppDescription
	app.GOptsBinder = func(gf *gcli.Flags) {
		gf.BoolOpt(&commands.GlobalOptions.PProfMode, "pprof", "p", false, "enable PProf mode")
		gf.BoolOpt(&commands.GlobalOptions.DebugMode, "debug", "", false, "enable debug log level")
	}

	// FIXME: 后续版本移除
	if slices.Contains(os.Args, "migrate") {
		app.Add(commands.NewMigrateCommand())
		app.Run()
		return
	}

	loadConfig()

	util.PrimaryColor = configs.AppConfig.Theme.PrimaryColor
	var (
		logo         = util.GetAlphaAscii(app.Name)
		randomColor  = util.GetPrimaryColor()
		logoColorful = util.SetFgStyle(logo, randomColor)
	)

	gcli.AppHelpTemplate = fmt.Sprintf(types.AppHelpTemplate, logoColorful)
	app.Logo.Text = logoColorful

	// 更新netease配置
	neteaseutil.UNMSwitch = configs.AppConfig.UNM.Enable
	neteaseutil.Sources = configs.AppConfig.UNM.Sources
	neteaseutil.SearchLimit = configs.AppConfig.UNM.SearchLimit
	neteaseutil.EnableLocalVip = configs.AppConfig.UNM.EnableLocalVip
	neteaseutil.UnlockSoundEffects = configs.AppConfig.UNM.UnlockSoundEffects

	playerCommand := commands.NewPlayerCommand()
	app.Add(playerCommand)
	app.Add(commands.NewConfigCommand())
	app.Add(commands.NewMigrateCommand())
	app.DefaultCommand(playerCommand.Name)

	app.Run()
}

// loadConfig 加载配置
func loadConfig() {
	configDir := app.ConfigDir()
	configPath := app.ConfigFilePath()

	// 检测旧版 INI 配置文件（migrate 命令不需要检测）
	iniPath := filepath.Join(configDir, "go-musicfox.ini")
	if _, err := os.Stat(iniPath); err == nil {
		fmt.Fprintf(os.Stderr, "⚠ 检测到旧版 INI 配置文件\n")
		fmt.Fprintf(os.Stderr, "新版本不再支持 INI 格式的配置文件，请运行以下命令进行迁移：\n")
		fmt.Fprintf(os.Stderr, "  %s migrate\n\n", types.AppName)
		fmt.Fprintf(os.Stderr, "或手动删除旧配置文件后重新运行：\n")
		fmt.Fprintf(os.Stderr, "  rm %s\n\n", iniPath)
		fmt.Fprintf(os.Stderr, "迁移完成后，应用将使用新的 TOML 格式配置文件。\n")
		os.Exit(1)
	}

	// 如果配置文件不存在，从内嵌文件复制
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		_ = filex.CopyFileFromEmbed("embed/"+types.AppTomlFile, configPath)
	}

	// 加载 TOML 配置
	cfg, err := configs.NewConfigFromTomlFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("fatal: failed to load configuration: %v", err))
	}
	configs.AppConfig = cfg
}
