//go:build darwin
// +build darwin

package entry

import (
	"fmt"
	"github.com/gookit/gcli/v2"
	"github.com/progrium/macdriver/cocoa"
	"go-musicfox/pkg/commands"
	"go-musicfox/pkg/constants"
	"go-musicfox/pkg/ui"
	"go-musicfox/utils"
)

func AppEntry() {
	go func() {
		app := gcli.NewApp()
		app.Name = constants.AppName
		app.Version = constants.AppVersion
		app.Description = constants.AppDescription

		// 加载config
		utils.LoadIniConfig()

		logo := utils.GetAlphaAscii(app.Name)
		randomColor := ui.GetPrimaryColor()
		logoColorful := ui.SetFgStyle(logo, randomColor)

		gcli.AppHelpTemplate = fmt.Sprintf(constants.AppHelpTemplate, logoColorful)
		app.Logo.Text = logoColorful

		playerCommand := commands.NewPlayerCommand()
		app.Add(playerCommand)
		app.DefaultCommand(playerCommand.Name)

		app.Run()
	}()

	app := cocoa.NSApp()
	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)
	app.Run()
}
