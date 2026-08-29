// Package ui implements the foxful-cli TUI layer: the thin-shell Netease
// coordinator (navigation/assembly/event dispatch), provider-registered menus
// and pages, and the lyric/cover/spectrum renderers composed by the shell.
// Business capabilities are resolved by name through internal/framework.
package ui

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	uv "github.com/charmbracelet/ultraviolet"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/headless"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/filex"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/version"
)

// connectMode marks the TUI-connect remote-shell assembly (S6). It relaxes the
// main-menu after-anchor chain assertions (the 9 business plugins never Start,
// so their anchor keys are absent and the built-in 帮助 entry's after=last_fm
// anchor legitimately goes missing) and is set by the two shell constructors:
// NewNeteaseRemote → true, NewNetease → false. Package-level because the menu
// chain walk (orderMainMenuEntries) has no shell context.
var connectMode bool

type Netease struct {
	user *structs.User

	// engine owns the UI-free service assembly, startup sequence and cleanup
	// (internal/core). The shell shares its user slot with the engine and wraps
	// the core player with the TUI observer/loading/locator seams below.
	engine *core.Engine

	*model.App
	search *SearchPage

	lyricRenderer       *LyricRenderer
	songInfoRenderer    *SongInfoRenderer
	progressRenderer    *ProgressRenderer
	coverRenderer       *CoverRenderer
	spectrumRenderer    *SpectrumRenderer
	spectrogramRenderer *SpectrogramRenderer

	player *Player

	// frontendScope is the TUI frontend scope (P5): it mounts uiServicesPlugin
	// now and the 9 business plugins in P5-2. Owned by the shell — CloseHook
	// disposes it after the engine root scope cleanup.
	frontendScope *framework.Scope

	// wasmScope is the frontend-scope child (P6) that owns the app-wide WASM
	// manager (wasm.ManagerPlugin) and the dynamically loaded per-directory
	// wasmPlugin adapters (added by loadWasmPlugins via wasm.LoadIntoScope). Its
	// lifecycle rides the frontend scope's Dispose in CloseHook, so the shell
	// keeps no separate manager reference/cleanup.
	wasmScope *framework.Scope

	// ctx is the app-wide service registry context owned by the engine.
	ctx *framework.Context

	playbarHoveredElement PlaybarElement

	// Theme switch notification: update in-place when visible, recreate when expired.
	themeNotifID    model.NotificationID
	themeNotifTimer *time.Timer
}

