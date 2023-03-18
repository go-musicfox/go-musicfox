package commands

import (
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/pkg/ui"
	"github.com/go-musicfox/go-musicfox/utils"

	tea "github.com/anhoder/bubbletea"
	"github.com/gookit/gcli/v2"
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

	http.DefaultClient.Timeout = constants.AppHttpTimeout
	neteaseModel := ui.NewNeteaseModel(configs.ConfigRegistry.StartupLoadingDuration)
	program := tea.NewProgram(neteaseModel)
	neteaseModel.BindProgram(program)
	if configs.ConfigRegistry.MainAltScreen {
		program.EnterAltScreen()
		defer program.ExitAltScreen()
	}
	return program.Start()
}
