package commands

import (
	"fmt"

	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/gookit/gcli/v2"
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