func NewNetease(app *model.App) *Netease {
	connectMode = false
	n := new(Netease)

	// The user slot must exist before the engine captures &n.user.
	n.user = nil

	// The playback coordinator and its service assembly are UI-free: build the
	// core engine first (it owns the framework context/scope, the service
	// instances and the user slot), then wrap it with the TUI shell
	// (observer/loading/locator wiring happens in NewPlayer below).
	n.engine = core.NewEngine(core.EngineOptions{User: &n.user, DesktopLyrics: true})
	n.player = NewPlayer(n, n.engine.Player())
	n.ctx = n.engine.Ctx()

	// The framework context must exist before any newMenuServices call: the
	// accessor snapshots n.ctx at construction time (Player and renderers below
	// build their svc through it), so a nil ctx would make every service lookup
	// degrade to nil at play time. Services themselves are registered by the
	// engine (core) plus registerUIExtraServices below.
	n.lyricRenderer = NewLyricRenderer(newMenuServices(n), n.engine.LyricService(), configs.AppConfig.Main.Lyric.Show)
	n.songInfoRenderer = NewSongInfoRenderer(newMenuServices(n), n.player)
	n.progressRenderer = NewProgressRenderer(newMenuServices(n), n.player)
	n.coverRenderer = NewCoverRenderer(newMenuServices(n), n.player)
	n.spectrumRenderer = NewSpectrumRenderer(n.player)
	n.spectrogramRenderer = NewSpectrogramRenderer(n.player)

	// The engine registers its 8 core services into the app-wide context; the
	// TUI-only services (coverRenderer/menuRegistry/pageRegistry) are provided
	// by the frontend scope's uiServicesPlugin, and the 9 business plugins
	// (mounted by NewFrontendScope, filtered by [plugins] disabled) register
	// their menus/pages/main-menu items inside their own Start. The scope Start
	// is synchronous and replaces the former direct registerUIExtraServices +
	// init()-time plugin registration with identical effect.
	n.frontendScope = NewFrontendScope(n.engine, n)
	if err := n.frontendScope.Start(n.ctx); err != nil {
		slog.Error("framework frontend scope start failed", slogx.Error(err))
		return nil
	}

	// The search page is a shell-owned singleton: its wordsInput/result/
	// searchType state is shared with the SearchResultMenu flow (and operate.go
	// searchSong), so the shell keeps one instance built through the registry.
	// The login page is built per-navigation in ToLoginPage (no cross-component
	// state to preserve), so the shell holds no login field. Since P5 the
	// "search" page provider is registered by the search plugin's Start (inside
	// the frontend scope above), so this build must run after scope.Start.
	searchPage, err := BuildPage("search", SearchPageOpts{Netease: n})
	if err != nil {
		slog.Error("build search page failed", slogx.Error(err))
		return nil
	}
	n.search = searchPage.(*SearchPage) // BuildPage returns model.Page; concrete type asserted back
	n.App = app

	// WASM plugins must register here (inside NewNetease, before the main menu
	// is constructed in internal/commands) so their command providers and
	// main-menu items participate in the after-anchor chain from the start.
	// loadWasmPlugins mounts one wasmPlugin adapter per loaded plugin directory
	// into the wasm sub-scope (which the frontend scope Start already brought
	// up); the commands land in the frontend registry via the tui sink.
	n.loadWasmPlugins(context.Background())

	// Track-B commands become TUI CommandMenu entries after WASM plugins load
	// (and before the main menu is constructed in internal/commands), so their
	// providers and main-menu items join the after-anchor chain from the start.
	registerCommandMenus()

	// Startup completeness: a provider set missing any canonical key is a
	// programmer error; fail loudly instead of surfacing it at navigation time.
	// The assertions run after every contribution is registered (frontend scope
	// business-plugin Start → WASM → registerCommandMenus) and before the main
	// menu is constructed downstream (internal/ui/frontend.go NewMainMenu).
	// `expectedMenuKeys`/`expectedPageKeys` only lock the built-in set — keys
	// supplied by enabled plugins are intentionally absent.
	AssertMenuRegistryComplete(expectedMenuKeys...)
	AssertPageRegistryComplete(expectedPageKeys...)

	return n
}

// NewNeteaseRemote is the TUI-connect shell constructor (D-TC-1 方案 B): it
// shares the standalone assembly shape but builds no engine, registers no
// business services and runs no Startup sequence (B9). The player data surface
// is RemotePlayer (subscribed to the daemon via client); the embedded
// *core.Player is never constructed, so the ui.Player shadowing methods (see
// player.go) are the only reachable surface — anything missed would
// nil-dereference as the sentinel panic. Renderer degradation: lyric and
// spectrum renderers are skipped (B5/B7), the cover renderer is built with a
// PicUrl-stripped state (empty render, B6), and song info + progress renderers
// read the remote state.
//
// The frontend scope (S6-R1) mounts the 8 engine-independent business plugins
// (checkupdate/search/dj/album/artist/recommend/playlist/song; lastfm is
// excluded — its Deps needs engine services), so the local browsing tree
// (search / ranks / playlists / album / artist / DJ / recommend) is available
// in the remote shell per roadmap §8.4. The WASM sub-scope, loadWasmPlugins
// and registerCommandMenus stay skipped (B10: the command surface is disabled
// in connect mode).
func NewNeteaseRemote(app *model.App, client *headless.SubscribeClient) *Netease {
	connectMode = true
	n := new(Netease)
	n.user = nil

	// The App must be attached BEFORE the remote player is constructed:
	// NewRemotePlayer starts the event-consumer goroutine, which can render
	// the already-buffered snapshot frame before this constructor returns.
	// The consumer reads n.App through rerender, so a late assignment would
	// race the goroutine (and could nil-dereference it).
	n.App = app

	// No engine: the wrapper is built directly around the remote data surface.
	n.player = &Player{
		netease: n,
		svc:     newMenuServices(n),
		remote:  NewRemotePlayer(n, client),
	}
	// No framework context: every service lookup (track/lyric/user/login/...)
	// degrades to nil through the menuServices accessor, which is exactly the
	// degradation connect mode wants.
	n.ctx = nil

	// The connect frontend scope mounts the 8 engine-independent business
	// plugins (S6-R1). It must start BEFORE BuildPage("search") below: the
	// search plugin registers the "search" page provider inside its Start.
	// Started against a nil context — the mounted plugins ignore ctx.
	n.frontendScope = NewConnectFrontendScope(n)
	if err := n.frontendScope.Start(nil); err != nil {
		slog.Error("framework connect frontend scope start failed", slogx.Error(err))
		return nil
	}

	// Fallback guard: the "search" page provider normally comes from the
	// search plugin's Start above; only when [plugins] disabled contains
	// "search" does the shell re-provide it so the shared search-flow call
	// sites keep a non-nil singleton.
	registerConnectProviders()

	// Renderers: lyric/spectrum skipped (B5/B7), cover empty (B6). The svc is
	// resolved through newMenuServices(n) so every accessor is nil-degraded.
	n.songInfoRenderer = NewSongInfoRenderer(newMenuServices(n), n.player)
	n.progressRenderer = NewProgressRenderer(newMenuServices(n), n.player)
	n.coverRenderer = NewCoverRenderer(newMenuServices(n), connectCoverState{p: n.player})

	// Shell-owned search singleton (same as standalone): the provider was
	// registered by the search plugin's Start (or the fallback above); the
	// page renders locally and the search API call is local.
	searchPage, err := BuildPage("search", SearchPageOpts{Netease: n})
	if err != nil {
		slog.Error("build connect search page failed", slogx.Error(err))
		return nil
	}
	n.search = searchPage.(*SearchPage)

	// Startup completeness: the built-in provider set must be complete after
	// the connect scope Start (the plugin-supplied keys are intentionally not
	// asserted, mirroring standalone NewNetease).
	AssertMenuRegistryComplete(expectedMenuKeys...)
	AssertPageRegistryComplete(expectedPageKeys...)

	return n
}

