package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/anhoder/foxful-cli/util"
	neteaseutil "github.com/go-musicfox/netease-music/util"
	"github.com/gookit/gcli/v2"

	"github.com/go-musicfox/go-musicfox/internal/commands"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/runtime"
	"github.com/go-musicfox/go-musicfox/internal/types"
	mfoxapp "github.com/go-musicfox/go-musicfox/utils/app"
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
		gf.BoolOpt(&commands.GlobalOptions.PureMode, "pure", "", false, "start with default config in a temporary directory")
	}

	// --pure flag: start with a temporary directory as MUSICFOX_ROOT
	if isFlagTrue("pure") {
		pureDir, err := os.MkdirTemp("", "musicfox-pure-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create temp dir for pure mode: %v\n", err)
			os.Exit(1)
		}
		mfoxapp.SetupPureRoot(pureDir)
		commands.GlobalOptions.PureMode = true
		// 纯净模式下忽略会绕过 MUSICFOX_ROOT 的登录态环境变量，确保干净启动
		_ = os.Unsetenv("MUSICFOX_COOKIE")
		fmt.Fprintf(os.Stderr, "Pure mode enabled. Temp root: %s\n", pureDir)
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
	neteaseutil.UNMProxyURL = configs.AppConfig.UNM.ProxyURL

	playerCommand := commands.NewPlayerCommand()
	app.Add(playerCommand)
	app.Add(commands.NewConfigCommand())
	app.Add(commands.NewUpgradeConfigCommand())
	app.Add(commands.NewResetCommand())
	app.DefaultCommand(playerCommand.Name)

	app.Run()
}

// loadConfig 加载配置
func loadConfig() {
	configPath := mfoxapp.ConfigFilePath()

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

	// 加载主题文件（内置 + 用户自定义）
	configs.LoadThemeRegistry(mfoxapp.ConfigDir())
}

// isFlagTrue checks whether a boolean flag is set to true in os.Args,
// before gcli parses the flags. Handles --flag and --flag=true.
func isFlagTrue(name string) bool {
	prefix := "--" + name
	for _, arg := range os.Args {
		if arg == prefix {
			return true
		}
		if after, ok := strings.CutPrefix(arg, prefix+"="); ok {
			switch after {
			case "true", "1", "yes":
				return true
			}
		}
	}
	return false
}
