package main

import (
	"fmt"

	"github.com/anhoder/foxful-cli/util"
	neteaseutil "github.com/go-musicfox/netease-music/util"
	"github.com/gookit/gcli/v2"

	"github.com/go-musicfox/go-musicfox/internal/commands"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/runtime"
	"github.com/go-musicfox/go-musicfox/internal/types"
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
	}

	// 加载config
	filex.LoadIniConfig()

	util.PrimaryColor = configs.ConfigRegistry.Main.PrimaryColor
	util.MenuTitleColor = configs.ConfigRegistry.Main.MenuTitleColor
	util.ProgressColorExcludeRanges = configs.ConfigRegistry.Main.ProgressColorExcludeRanges
	util.ProgressColorSaturation = configs.ConfigRegistry.Main.ProgressColorSaturation
	util.ProgressColorBrightness = configs.ConfigRegistry.Main.ProgressColorBrightness
	var (
		logo         = util.GetAlphaAscii(app.Name)
		randomColor  = util.GetPrimaryColor()
		logoColorful = util.SetFgStyle(logo, randomColor)
	)

	gcli.AppHelpTemplate = fmt.Sprintf(types.AppHelpTemplate, logoColorful)
	app.Logo.Text = logoColorful

	// 更新netease配置
	neteaseutil.UNMSwitch = configs.ConfigRegistry.UNM.Enable
	neteaseutil.Sources = configs.ConfigRegistry.UNM.Sources
	neteaseutil.SearchLimit = configs.ConfigRegistry.UNM.SearchLimit
	neteaseutil.EnableLocalVip = configs.ConfigRegistry.UNM.EnableLocalVip
	neteaseutil.UnlockSoundEffects = configs.ConfigRegistry.UNM.UnlockSoundEffects

	playerCommand := commands.NewPlayerCommand()
	app.Add(playerCommand)
	app.Add(commands.NewConfigCommand())
	app.DefaultCommand(playerCommand.Name)

	app.Run()
}
