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
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/filex"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/version"
)

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

	// wasmManager owns the loaded WASM plugin instances (created lazily in
	// NewNetease; nil when the runtime failed to initialize).
	wasmManager *wasm.Manager

	// ctx is the app-wide service registry context owned by the engine.
	ctx *framework.Context

	playbarHoveredElement PlaybarElement

	// Theme switch notification: update in-place when visible, recreate when expired.
	themeNotifID    model.NotificationID
	themeNotifTimer *time.Timer
}

func NewNetease(app *model.App) *Netease {
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

	// The search page is a shell-owned singleton: its wordsInput/result/
	// searchType state is shared with the SearchResultMenu flow (and operate.go
	// searchSong), so the shell keeps one instance built through the registry.
	// The login page is built per-navigation in ToLoginPage (no cross-component
	// state to preserve), so the shell holds no login field.
	searchPage, err := BuildPage("search", SearchPageOpts{Netease: n})
	if err != nil {
		return nil
	}
	n.search = searchPage.(*SearchPage) // BuildPage returns model.Page; concrete type asserted back
	n.App = app

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

	// WASM plugins must register here (inside NewNetease, before the main menu
	// is constructed in internal/commands) so their command providers and
	// main-menu items participate in the after-anchor chain from the start.
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

func (n *Netease) Components() []model.Component {
	var components []model.Component
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
	n.search.searchType = searchType
	n.coverRenderer.ClearDisplayed()
	return n.search, tickSearch(time.Nanosecond)
}

func (n *Netease) InitHook(_ *model.App) {
	// 注册 TUI 内 toast 回调（此时 App.Run 已启动，program 就绪）
	n.registerToastHook()

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
	// The frontend scope owns the TUI-only plugins (uiServicesPlugin now, the
	// 9 business plugins in P5-2). Dispose it before the engine's root scope so
	// frontend plugins are finalized ahead of the service constructors
	// (docs/plugin_ecosystem.md §五 step 8).
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

	if n.wasmManager != nil {
		n.wasmManager.Close(context.Background())
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
