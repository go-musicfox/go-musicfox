package ui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/composer"
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

// Player resolves the player service.
func (s *menuServices) Player() *Player {
	if svc, ok := framework.ServiceOf[*Player](s.ctx, ServicePlayer); ok {
		return svc
	}
	s.missing(ServicePlayer)
	return nil
}

// User resolves the current user from the userService slot (nil until login).
func (s *menuServices) User() *structs.User {
	if svc, ok := framework.ServiceOf[*UserService](s.ctx, ServiceUserService); ok && svc.User != nil {
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
