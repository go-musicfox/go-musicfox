package commands

import (
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/pkg/ui"
	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/mattn/go-runewidth"

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
	runewidth.DefaultCondition.EastAsianWidth = false

	var opts = model.DefaultOptions()
	configs.ConfigRegistry.FillToModelOpts(opts)

	var (
		netease      = ui.NewNetease(model.NewApp(opts))
		eventHandler = ui.NewEventHandler(netease)
	)
	netease.App.With(
		model.WithHook(netease.InitHook, netease.CloseHook),
		model.WithMainMenu(ui.NewMainMenu(netease), &model.MenuItem{Title: "网易云音乐"}),
		func(options *model.Options) {
			options.Components = append(options.Components, netease.Player())
			options.KBControllers = append(options.KBControllers, eventHandler)
			options.MouseControllers = append(options.MouseControllers, eventHandler)
		},
	)

	return netease.Run()
}
