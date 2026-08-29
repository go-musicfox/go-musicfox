package ui

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/headless"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// remoteEventWireNames returns the core event-bus wire names the TUI connect
// shell subscribes to via headless.DialSubscribe. They are the raw core.Ev*
// constants (internal/core/events.go); the daemon filters subscription side
// frames by these names. They mirror the wire side of
// internal/webui/events.go eventWireToFrame — the WebUI only renames a subset
// to frame names, while the TUI shell subscribes to the full playback and
// lifecycle set (playback events, playlist end, rerender, login, startup
// phase). The event frames on the wire carry the original wire name, so the
// consumer can map them 1:1 back to core.Ev*.
func remoteEventWireNames() []string {
	return []string{
		core.EvSongChanged,
		core.EvStateChanged,
		core.EvPosition,
		core.EvPlaylistEnd,
		core.EvRerender,
		core.EvLogin,
		core.EvStartupPhase,
	}
}

// RunConnect runs the TUI as a remote shell (D-TC-1): it connects to the local
// headless daemon via DialSubscribe, probing for a live daemon, and returns a
// fail-fast error when none is running. No engine is built and no Startup
// sequence runs (B9) — the subscription session replaces the InitHook
// lifecycle (InitHook only finishes the shell assembly). The assembly mirrors
// tuiFrontend.Run's order; only the Player/User data surface and the renderer
// set are swapped.
//
// TUI-connect degradation summary (D-TC-3, roadmap §8.4):
//
//	播放控制（next/prev/pause/resume/toggle/stop/seek/volume/repeat/shuffle/
//	  like/dislike）      Call 转发 daemon ✅（PlaySong/选歌播放除外，见下）
//	播放状态/进度/模式/音量 订阅事件 + 快照缓存 ✅
//	播放队列显示           快照精简只读列表 △（本地建列表类操作降级 toast）
//	浏览/搜索（本地 API）   本地 API 照常；菜单由插件提供（connect 不加载）
//	                      → 入口 toast 降级（search 页面仍构建以保 shell 单例）
//	选歌播放               toast「遥控模式：daemon 不支持该操作」（P2:
//	                      daemon play_song 命令扩展）
//	登录                   daemon 登录态（status.user 昵称）；TUI 侧登录
//	                      禁用、需登录菜单 toast 降级（B8）
//	歌词/封面/频谱           隐藏/空渲染（B5/B6/B7，P2 扩展）
//	智能/心动模式           禁用 toast（P2 扩展）
//	命令面（轨 B / WASM）    禁用（B10：不加载 WASM、不注册命令菜单）
//	断线                   事件通道关闭 → ready=false + toast；不自动重连
//	                      （D-TC-4）
func RunConnect(ctx context.Context) error {
	client, err := headless.DialSubscribe(remoteEventWireNames())
	if err != nil {
		return fmt.Errorf("connect 模式需要 headless daemon 正在运行: %w", err)
	}
	defer client.Close()

	opts := model.DefaultOptions()
	configs.AppConfig.FillToModelOpts(opts)

	// Global assignments must stay before NewNeteaseRemote (align the
	// standalone assembly order, frontend.go).
	model.Submit = types.SubmitText
	model.SearchPlaceholder = types.SearchPlaceholder
	model.SearchResult = types.SearchResult
	SetupI18n(configs.AppConfig.Main.Locale)

	netease := NewNeteaseRemote(model.NewApp(opts), client)
	eventHandler := NewEventHandler(netease)
	eventHandler.RegisterGlobalHotkeys(opts)
	netease.With(
		model.WithHook(netease.InitHook, netease.CloseHook),
		model.WithMainMenu(NewMainMenu(NewBaseMenu(netease)), &model.MenuItem{Title: "网易云音乐"}),
		// D-S1-1: a "view" command result is dispatched to the command_view
		// page. In connect mode no commands are registered (B10), so this
		// never fires; the handler is kept for parity with the standalone
		// assembly. The page transition point must clear the cover first
		// (AGENTS.md standalone-page rule).
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
			// Resolved through the Player wrapper, which forwards to the
			// remote player's ticker in connect mode.
			options.Ticker = netease.Player().RenderTicker()
			options.DynamicRowCount = configs.AppConfig.Theme.DynamicMenuRows
			options.CenterEverything = configs.AppConfig.Theme.CenterEverything

			// 状态栏：若配置启用，注入队列位置与音质中间文本（读 remote 快照）。
			if options.StatusBar != nil {
				options.StatusBar = NewQueueQualityStatusBar(netease.Player())
			}

			if options.DynamicRowCount {
				// Lyric and spectrum renderers are skipped in connect mode
				// (B5/B7), so their bottom-height reserves are not needed.
				options.BottomHeight = DynamicMenuOverhead + DynamicMenuBottomLines
			}
		},
	)

	return netease.Run()
}

// registerConnectProviders registers the providers the TUI-connect shell needs
// beyond the built-in set. The frontend scope — which Starts the 9 business
// plugins and registers their menus/pages — is skipped in connect mode (B10;
// and the lastfm plugin's Deps requires engine services, so it cannot Start
// without an engine), so the shell re-provides only the providers whose types
// live in ui. The "search" page is one of them: the shell-owned singleton must
// stay non-nil because shared search-flow call sites (ToSearchPage /
// searchSong / SearchResultMenu hooks) dereference n.search unconditionally.
// The plugin-owned menus (search_type / search_result / detail jumps) cannot
// be re-provided from ui (they live in the plugins, and ui must not import
// them), so the menu-driven search flow degrades to toast — documented in the
// RunConnect degradation summary.
func registerConnectProviders() {
	if _, ok := pageRegistry["search"]; !ok {
		RegisterPage("search", func(opts SearchPageOpts) (model.Page, error) {
			return NewSearchPage(opts.Netease), nil
		})
	}
}

// connectCoverState adapts the remote player for the cover renderer in connect
// mode: it strips the album art URL so the cover degrades to an empty render
// (B6). The daemon snapshot carries no PicUrl in the MVP; the song_changed
// event does carry one, but the MVP cover surface stays hidden until P2 adds
// the field to the snapshot.
type connectCoverState struct{ p *Player }

func (s connectCoverState) CurSong() structs.Song {
	song := s.p.CurSong()
	song.PicUrl = ""
	return song
}
func (s connectCoverState) CurSongIndex() int         { return s.p.CurSongIndex() }
func (s connectCoverState) PassedTime() time.Duration { return s.p.PassedTime() }
func (s connectCoverState) State() types.State        { return s.p.State() }
func (s connectCoverState) Volume() int               { return s.p.Volume() }
func (s connectCoverState) Mode() types.Mode          { return s.p.Mode() }
func (s connectCoverState) Playlist() []structs.Song  { return s.p.Playlist() }
