package commands

import (
	"fmt"
	"github.com/gookit/gcli/v2"
	"go-musicfox/pkg/constants"
	"go-musicfox/utils"
)

func NewConfigCommand() *gcli.Command {
	cmd := &gcli.Command{
		Name:   "config",
		UseFor: "Print configuration file to be loaded",
		Func: func(_ *gcli.Command, _ []string) error {
			fmt.Printf("\nLoaded Configuration File: %s/%s\n", utils.GetLocalDataDir(), constants.AppIniFile)
			return nil
		},
	}
	return cmd
}
