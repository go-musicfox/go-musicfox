package framework

import "log/slog"

// StartupHook is one registered startup task. PluginID names the owning plugin
// (empty for shell-level hooks, which are never filtered).
type StartupHook struct {
	PluginID string
	Fn       func()
}

// startupHooks holds the registered startup tasks in registration order.
var startupHooks []StartupHook

// RegisterStartupHook registers a plugin startup task. pluginID is the owning
// plugin id (empty for shell-level hooks). Panics on a nil fn.
func RegisterStartupHook(pluginID string, fn func()) {
	if fn == nil {
		panic("RegisterStartupHook: nil hook")
	}
	startupHooks = append(startupHooks, StartupHook{PluginID: pluginID, Fn: fn})
}

// RunStartupHooks invokes registered hooks in order. enabled(pluginID) gates
// each hook (skipped when enabled != nil and it returns false); each hook runs
// with panic isolation (recover + log, does not stop the rest).
func RunStartupHooks(enabled func(pluginID string) bool) {
	for _, hook := range startupHooks {
		if enabled != nil && !enabled(hook.PluginID) {
			continue
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("startup hook panicked", "plugin", hook.PluginID, "recover", r)
				}
			}()
			hook.Fn()
		}()
	}
}

// StartupHooks returns a snapshot (for tests/PluginInfo counts if needed).
func StartupHooks() []StartupHook {
	hooks := make([]StartupHook, len(startupHooks))
	copy(hooks, startupHooks)
	return hooks
}

// RegisterStartupHookWithScope registers a startup hook on a scope instead of
// the package-level registry. pluginID names the owning plugin (empty for
// shell-level hooks). The hook is collected by (Scope).StartupHooks and is the
// scope-driven replacement for RegisterStartupHook; consumers switch to it in
// a later phase. Panics on a nil fn or nil scope.
func RegisterStartupHookWithScope(scope *Scope, pluginID string, fn func()) {
	if scope == nil {
		panic("RegisterStartupHookWithScope: nil scope")
	}
	if fn == nil {
		panic("RegisterStartupHookWithScope: nil hook")
	}
	scope.mu.Lock()
	scope.startupHooks = append(scope.startupHooks, StartupHook{PluginID: pluginID, Fn: fn})
	scope.mu.Unlock()
}

// StartupHooks returns the runnable startup hooks collected on the scope: the
// scope's own registered hooks (RegisterStartupHookWithScope) followed by the
// hooks of enabled plugins implementing StartupHookContributor (in registration
// order), then child scopes recursively. Plugins registered disabled
// (AddWithEnabled with enabled=false) are skipped. It is the scope-driven
// replacement for the package-level startup-hook registry and is meant to be
// consumed by the RunStartupHooks successor in a later phase.
func (s *Scope) StartupHooks() []func() {
	s.mu.Lock()
	hooks := make([]func(), 0, len(s.startupHooks))
	for _, h := range s.startupHooks {
		hooks = append(hooks, h.Fn)
	}
	plugins := append([]Plugin(nil), s.plugins...)
	enabled := append([]bool(nil), s.enabled...)
	children := append([]*Scope(nil), s.children...)
	s.mu.Unlock()

	for i, p := range plugins {
		if !enabled[i] {
			continue
		}
		if c, ok := p.(StartupHookContributor); ok {
			if fn := c.StartupHook(); fn != nil {
				hooks = append(hooks, fn)
			}
		}
	}
	for _, child := range children {
		hooks = append(hooks, child.StartupHooks()...)
	}
	return hooks
}
