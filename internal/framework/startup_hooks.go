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
