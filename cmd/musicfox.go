package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	neteaseutil "github.com/go-musicfox/netease-music/util"
	"github.com/gookit/gcli/v2"

	"github.com/go-musicfox/go-musicfox/internal/commands"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins" // 编译期插件聚合器：触发各插件 init() 注册
	"github.com/go-musicfox/go-musicfox/internal/runtime"
	"github.com/go-musicfox/go-musicfox/internal/types"
	mfoxapp "github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/filex"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

func main() {
	// 以包装进程方式运行：崩溃（fatal error）时恢复终端并提示创建 issue。
	// 子进程内可恢复的 panic 由 errorx.Recover 自行处理。
	errorx.RunWrapped(func() {
		// 尽早初始化日志（幂等），确保后续 panic/崩溃可被捕获到日志文件
		slogx.Init()
		runtime.Run(musicfox)
	})
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

	// 主色：优先取活跃主题文件（theme 文件机制）的 primary；旧 config 字段
	// （deprecated，Kept for backward compatibility）仅作回退。
	util.PrimaryColor = configs.AppConfig.Theme.PrimaryColor //nolint:staticcheck // 旧字段向后兼容回退
	if tf, ok := configs.CurrentThemeRegistry().ActiveThemeOrDefault(configs.AppConfig.Theme.ActiveTheme); ok {
		variant := tf.Dark
		if !style.HasDarkBackground() {
			variant = tf.Light
		}
		if variant.Primary != "" {
			util.PrimaryColor = variant.Primary
		}
	}
	var (
		logo         = util.GetAlphaAscii(app.Name)
		randomColor  = util.GetPrimaryColor()
		logoColorful = style.FG(logo, randomColor)
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
