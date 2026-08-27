package ui

import (
	"log/slog"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// TUI frontend scope (P5): the shell builds one framework.Scope per Netease
// that mounts the frontend-only plugins. The scope shares the engine's app-wide
// framework context (the caller passes it to Scope.Start) and is owned by the
// Netease shell — CloseHook disposes it after the engine's root scope cleanup
// ordering is done (see netease.go CloseHook).
//
// P5-1 mounted only uiServicesPlugin; P5-2 mounts the 9 business plugins right
// after it. Each business plugin's Start registers its contributions (menus /
// pages / main-menu items / startup hooks) into the package-level registries
// inside a ui.WithPlugin scope, so the frontend scope Start has exactly the
// same observable effect as the former init()-time registration.

// businessPluginIDs is the ordered id list of the 9 compile-time business
// plugins, mirroring the blank-import declaration order in
// internal/plugins/plugins.go (registration order = Start order = menu
// registration order).
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

	// Mount the 9 business plugins (businessPluginIDs order), filtered by
	// configs.IsPluginEnabled. A constructor missing from the framework
	// registry is skipped with a warning: in a binary where the plugin package
	// is not linked (e.g. the ui test binary) there is nothing to mount, and
	// during a partial migration the unmigrated plugin keeps its init()-time
	// registration path. AddWithEnabled gives the P5 "disabled = nonexistent"
	// semantics: a disabled plugin never starts, so its registrations (menus /
	// pages / main-menu items / startup hooks) never happen. Each mounted
	// plugin is wrapped in identifiedPlugin so PluginInfos() (P8, scope-driven
	// collection) can attribute it; the underlying lifecycle is forwarded.
	constructors := framework.PluginConstructors()
	for _, id := range businessPluginIDs {
		ctor, ok := constructors[id]
		if !ok {
			slog.Warn("framework business plugin constructor not registered, skipping mount", "plugin", id)
			continue
		}
		if err := scope.AddWithEnabled(&identifiedPlugin{Plugin: ctor(), id: id}, configs.IsPluginEnabled(id)); err != nil {
			slog.Error("framework business plugin registration failed", "plugin", id, slogx.Error(err))
		}
	}

	// WASM sub-scope (P6): a frontend-scope child that owns the app-wide WASM
	// manager (wasm.ManagerPlugin) and, from loadWasmPlugins onwards, one
	// wasmPlugin adapter per loaded plugin directory. Being a child of the
	// frontend scope makes its cleanup (Stop unregisters commands, Dispose
	// closes plugin instances + manager) ride the frontend scope's Dispose in
	// CloseHook. The ManagerPlugin must be registered before any dynamic
	// wasmPlugin adapter so its Start (which provides ServiceWasmManager into
	// the app context) runs first; loadWasmPlugins AddAndStarts the adapters on
	// this already-started child scope via wasm.LoadIntoScope.
	wasmScope := scope.NewScope()
	if err := wasmScope.Add(&wasm.ManagerPlugin{}); err != nil {
		slog.Error("framework frontend wasmManagerPlugin registration failed", slogx.Error(err))
	}
	n.wasmScope = wasmScope

	// Record the scope so PluginInfos() can collect the active plugin set from
	// it (P8): the frontend scope is the single source of truth for which
	// plugins are mounted (the wasm sub-scope contributes the WASM adapters).
	activeFrontendScope = scope

	return scope
}

// identifiedPlugin decorates a mounted business plugin with its plugin id so
// PluginInfos() (scope-driven collection) can attribute the scope entry to the
// plugin's WithPlugin record. It forwards the whole lifecycle through the
// embedded framework.Plugin, and carries an embedded PluginBase whose Enabled
// flag is written by AddWithEnabled — PluginInfos reads it back to exclude
// plugins that never started. PluginName stays empty — the display name comes
// from the WithPlugin record the plugin's Start creates (all 9 business
// plugins record inside Start).
type identifiedPlugin struct {
	framework.Plugin
	framework.PluginBase
	id string
}

// PluginID reports the plugin id the scope mounted it under.
func (p *identifiedPlugin) PluginID() string { return p.id }

// PluginName is not carried by the decorator (the WithPlugin record supplies
// it); returns empty so PluginInfos() falls back to the record's name.
func (p *identifiedPlugin) PluginName() string { return "" }

// IsEnabled reports the enabled flag AddWithEnabled wrote onto the embedded
// PluginBase (PluginInfos exclusion filter for never-started plugins).
func (p *identifiedPlugin) IsEnabled() bool { return p.Enabled }

// Deps forwards to the underlying plugin when it implements PluginWithDeps
// (the decorator embeds the base Plugin interface, which has no Deps slot).
func (p *identifiedPlugin) Deps(ctx *framework.Context) error {
	if withDeps, ok := p.Plugin.(framework.PluginWithDeps); ok {
		return withDeps.Deps(ctx)
	}
	return nil
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
