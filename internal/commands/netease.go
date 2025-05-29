package commands

import (
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/gookit/gcli/v2"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
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
		errorx.Go(func() {
			panic(http.ListenAndServe(":"+strconv.Itoa(configs.ConfigRegistry.Main.PProfPort), nil))
		}, true)
	}

	http.DefaultClient.Timeout = types.AppHttpTimeout
	runewidth.DefaultCondition.EastAsianWidth = false

	opts := model.DefaultOptions()
	configs.ConfigRegistry.FillToModelOpts(opts)

	model.Submit = types.SubmitText
	model.SearchPlaceholder = types.SearchPlaceholder
	model.SearchResult = types.SearchResult

	var (
		netease      = ui.NewNetease(model.NewApp(opts))
		eventHandler = ui.NewEventHandler(netease)
	)
	eventHandler.RegisterGlobalHotkeys(opts)
	netease.With(
		model.WithHook(netease.InitHook, netease.CloseHook),
		model.WithMainMenu(ui.NewMainMenu(netease), &model.MenuItem{Title: "网易云音乐"}),
		func(options *model.Options) {
			options.LocalSearchMenu = ui.NewLocalSearchMenu(netease)
			options.Components = append(options.Components, netease.Player())
			options.KBControllers = append(options.KBControllers, eventHandler)
			options.MouseControllers = append(options.MouseControllers, eventHandler)
			options.Ticker = netease.Player().RenderTicker()
			options.DynamicRowCount = configs.ConfigRegistry.Main.DynamicMenuRows
			options.CenterEverything = configs.ConfigRegistry.Main.CenterEverything
		},
	)

	return netease.Run()
}
