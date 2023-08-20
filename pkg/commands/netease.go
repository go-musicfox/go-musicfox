package commands

import (
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/pkg/ui"
	"github.com/go-musicfox/go-musicfox/utils"

	tea "github.com/charmbracelet/bubbletea"
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
		go utils.PanicRecoverWrapper(true, func() {
			panic(http.ListenAndServe(":"+strconv.Itoa(configs.ConfigRegistry.MainPProfPort), nil))
		})
	}

	http.DefaultClient.Timeout = constants.AppHttpTimeout
	neteaseModel := ui.NewNeteaseModel(configs.ConfigRegistry.StartupLoadingDuration)

	var opts []tea.ProgramOption
	if configs.ConfigRegistry.MainEnableMouseEvent {
		opts = append(opts, tea.WithMouseCellMotion())
	}
	if configs.ConfigRegistry.MainAltScreen {
		opts = append(opts, tea.WithAltScreen())
	}
	program := tea.ReplaceWithFoxfulRenderer(tea.NewProgram(neteaseModel, opts...))
	neteaseModel.BindProgram(program)

	_, err := program.Run()
	return err
}
