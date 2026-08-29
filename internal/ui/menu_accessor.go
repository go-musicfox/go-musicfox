package ui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/track"
)

// menuServices is the type-safe accessor that replaces direct *Netease field
// reach-throughs in menus (baseMenu.netease → accessor, Phase 3.1.4). Each
// accessor resolves its value from the framework service registry; when a
// service is missing it returns the zero value (nil) with a warning log.
// Services are registered at startup before any menu renders, so the nil case
// should never happen in practice.
type menuServices struct {
	ctx *framework.Context
	n   *Netease
}

// MenuServices is the exported interface mirroring the public surface of the
// menuServices accessor (Phase 3.9 plugin boundary, P1-4 ticket). External
// packages (plugins) reference it in their signatures, opts fields and page
// constructors — e.g. `NewLastfmAuthPage(svc ui.MenuServices)`. Internal code
// keeps using menuServices unchanged; *menuServices implements the interface
// implicitly, so the type name change from alias to interface keeps all plugin
// call sites compiling untouched.
//
// WARNING: unlike the former type alias (type MenuServices = *menuServices),
// this interface carries typed-nil semantics — a nil *menuServices stored in a
// MenuServices value is a non-nil interface. Do NOT compare MenuServices
// values against nil or perform `.(*menuServices)` type assertions; callers
// that need nil checks must hold the concrete *menuServices. Exception:
// BaseMenu.Services() deliberately returns a true nil interface for a zero
// base, so its return value may be nil-compared (same as the old alias).
type MenuServices interface {
	Player() *Player
	User() *structs.User
	TrackManager() *track.Manager
	LyricService() *lyric.Service
	DesktopLyrics() desktop_lyrics.Controller
	CoverRenderer() *CoverRenderer
	ShareSvc() *composer.ShareService
	Lastfm() *lastfm.Client
	Ctx() *framework.Context
	Netease() *Netease
	ToLoginPage(callback func() model.Page) (model.Page, tea.Cmd)
	ToSearchPage(searchType SearchType) (model.Page, tea.Cmd)
	App() *model.App
	Main() *model.Main
	MustMain() *model.Main
	Rerender(force bool)
	Search() *SearchPage
	SaveActiveTheme(name string)
	NotifyThemeSwitch(app *model.App, title, name string)
	PlaybarHoveredElement() PlaybarElement
	SetPlaybarHoveredElement(e PlaybarElement)
	EffectiveWindowHeight() int
	SpectrumLines(main *model.Main) int
	GetCoverWidth() int
	GetCoverEndColumn() int
	GetLyricPosition() (startRow int, lineCount int)
}

// Compile-time assertion: *menuServices satisfies the exported MenuServices
// interface.
var _ MenuServices = (*menuServices)(nil)

// NewMenuServices builds an accessor rooted at the framework context, without
// attaching to a Netease shell (shell-dependent forwards — MustMain/App/
// ToLoginPage etc. — degrade to nil/zero). Exported for plugin page flows and
// tests that resolve services from a context.
func NewMenuServices(ctx *framework.Context) MenuServices {
	return &menuServices{ctx: ctx}
}

// newMenuServices builds the accessor from the framework context owned by the
// Netease shell.
func newMenuServices(n *Netease) *menuServices {
	if n == nil {
		return &menuServices{}
	}
	return &menuServices{ctx: n.ctx, n: n}
}

// missing logs a warning for an unresolvable service.
func (s *menuServices) missing(name string) {
	slog.Warn("menuServices: service not resolvable, returning nil", "service", name)
}

// Player resolves the TUI player wrapper. The engine registers the core
// player under ServicePlayer (headless consumers resolve *core.Player); the
// TUI wrapper is a shell-owned instance carrying the render ticker and
// playing-menu state, so the accessor prefers the shell's wrapper and falls
// back to a context-registered wrapper.
func (s *menuServices) Player() *Player {
	// TUI-connect remote shell: the shell-owned wrapper carries the remote
	// data surface; the framework player service never exists (no engine, B9).
	if s.n != nil && s.n.player != nil && s.n.player.remote != nil {
		return s.n.player
	}
	if svc, ok := framework.ServiceOf[*Player](s.ctx, ServicePlayer); ok {
		return svc
	}
	if s.n != nil && s.n.player != nil {
		return s.n.player
	}
	s.missing(ServicePlayer)
	return nil
}

// User resolves the current user. In connect mode it returns the daemon
// snapshot's user (nickname only, B8) — the engine's user service never
// exists and TUI-side login is disabled.
func (s *menuServices) User() *structs.User {
	if s.n != nil && s.n.player != nil && s.n.player.remote != nil {
		return s.n.player.remote.User()
	}
	if svc, ok := framework.ServiceOf[*core.UserService](s.ctx, ServiceUserService); ok && svc.User != nil {
		return *svc.User
	}
	s.missing(ServiceUserService)
	return nil
}

// TrackManager resolves the track manager service.
func (s *menuServices) TrackManager() *track.Manager {
	if svc, ok := framework.ServiceOf[*track.Manager](s.ctx, ServiceTrackManager); ok {
		return svc
	}
	s.missing(ServiceTrackManager)
	return nil
}

// LyricService resolves the lyric service.
func (s *menuServices) LyricService() *lyric.Service {
	if svc, ok := framework.ServiceOf[*lyric.Service](s.ctx, ServiceLyricService); ok {
		return svc
	}
	s.missing(ServiceLyricService)
	return nil
}

// DesktopLyrics resolves the desktop lyrics controller.
func (s *menuServices) DesktopLyrics() desktop_lyrics.Controller {
	if svc, ok := framework.ServiceOf[desktop_lyrics.Controller](s.ctx, ServiceDesktopLyrics); ok {
		return svc
	}
	s.missing(ServiceDesktopLyrics)
	return nil
}