// ConnectMode reports whether the shell runs in TUI-connect remote-shell mode
// (--frontend=tui --mode=connect): the player data surface is RemotePlayer and
// no engine/services/Startup exist (D-TC-1/B9).
func (n *Netease) ConnectMode() bool {
	return n.player != nil && n.player.remote != nil
}

func (n *Netease) Components() []model.Component {
	var components []model.Component
	if n.ConnectMode() {
		// TUI-connect renderer set (D-TC-3): lyric and spectrum renderers are
		// not built (B5/B7); song info + progress read the remote state; the
		// cover renderer is added but renders empty (PicUrl stripped, B6).
		components = append(components, n.songInfoRenderer, n.progressRenderer)
		if n.coverRenderer != nil && n.coverRenderer.IsEnabled() {
			components = append(components, n.coverRenderer)
		}
		return components
	}
	if n.spectrumRenderer.IsEnabled() {
		components = append(components, n.spectrumRenderer)
	}
	if n.spectrogramRenderer.IsEnabled() {
		components = append(components, n.spectrogramRenderer)
	}
	components = append(components, n.lyricRenderer)
	components = append(components, n.songInfoRenderer, n.progressRenderer)
	// CoverRenderer uses absolute positioning and returns 0 lines, so it must
	// be rendered last to overlay the normal layout.
	if n.coverRenderer.IsEnabled() {
		components = append(components, n.coverRenderer)
	}
	return components
}

func (n *Netease) SpectrumLines(main *model.Main) int {
	windowHeight := n.EffectiveWindowHeight()
	menuBottomRow := main.MenuBottomRow()
	if n.spectrumRenderer != nil && n.spectrumRenderer.IsEnabled() {
		return n.spectrumRenderer.LineCount(windowHeight, menuBottomRow)
	}
	if n.spectrogramRenderer != nil && n.spectrogramRenderer.IsEnabled() {
		return n.spectrogramRenderer.LineCount(windowHeight, menuBottomRow)
	}
	return 0
}

// EffectiveWindowHeight returns the available content height, excluding the
// status bar if present. Components should use this instead of WindowHeight()
// for bottom-anchored layout to avoid overlapping the status bar.
func (n *Netease) EffectiveWindowHeight() int {
	return n.MustMain().EffectiveWindowHeight(n.App)
}

// ToLoginPage 需要登录的处理
func (n *Netease) ToLoginPage(callback func() model.Page) (model.Page, tea.Cmd) {
	// connect mode runs the same flow (TC-6): the LoginPage renders only the
	// QR entry and the QR page sources the login from the daemon (D-TC-7), so
	// no connect-specific branch is needed here — the page itself guards
	// against the engine-dependent local login paths via n.ConnectMode().
	page := buildPageOrToast("login", LoginPageOpts{Netease: n})
	if page == nil {
		return nil, nil
	}
	page.(*LoginPage).AfterLogin = callback
	n.coverRenderer.ClearDisplayed()
	return page, tickLogin(time.Nanosecond)
}

