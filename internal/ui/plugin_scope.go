package ui

import (
	"log/slog"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// TUI frontend scope (P5): the shell builds one framework.Scope per Netease
// that mounts the frontend-only plugins. The scope shares the engine's app-wide
// framework context (the caller passes it to Scope.Start) and is owned by the
// Netease shell — CloseHook disposes it after the engine's root scope cleanup
// ordering is done (see netease.go CloseHook).
//
// P5-1 mounts only uiServicesPlugin; the 9 business plugins migrate in P5-2
// (their Add list below is the only thing that changes). Until then the
// business plugins keep registering through init() + ui.WithPlugin, so the
// frontend scope Start has the same observable effect as the former direct
// registerUIExtraServices call.

// businessPluginIDs is the ordered id list of the 9 compile-time business
// plugins, mirroring the blank-import declaration order in
// internal/plugins/plugins.go (registration order = Start order = menu
// registration order).
//
// TODO(P5-2): when each plugin migrates from init() + ui.WithPlugin to
// framework.RegisterPlugin + Start-time registration, NewFrontendScope mounts
// the enabled subset right after uiServicesPlugin:
//
//	constructors := framework.PluginConstructors()
//	for _, id := range businessPluginIDs {
//		ctor, ok := constructors[id]
//		if !ok {
//			continue // not yet migrated: keep its init() registration path
//		}
//		if err := scope.AddWithEnabled(ctor(), configs.IsPluginEnabled(id)); err != nil {
//			slog.Error("framework business plugin registration failed", "plugin", id, slogx.Error(err))
//		}
//	}
//
// Disabled semantics at that point move from consumption-point filtering to
// registration-time exclusion: a disabled plugin never starts, so its
// registrations (menus/pages/main-menu items/commands) never happen.
var businessPluginIDs = []string{
	"checkupdate",
	"lastfm",
	"dj",
	"album",
	"artist",
	"recommend",
	"playlist",
	"search",
	"song",
}

// NewFrontendScope builds the TUI frontend scope. e anchors the scope to the
// engine whose context it starts against; n is the thin-shell reference the
// uiServicesPlugin needs to reach the TUI-only service instances (cover
// renderer etc.). It only assembles the scope — the caller (NewNetease) starts
// it so a Start failure can be surfaced with the surrounding boot logic.
func NewFrontendScope(e *core.Engine, n *Netease) *framework.Scope {
	if e == nil {
		panic("NewFrontendScope: nil engine")
	}
	if n == nil {
		panic("NewFrontendScope: nil Netease")
	}

	scope := framework.NewScope()
	if err := scope.Add(&uiServicesPlugin{n: n}); err != nil {
		// Unreachable for a fresh scope; log and return the (empty) scope so
		// the caller keeps a non-nil scope instead of nil-dereferencing.
		slog.Error("framework frontend uiServicesPlugin registration failed", slogx.Error(err))
		return scope
	}

	// TODO(P5-2): mount the 9 business plugins here (businessPluginIDs order),
	// filtered by configs.IsPluginEnabled — see the list comment above.

	return scope
}

// uiServicesPlugin registers the TUI-only services (coverRenderer /
// menuRegistry / pageRegistry) into the app-wide framework context when the
// frontend scope starts, replacing the former direct registerUIExtraServices
// call in NewNetease with identical effect. It embeds NoopPlugin: it holds no
// per-scope state and needs no cleanup.
type uiServicesPlugin struct {
	framework.NoopPlugin
	n *Netease
}

func (p *uiServicesPlugin) Start(ctx *framework.Context) error {
	return registerUIExtraServices(ctx, p.n)
}
