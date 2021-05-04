package main

import (
    "fmt"
    "github.com/google/gops/agent"
    "github.com/gookit/gcli/v2"
    "go-musicfox/commands"
    "go-musicfox/constants"
    "go-musicfox/ui"
    "go-musicfox/utils"
    "log"
)

func main() {
    if err := agent.Listen(agent.Options{}); err != nil {
        log.Fatalf("agent.Listen err: %v", err)
    }

    app := gcli.NewApp()
    app.Name = constants.AppName
    app.Version = constants.AppVersion
    app.Description = constants.AppDescription

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