// ToSearchPage 搜索
func (n *Netease) ToSearchPage(searchType SearchType) (model.Page, tea.Cmd) {
	// Search is fully local in connect mode (S6-R1): the search plugin is
	// mounted in the connect frontend scope, so the search page singleton and
	// the search_type/search_result/detail-jump menus all exist and the search
	// API call is local (no login required). Same flow as standalone.
	n.search.searchType = searchType
	n.coverRenderer.ClearDisplayed()
	return n.search, tickSearch(time.Nanosecond)
}

func (n *Netease) InitHook(_ *model.App) {
	// 注册 TUI 内 toast 回调（此时 App.Run 已启动，program 就绪）
	n.registerToastHook()

	if n.ConnectMode() {
		// B9: no engine Startup. The daemon connection + subscription were
		// established by RunConnect (DialSubscribe) and the remote player's
		// event consumer goroutine is already running; the shell was fully
		// assembled in NewNeteaseRemote. Nothing further to run.
		// The App is running now (InitHook fires after App.Run started the
		// program), so the consumer's render pokes can reach the shell.
		if n.player != nil && n.player.remote != nil {
			n.player.remote.markRunning()
		}
		return
	}

	// The engine owns the startup sequence (jar → user restore → playlist →
	// hooks → autoplay; see core.Startup). It runs in its own goroutine, and
	// the TUI-only changelog popup runs after the sequence completes — the
	// whole sequence used to run in one errorx.Go goroutine before the split.
	ctx := context.Background()
	errorx.Go(func() {
		_ = n.engine.Startup(ctx, n.player)
		n.maybeShowChangelog()
	})
}

// maybeShowChangelog shows the changelog popup on the first launch of a new
// version (or in debug mode). TUI-only; the engine startup is agnostic. Uses an
// AfterFunc delay so the startup page completes and the main page is entered.
func (n *Netease) maybeShowChangelog() {
	table := storage.NewTable()

	// changelog: 首次启动新版本或 debug 模式 → 弹更新日志
	// 使用 AfterFunc 延迟弹窗，确保 startup 页完成、主页面已进入
	{
		slog.Debug("changelog: entering check",
			"debug", configs.AppConfig.Main.Debug,
			"appVersion", types.AppVersion,
			"app", n.App != nil,
		)
		jsonStr, _ := table.GetByKVModel(storage.ChangelogSeen{})
		var seen storage.ChangelogSeen
		if len(jsonStr) > 0 {
			_ = json.Unmarshal(jsonStr, &seen)
		}
		shouldShow := configs.AppConfig.Main.Debug || seen.Version == "" || version.CompareVersion(types.AppVersion, seen.Version, false)
		slog.Debug("changelog: shouldShow",
			"shouldShow", shouldShow,
			"debug", configs.AppConfig.Main.Debug,
			"seenVersion", seen.Version,
		)
		if shouldShow {
			if !configs.AppConfig.Main.Debug {
				if err := table.SetByKVModel(storage.ChangelogSeen{}, storage.ChangelogSeen{Version: types.AppVersion}); err != nil {
					slog.Error("changelog: failed to persist seen version", slogx.Error(err))
				} else {
					slog.Debug("changelog: persisted seen version", "version", types.AppVersion)
				}
			}
			app := n.App
			slog.Debug("changelog: scheduling AfterFunc", "hasApp", app != nil)
			time.AfterFunc(max(configs.AppConfig.Startup.ToModel().LoadingDuration, time.Second)-750*time.Millisecond, func() {
				slog.Debug("changelog: AfterFunc triggered, showing popup")
				showChangelogPopup(app)
			})
		}
	}
}

func (n *Netease) CloseHook(_ *model.App) {
	// The frontend scope owns the TUI-only plugins (uiServicesPlugin, the 9
	// business plugins and the WASM sub-scope — the manager + plugin instances
	// are closed by the recursive child-scope Dispose). Dispose it before the
	// engine's root scope so frontend plugins are finalized ahead of the
	// service constructors (docs/plugin_ecosystem.md §五 step 8).
	if n.frontendScope != nil {
		_ = n.frontendScope.Dispose()
	}

	// The engine owns the player/lastfm/desktopLyrics/scope cleanup.
	if n.engine != nil {
		_ = n.engine.Close()
	}

	if n.coverRenderer != nil {
		n.coverRenderer.Close()
	}

	CloseGohookLogger()
}

// Ctx returns the app-wide framework context holding the service registry.
func (n *Netease) Ctx() *framework.Context {
	return n.ctx
}

