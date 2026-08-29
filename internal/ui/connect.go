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
// TUI-connect capability summary (roadmap §8.4/§8.10; TC-5..TC-8 land the
// full login + playback extension — S6-R1 restores the local browsing tree by
// mounting the engine-independent business plugins):
//
//	播放控制/状态/进度/模式/音量  Call 转发 + 订阅快照 ✅
//	浏览/搜索（本地 API）   本地照常 ✅（S6-R1：8 个 engine 无关业务插件已
//	                      挂载，lastfm 除外——其 Deps 依赖 engine 服务，
//	                      connect 无 engine；搜索/排行/精选歌单/专辑/歌手/
//	                      DJ 免登录浏览可用，菜单完整）
//	登录                   daemon 侧 QR 登录、TUI 遥控扫码 ✅（ToLoginPage
//	                      → connect LoginPage 只显 QR 入口 → CallQRKey/
//	                      CallQRStatus 数据源；EvLogin 驱动用户态刷新，
//	                      D-TC-7/TC-6）
//	用户态展示               昵称 + UserId（status.userId 快照幂等 + EvLogin
//	                      增量；门控仍剥离 UserId，D-TC-8）
//	选歌播放                 play_list 整列表投递 daemon ✅（PlaySong/
//	                      ReinitializePlaylist/StartPlay 遮蔽升级，响应
//	                      playlist 写回缓存同步队列，D-TC-9/TC-7）
//	播放队列显示             快照精简只读列表 △ + 投递后响应同步（next/prev
//	                      正确）
//	需登录浏览               收藏/我的歌单/云盘/每日推荐/最近播放/私人FM
//	                      无条件 toast 降级（本地进程无 cookie，D-TC-7
//	                      边界；P2：cookie 回拉）
//	播放队列编辑/智能/心动     toast 降级（P2：daemon 队列编辑命令）
//	歌词/封面/频谱           隐藏/空渲染（B5/B6/B7，P2 扩展）
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

// registerConnectProviders is the fallback guard for the "search" page
// provider in the TUI-connect shell (S6-R1). The provider is normally
// registered by the search plugin's Start inside the connect frontend scope;
// this only fires when [plugins] disabled contains "search" — the plugin is
// not mounted, so the shell re-provides the page to keep the shared
// search-flow call sites (ToSearchPage / searchSong / SearchResultMenu hooks)
// holding a non-nil singleton. Idempotent: it skips when the provider already
// exists.
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
