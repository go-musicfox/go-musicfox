package framework

import "sync"

// Package-level plugin constructor registry (P5): compile-time plugins
// register a constructor via RegisterPlugin from their init() (the aggregator
// in internal/plugins keeps blank-importing each plugin package), and each
// frontend scope build resolves the enabled subset through PluginConstructors
// and mounts it with AddWithEnabled. The registry carries no business
// dependency: the constructor is a plain func() Plugin and the enabled
// filtering (configs.IsPluginEnabled) stays on the caller side.

var (
	pluginRegistryMu sync.Mutex
	pluginRegistry   = map[string]func() Plugin{}
)

// RegisterPlugin registers id → plugin constructor (package-level, compile-time
// init() call). Panics on an empty id, a nil constructor or a duplicate id
// (programmer error — a plugin must declare its id exactly once). Registrations
// must happen before any frontend scope build reads PluginConstructors.
func RegisterPlugin(id string, new func() Plugin) {
	if id == "" {
		panic("framework: RegisterPlugin: empty id")
	}
	if new == nil {
		panic("framework: RegisterPlugin: nil constructor for id " + id)
	}
	pluginRegistryMu.Lock()
	defer pluginRegistryMu.Unlock()
	if _, dup := pluginRegistry[id]; dup {
		panic("framework: RegisterPlugin: duplicate id " + id)
	}
	pluginRegistry[id] = new
}

// PluginConstructors returns a snapshot of id → constructor (frontend scope
// build filters by enabled and AddWithEnabled). The returned map does not alias
// the live registry.
func PluginConstructors() map[string]func() Plugin {
	pluginRegistryMu.Lock()
	defer pluginRegistryMu.Unlock()
	snapshot := make(map[string]func() Plugin, len(pluginRegistry))
	for id, ctor := range pluginRegistry {
		snapshot[id] = ctor
	}
	return snapshot
}