// CoverRenderer resolves the cover renderer.
func (s *menuServices) CoverRenderer() *CoverRenderer {
	if svc, ok := framework.ServiceOf[*CoverRenderer](s.ctx, ServiceCoverRenderer); ok {
		return svc
	}
	s.missing(ServiceCoverRenderer)
	return nil
}

// ShareSvc resolves the share service.
func (s *menuServices) ShareSvc() *composer.ShareService {
	if svc, ok := framework.ServiceOf[*composer.ShareService](s.ctx, ServiceShareSvc); ok {
		return svc
	}
	s.missing(ServiceShareSvc)
	return nil
}

// Lastfm resolves the Last.fm client.
func (s *menuServices) Lastfm() *lastfm.Client {
	if svc, ok := framework.ServiceOf[*lastfm.Client](s.ctx, ServiceLastfm); ok {
		return svc
	}
	s.missing(ServiceLastfm)
	return nil
}

// Ctx returns the app-wide framework context backing this accessor.
func (s *menuServices) Ctx() *framework.Context {
	return s.ctx
}

// Netease returns the Netease shell. Escape hatch for the migration window
// (3.1.5 removes the remaining *Netease reach-throughs).
func (s *menuServices) Netease() *Netease {
	return s.n
}

// ToLoginPage forwards to the thin-shell login navigation (builds a fresh
// login page through the "login" provider and wires the AfterLogin callback).
func (s *menuServices) ToLoginPage(callback func() model.Page) (model.Page, tea.Cmd) {
	if s.n == nil {
		return nil, nil
	}
	return s.n.ToLoginPage(callback)
}

// ToSearchPage forwards to the thin-shell search navigation (returns the
// shell-owned search singleton with searchType set).
func (s *menuServices) ToSearchPage(searchType SearchType) (model.Page, tea.Cmd) {
	if s.n == nil {
		return nil, nil
	}
	return s.n.ToSearchPage(searchType)
}

// --- Thin-shell methods (Phase 3.3.3): shell capabilities routed through the
// accessor so deep-coupled files no longer reach into the *Netease shell
// directly. Each method forwards to the shell and is nil-safe (missing shell
// degrades to the zero value with no panic).

// App returns the foxful app shell (embedded *model.App).
func (s *menuServices) App() *model.App {
	if s.n == nil {
		return nil
	}
	return s.n.App
}

// Main returns the foxful main page (nil when the app has not started).
func (s *menuServices) Main() *model.Main {
	if s.n == nil || s.n.App == nil {
		return nil
	}
	return s.n.Main()
}

// MustMain returns the foxful main page (nil when the app has not started;
// callers use it where the embedded model.App method was reached before).
func (s *menuServices) MustMain() *model.Main {
	if s.n == nil || s.n.App == nil {
		return nil
	}
	return s.n.MustMain()
}

// Rerender triggers a foxful re-render on the app shell.
func (s *menuServices) Rerender(force bool) {
	if s.n == nil || s.n.App == nil {
		return
	}
	s.n.Rerender(force)
}

// Search returns the shell-owned search page singleton (shared state with the
// SearchResultMenu flow; nil when the shell is missing).
func (s *menuServices) Search() *SearchPage {
	if s.n == nil {
		return nil
	}
	return s.n.search
}

// SaveActiveTheme persists the active theme name (theme_persistence.go).
func (s *menuServices) SaveActiveTheme(name string) {
	if s.n != nil {
		s.n.saveActiveTheme(name)
	}
}

// NotifyThemeSwitch shows or updates the theme-switch notification
// (netease.go notifyThemeSwitch).
func (s *menuServices) NotifyThemeSwitch(app *model.App, title, name string) {
	if s.n != nil {
		s.n.notifyThemeSwitch(app, title, name)
	}
}

// PlaybarHoveredElement returns the hovered playbar element tracked by the shell.
func (s *menuServices) PlaybarHoveredElement() PlaybarElement {
	if s.n == nil {
		return PlaybarElementNone
	}
	return s.n.playbarHoveredElement
}

// SetPlaybarHoveredElement records the hovered playbar element on the shell.
func (s *menuServices) SetPlaybarHoveredElement(e PlaybarElement) {
	if s.n == nil {
		return
	}
	s.n.playbarHoveredElement = e
}

// EffectiveWindowHeight returns the available content height, excluding the
// status bar if present (forwarded from the shell; Phase 3.8).
func (s *menuServices) EffectiveWindowHeight() int {
	if s.n == nil {
		return 0
	}
	return s.n.EffectiveWindowHeight()
}

// SpectrumLines returns the line count consumed by the enabled spectrum or
// spectrogram renderer for the given main page (forwarded from the shell).
func (s *menuServices) SpectrumLines(main *model.Main) int {
	if s.n == nil {
		return 0
	}
	return s.n.SpectrumLines(main)
}

// GetCoverWidth returns the cover image width in columns, or 0 if the cover
// renderer is disabled (forwarded from the shell).
func (s *menuServices) GetCoverWidth() int {
	if s.n == nil {
		return 0
	}
	return s.n.GetCoverWidth()
}

// GetCoverEndColumn returns the column where the cover ends (start column +
// width), or 0 if the cover renderer is disabled (forwarded from the shell).
func (s *menuServices) GetCoverEndColumn() int {
	if s.n == nil {
		return 0
	}
	return s.n.GetCoverEndColumn()
}

// GetLyricPosition returns the current lyric display position
// (startRow, lineCount); (0, 0) when lyrics are not visible (forwarded from
// the shell).
func (s *menuServices) GetLyricPosition() (startRow int, lineCount int) {
	if s.n == nil {
		return 0, 0
	}
	return s.n.GetLyricPosition()
}
