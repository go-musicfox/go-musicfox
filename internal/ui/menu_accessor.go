package ui

import (
	"log/slog"

	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/internal/framework"
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

// Ctx returns the app-wide framework context backing this accessor.
func (s *menuServices) Ctx() *framework.Context {
	return s.ctx
}

// Netease returns the Netease shell. Escape hatch for the migration window
// (3.1.5 removes the remaining *Netease reach-throughs).
func (s *menuServices) Netease() *Netease {
	return s.n
}
