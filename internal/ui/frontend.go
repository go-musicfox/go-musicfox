package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// tuiFrontend adapts the TUI frontend to the frontend.Frontend contract.
type tuiFrontend struct{}

func (tuiFrontend) ID() string   { return "tui" }
func (tuiFrontend) Name() string { return "TUI" }

// Run assembles and runs the TUI frontend. The assembly was moved here
// verbatim from commands/netease.go (P1-2); order is preserved — global
// assignments like model.Submit / ui.SetupI18n must stay before NewNetease.
func (tuiFrontend) Run(ctx context.Context, launchOpts frontend.LaunchOptions) error {
	if launchOpts.Mode == frontend.ModeConnect {
		return RunConnect(ctx) // D-TC-1: remote shell, no engine is built
	}
	opts := model.DefaultOptions()
	configs.AppConfig.FillToModelOpts(opts)

	model.Submit = types.SubmitText
	model.SearchPlaceholder = types.SearchPlaceholder
	model.SearchResult = types.SearchResult
	SetupI18n(configs.AppConfig.Main.Locale)

	var (
		netease      = NewNetease(model.NewApp(opts))
		eventHandler = NewEventHandler(netease)
	)
	eventHandler.RegisterGlobalHotkeys(opts)
	netease.With(
		model.WithHook(netease.InitHook, netease.CloseHook),
		model.WithMainMenu(NewMainMenu(NewBaseMenu(netease)), &model.MenuItem{Title: "网易云音乐"}),
		// D-S1-1: a "view" command result is dispatched to the command_view
		// page. The page transition point must clear the cover image first
		// (AGENTS.md standalone-page rule); a nil page from buildPageOrToast
		// degrades to ignoring the message (toast already fired).
		model.WithUnknownMsgHandler(func(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
			vm, ok := msg.(commandViewMsg)
			if !ok {
				return nil, nil
			}
			netease.coverRenderer.ClearDisplayed()
			if page := buildPageOrToast("command_view", CommandViewOpts{Title: vm.Title, Lines: vm.Lines}); page != nil {
				return page, nil
			}
			return nil, nil
		}),
		func(options *model.Options) {
			options.TeaOptions = []tea.ProgramOption{
				tea.WithHardTabs(false),
			}
			options.LocalSearchMenu = NewLocalSearchMenu(NewBaseMenu(netease))
			options.Components = append(options.Components, netease.Components()...)
			options.KBControllers = append(options.KBControllers, eventHandler)
			options.MouseControllers = append(options.MouseControllers, eventHandler)
			options.Ticker = netease.Player().RenderTicker()
			options.DynamicRowCount = configs.AppConfig.Theme.DynamicMenuRows
			options.CenterEverything = configs.AppConfig.Theme.CenterEverything

			// 状态栏：若配置启用，注入队列位置与音质中间文本
			if options.StatusBar != nil {
				options.StatusBar = NewQueueQualityStatusBar(netease.Player())
			}

			if options.DynamicRowCount {
				// BottomHeight 是底部组件的最大预估高度，用于告诉 foxful-cli
				// 至少预留多少行给底部组件。菜单会根据剩余空间自适应调整行数。
				// 当窗口较小时，频谱会根据可用空间自动缩小或隐藏，
				// 歌词也会从 5 行缩减到 3 行或完全隐藏。
				// 详见 ui/spectrum_renderer.go 和 ui/lyric_renderer.go 的自适应逻辑。
				spectrumRows := 0
				if configs.AppConfig.Main.Visualizer.Enable {
					spectrumRows = DynamicMenuSpectrumLines // 最大预估频谱行数
				}
				options.BottomHeight = DynamicMenuOverhead + DynamicMenuLyricLines + spectrumRows + DynamicMenuBottomLines
			}
		},
	)

	return netease.Run()
}

func init() {
	frontend.Register(tuiFrontend{})
}
