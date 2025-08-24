package commands

import (
	"fmt"

	"github.com/anhoder/foxful-cli/util"
	"github.com/gookit/gcli/v2"
	"github.com/muesli/termenv"

	"github.com/go-musicfox/go-musicfox/utils/app"
)

func NewConfigCommand() *gcli.Command {
	cmd := &gcli.Command{
		Name:   "config",
		UseFor: "Print configuration file to be loaded",
		Func: func(_ *gcli.Command, _ []string) error {
			var configPath = util.SetFgStyle(app.ConfigFilePath(), termenv.ANSICyan)
			fmt.Printf("Loaded Configuration File:\n\t%s\n", configPath)
			return nil
		},
	}
	return cmd
}
