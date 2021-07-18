package main

import (
    "fmt"
    "github.com/gookit/gcli/v2"
    "go-musicfox/commands"
    "go-musicfox/config"
    "go-musicfox/constants"
    "go-musicfox/ui"
    "go-musicfox/utils"
)

func main() {
    app := gcli.NewApp()
    app.Name = constants.AppName
    app.Version = constants.AppVersion
    app.Description = constants.AppDescription

    // 加载config
    config.ConfigRegistry = utils.LoadIniConfig()

    logo := utils.GetAlphaAscii(app.Name)
    randomColor := ui.GetPrimaryColor()
    logoColorful := ui.SetFgStyle(logo, randomColor)

    gcli.AppHelpTemplate = fmt.Sprintf(constants.AppHelpTemplate, logoColorful)
    app.Logo.Text = logoColorful

    playerCommand := commands.NewPlayerCommand()
    app.Add(playerCommand)
    app.DefaultCommand(playerCommand.Name)

    app.Run()
}