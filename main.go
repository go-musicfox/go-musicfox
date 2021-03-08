package main

import (
	"fmt"
	"github.com/anhoder/go-musicfox/constants"
	"github.com/anhoder/go-musicfox/utils"
	"github.com/gookit/gcli/v2"
)

func main() {
	app := gcli.NewApp()
	app.Name = constants.AppName
	app.Version = constants.AppVersion
	app.Description = constants.AppDescription

	logo := utils.GetAlphaAscii(app.Name)
	randomColor := utils.GetRandomColor()

	gcli.AppHelpTemplate = fmt.Sprintf(constants.AppHelpTemplate, randomColor, utils.GetAlphaAscii(app.Name))
	app.Logo.Text = logo
	app.Logo.Style = randomColor

	app.Add(&gcli.Command{
		Name: "player",
		// allow color tag and {$cmd} will be replace to 'demo'
		UseFor: "this is a description <info>message</> for command",
		Aliases: []string{"dm"},
		Func: func (cmd *gcli.Command, args []string) error {
			gcli.Println("hello, in the demo command")
			return nil
		},
	})
	app.DefaultCommand("player")

	app.Run()
}