package ui

import (
	"errors"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
)

// Canonical service names. The engine-owned core services (player / lyric /
// track manager / desktop lyrics / user / login / share / lastfm) are defined
// in internal/core; ui aliases them so existing call sites and tests keep
// compiling unchanged against the core values. The TUI-only services
// (coverRenderer / menuRegistry / pageRegistry) stay defined here and are
// registered into the engine context via registerUIExtraServices.
const (
	ServicePlayer        = core.ServicePlayer
	ServiceLyricService  = core.ServiceLyricService
	ServiceTrackManager  = core.ServiceTrackManager
	ServiceDesktopLyrics = core.ServiceDesktopLyrics
	ServiceUserService   = core.ServiceUserService
	ServiceLoginService  = core.ServiceLoginService
	ServiceShareSvc      = core.ServiceShareSvc
	ServiceLastfm        = core.ServiceLastfm

	ServiceCoverRenderer = "coverRenderer"
	ServiceMenuRegistry  = "menuRegistry"
	ServicePageRegistry  = "pageRegistry"
)

// registerUIExtraServices registers the TUI-only services into the app-wide
// framework context (the engine registers its 8 core services into the same
// context at boot). It is a plain callable so tests can run it against a fresh
// context without a full app boot.
func registerUIExtraServices(ctx *framework.Context, n *Netease) error {
	if ctx == nil || n == nil {
		return errors.New("registerUIExtraServices: nil ctx or Netease")
	}
	core.ProvideIfAbsent(ctx, ServiceCoverRenderer, n.coverRenderer)
	// Provider registries (Phase 3.2.1): the generic RegisterMenu/BuildMenu are
	// package-level functions (Go forbids generic methods); the service handles
	// make the registries resolvable as framework services for completeness
	// assertions, tests and future plugin boundaries.
	core.ProvideIfAbsent(ctx, ServiceMenuRegistry, MenuRegistry{})
	core.ProvideIfAbsent(ctx, ServicePageRegistry, PageRegistry{})
	return nil
}
