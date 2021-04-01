package main

import (
    "fmt"
    "github.com/gookit/gcli/v2"
    "go-musicfox/commands"
    "go-musicfox/constants"
    "go-musicfox/ui"
    "go-musicfox/utils"
)

func main() {
    app := gcli.NewApp()
    app.Name = constants.AppName
    app.Version = constants.AppVersion
    app.Description = constants.AppDescription

    logo := utils.GetAlphaAscii(app.Name)
    randomColor := ui.GetRandomLogoColor()
    logoColorful := ui.SetFgStyle(logo, randomColor)

    gcli.AppHelpTemplate = fmt.Sprintf(constants.AppHelpTemplate, logoColorful)
    app.Logo.Text = logoColorful

    playerCommand := commands.NewPlayerCommand()
    app.Add(playerCommand)
    app.DefaultCommand(playerCommand.Name)

    app.Run()
}