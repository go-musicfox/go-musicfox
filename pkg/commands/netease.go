package commands

import (
	tea "github.com/anhoder/bubbletea"
	"github.com/gookit/gcli/v2"
	"go-musicfox/pkg/configs"
	"go-musicfox/pkg/ui"
	"go-musicfox/utils"
	"net/http"
	"strconv"
)

func NewPlayerCommand() *gcli.Command {
	cmd := &gcli.Command{
		Name:   "netease",
		UseFor: "Command line player for Netease Cloud Music",
		Func:   runPlayer,
	}
	return cmd
}

func runPlayer(_ *gcli.Command, _ []string) error {
	if GlobalOptions.PProfMode {
		go func() {
			defer utils.Recover(true)
			panic(http.ListenAndServe(":"+strconv.Itoa(configs.ConfigRegistry.MainPProfPort), nil))
		}()
	}

	neteaseModel := ui.NewNeteaseModel(configs.ConfigRegistry.StartupLoadingDuration)
	program := tea.NewProgram(neteaseModel)
	neteaseModel.BindProgram(program)
	program.EnterAltScreen()
	defer program.ExitAltScreen()
	return program.Start()
}
