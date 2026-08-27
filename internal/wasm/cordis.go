package wasm

import (
	"context"
	"fmt"

	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// ServiceWasmManager is the framework service name under which the app-wide
// WASM Manager (the wazero runtime owner) is provided. The constant lives here
// rather than in internal/core because the manager is owned by the frontend wasm
// sub-scope (per frontend), not by the engine root scope — core's
// registeredServiceNames exact-set assertion therefore stays untouched, and wasm
// keeps its zero-dependency-on-core direction.
const ServiceWasmManager = "wasmManager"

// ManagerPlugin is the frontend wasm sub-scope plugin that creates and owns the
// app-wide WASM Manager (docs/plugin_ecosystem.md §3.3): Start = NewManager +
// provideIfAbsent(ServiceWasmManager); Dispose = Manager.Close. Exactly one
// instance per frontend (TUI / WebUI). Stop is a no-op so the manager — and the
// loaded plugin instances — survive a scope restart (Stop → re-Start) and are
// only released by Dispose.
type ManagerPlugin struct {
	mgr *Manager
}

// Start creates the manager when not yet created (LoadIntoScope may hand one
// over before the scope starts) and registers it into the shared app context so
// wasmPlugin Deps can resolve it.
func (p *ManagerPlugin) Start(ctx *framework.Context) error {
	if p.mgr == nil {
		mgr, err := NewManager()
		if err != nil {
			return fmt.Errorf("wasm: new manager: %w", err)
		}
		p.mgr = mgr
	}
	provideIfAbsent(ctx, ServiceWasmManager, p.mgr)
	return nil
}

// Stop is a no-op: the manager stays alive across scope restarts; only Dispose
// releases the wazero runtime.
func (p *ManagerPlugin) Stop() error { return nil }

// Dispose closes the manager (plugin instances in reverse load order, then the
// runtime). Idempotent.
func (p *ManagerPlugin) Dispose() error {
	if p.mgr == nil {
		return nil
	}
	p.mgr.Close(context.Background())
	p.mgr = nil
	return nil
}

// wasmPlugin adapts one loaded WASM plugin directory into a dynamic frontend
// wasm sub-scope plugin. Its lifecycle is the P6 core: Start maps the plugin's
// manifest menus into track-B commands and registers them through the sink
// (Replace semantics — a reloaded plugin replaces its previous command
// definitions instead of panicking on a duplicate key); Stop unregisters them;
// Dispose closes the plugin instance.
type wasmPlugin struct {
	framework.PluginBase
	mgr  *Manager     // Deps: ServiceWasmManager
	p    *Plugin      // LoadDir 产物
	sink RegistrySink // TUI: tuiWasmSink / WebUI: webuiWasmSink
	cmds []frontend.Command
}

// Deps resolves the app-wide manager from the shared context (the dependency
// that links the adapter to the scope's ManagerPlugin).
func (w *wasmPlugin) Deps(ctx *framework.Context) error {
	mgr, ok := framework.ServiceOf[*Manager](ctx, ServiceWasmManager)
	if !ok {
		return fmt.Errorf("wasmPlugin: service %q not resolved", ServiceWasmManager)
	}
	w.mgr = mgr
	return nil
}

// Start maps the plugin menus into commands and registers them through the
// sink. The command set is kept so Stop can unregister exactly what Start
// registered.
func (w *wasmPlugin) Start(ctx *framework.Context) error {
	w.cmds = CommandsOf(w.p)
	return w.sink.RegisterCommands(w.p, w.cmds)
}

// Stop unregisters every command Start registered (frontend.UnregisterCommand
// is a no-op for keys this adapter did not register).
func (w *wasmPlugin) Stop() error {
	for _, c := range w.cmds {
		frontend.UnregisterCommand(c.Key)
	}
	return nil
}

// Dispose closes the plugin instance. Idempotent — the manager's Close also
// releases instances in reverse load order, so the double close is safe.
func (w *wasmPlugin) Dispose() error {
	if w.p != nil {
		w.p.Close(context.Background())
	}
	return nil
}

// LoadIntoScope scans dir for plugin directories and mounts each successfully
// loaded plugin as a wasmPlugin adapter into scope. The scope must be (or
// contain) a wasm sub-scope carrying a ManagerPlugin — one is registered when
// absent. When the scope has already been started the adapters start
// immediately (Scope.AddAndStart); otherwise they are added and start with the
// scope's Start, whose framework context is fctx (must be non-nil).
//
// The returned manager is the one owned by the scope's ManagerPlugin (the same
// manager that performed the LoadDir). A single failing plugin or registration
// does not stop the others; all errors are collected and returned. Intended for
// one load per scope generation — a hot reload (P8) mounts a fresh scope.
func LoadIntoScope(ctx context.Context, fctx *framework.Context, scope *framework.Scope, dir string, sink RegistrySink) (*Manager, []error) {
	if fctx == nil {
		return nil, []error{fmt.Errorf("wasm: LoadIntoScope: nil framework context")}
	}
	mp, err := ensureManagerPlugin(scope)
	if err != nil {
		return nil, []error{fmt.Errorf("wasm: ensure manager plugin: %w", err)}
	}
	if mp.mgr == nil {
		mgr, err := NewManager()
		if err != nil {
			return nil, []error{fmt.Errorf("wasm: new manager: %w", err)}
		}
		mp.mgr = mgr
	}
	// Make the manager resolvable from the shared context whether or not the
	// scope has started yet (the ManagerPlugin Start provideIfAbsent is the
	// idempotent twin of this).
	provideIfAbsent(fctx, ServiceWasmManager, mp.mgr)

	mgr := mp.mgr
	errs := mgr.LoadDir(ctx, dir)
	for _, p := range mgr.Plugins() {
		wp := &wasmPlugin{p: p, sink: sink}
		if err := scope.AddAndStart(fctx, wp); err != nil {
			errs = append(errs, fmt.Errorf("wasm: start plugin %q: %w", p.ID, err))
		}
	}
	return mgr, errs
}

// ensureManagerPlugin returns the scope's ManagerPlugin, registering a fresh
// one when absent so LoadIntoScope works on a bare scope too (the scope must
// not be disposed).
func ensureManagerPlugin(scope *framework.Scope) (*ManagerPlugin, error) {
	for _, p := range scope.Plugins() {
		if mp, ok := p.(*ManagerPlugin); ok {
			return mp, nil
		}
	}
	mp := &ManagerPlugin{}
	if err := scope.Add(mp); err != nil {
		return nil, err
	}
	return mp, nil
}

// provideIfAbsent registers svc under name unless it is already present, so the
// scope-provided service and the ManagerPlugin Start can coexist without a
// duplicate-Provide panic.
func provideIfAbsent(ctx *framework.Context, name string, svc any) {
	if ctx.Service(name) == nil {
		ctx.Provide(name, svc)
	}
}