func (n *Netease) Player() *Player {
	return n.player
}

// DesktopLyrics returns the desktop lyrics controller (owned by the engine).
func (n *Netease) DesktopLyrics() desktop_lyrics.Controller {
	if n.engine == nil {
		return nil
	}
	return n.engine.DesktopLyrics()
}

// GetCoverWidth returns the cover image width in columns, or 0 if cover is disabled.
func (n *Netease) GetCoverWidth() int {
	if n.coverRenderer == nil {
		return 0
	}
	return n.coverRenderer.GetCoverWidth()
}

// GetCoverEndColumn returns the column where the cover ends, or 0 if cover is disabled.
func (n *Netease) GetCoverEndColumn() int {
	if n.coverRenderer == nil {
		return 0
	}
	return n.coverRenderer.GetCoverEndColumn()
}

// GetLyricPosition returns the current lyric display position.
// Returns (startRow, lineCount). If lyrics are not visible, returns (0, 0).
func (n *Netease) GetLyricPosition() (startRow int, lineCount int) {
	if n.lyricRenderer == nil {
		return 0, 0
	}
	return n.lyricRenderer.GetLyricPosition()
}

// showChangelogPopup reads the embedded CHANGELOG.md and displays it as a markdown popup.
func showChangelogPopup(app *model.App) {
	data, err := filex.ReadFileFromEmbed("embed/changelog.md")
	if err != nil {
		slog.Error("changelog load failure", slogx.Error(err))
		return
	}

	popupWidth := min(max(app.WindowWidth()*70/100, 40), 120)

	maxHeight := max(app.WindowHeight()*80/100, 10)

	popup, err := model.NewMarkdownPopup(model.MarkdownPopupSpec{
		Title:           types.AppVersion + " 更新日志",
		MarkdownContent: string(data),
		MaxWidth:        popupWidth,
		MaxHeight:       maxHeight,
		Anchor:          model.AnchorCenter,
		DisableResize:   false,
		CloseKeys:       []string{"esc", "q"},
	})
	if err != nil {
		slog.Error("changelog popup creation failure", slogx.Error(err))
		return
	}
	app.ShowPopup(popup)
	// Trigger a rerender so the popup appears immediately without waiting for a keypress
	app.Rerender(true)
}

// Update intercepts system background-change messages to rebuild the StyleSet
// from go-musicfox's ThemeRegistry with the correct dark/light variant.
// Foxful-cli's built-in onBackgroundChanged handles SetDarkBackground and popup
// cache invalidation; we override the StyleSet afterwards to use our theme files.
func (n *Netease) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.BackgroundColorMsg, uv.LightColorSchemeEvent, uv.DarkColorSchemeEvent:
		registry := configs.CurrentThemeRegistry()
		syncActiveThemePair(n.App, registry)
		// Let foxful-cli process the message first (SetDarkBackground, popup cache, etc.)
		_, cmd := n.App.Update(msg)
		// Rebuild StyleSet from our ThemeRegistry with correct dark/light variant
		isDark := style.HasDarkBackground()
		ss := registry.CurrentStyleSet(isDark)
		if ss != nil {
			style.SetStyleSet(*ss)
			n.SetStyleSet(*ss)
		}
		n.notifyThemeSwitch(n.App, "外观模式已切换", configs.CurrentThemeRegistry().CurrentName(isDark))
		return n, tea.Sequence(cmd, n.RerenderCmd(true))
	default:
		_, cmd := n.App.Update(msg)
		return n, cmd
	}
}

func syncActiveThemePair(app *model.App, registry *configs.ThemeRegistry) {
	dark, light, ok := registry.CurrentThemePair()
	if ok {
		app.SetThemePair(dark, light)
	}
}

// notifyThemeSwitch shows or updates a theme switch notification.
// If a previous notification is still visible, updates it in place.
// If it has expired, creates a new one.
func (n *Netease) notifyThemeSwitch(app *model.App, title, name string) {
	const timeout = 4 * time.Second
	spec := model.NotificationSpec{
		Level:   model.NotificationInfo,
		Title:   title,
		Message: name,
		Timeout: timeout,
	}
	if n.themeNotifID == 0 {
		n.themeNotifID = app.Notify(spec)
	} else {
		app.UpdateNotification(n.themeNotifID, spec)
	}
	if n.themeNotifTimer != nil {
		n.themeNotifTimer.Stop()
	}
	n.themeNotifTimer = time.AfterFunc(timeout, func() {
		n.themeNotifID = 0
	})
}
