package ui

import (
	"sync"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

// --- Plugin attribution registry (plugin configurable enable/disable) ---
//
// Plugins declare their identity and wrap all their registrations in
// ui.WithPlugin(id, name, func() { ... }) so the shell can attribute the
// registrations made inside the scope to that plugin. The shell then filters
// per-plugin *entry visibility* and *startup side effects* against the
// [plugins] disabled config: a disabled plugin's main-menu items are hidden and
// its startup hooks are skipped. The menu/page provider registries and
// BuildMenu jumps are NOT filtered — a disabled plugin's menus stay fully
// buildable by key, preserving cross-plugin jump integrity.
//
// Registration window (P5 transition): today the 9 business plugins call
// WithPlugin from their package init() (compile-time registration). From P5-2
// they call it from their plugin Start inside the frontend scope instead — the
// attribution stamp (currentPluginID) is a plain set/unset guard that works
// identically in both windows, and the record* helpers below are lock-guarded
// against a snapshot, so a runtime (Start-time) call has no side effects on
// PluginInfos(). The init() path stays supported until every plugin migrates.

// PluginInfo describes a declared plugin: its id/name and the registrations
// attributed to it (registration order preserved).
type PluginInfo struct {
	ID   string
	Name string

	// MenuKeys are the menu provider keys registered inside the plugin's
	// WithPlugin scope(s), in registration order.
	MenuKeys []string
	// PageKeys are the page provider keys registered inside the scope(s).
	PageKeys []string
	// MainMenuItems are the plugin-declared main-menu entry keys, in
	// registration order.
	MainMenuItems []string
	// CommandKeys are the track-B command keys registered inside the scope(s),
	// in registration order.
	CommandKeys []string
	// StartupHooks is the number of startup hooks registered inside the
	// scope(s).
	StartupHooks int
}

var (
	pluginMu        sync.Mutex
	currentPluginID string
	pluginInfos     []*PluginInfo
	pluginInfoByID  = map[string]*PluginInfo{}
)

// WithPlugin declares the plugin id/name owning the registrations made inside
// register: RegisterMenu / RegisterPage / RegisterMainMenuItem* /
// RegisterStartupHook calls in the scope are recorded under that plugin. The
// same id may be declared multiple times (e.g. a plugin split across several
// init() files, or the command-menu adapter re-declaring a command's plugin
// id); declarations merge idempotently onto the first PluginInfo — only the
// first declaration's name is kept. Panics on an empty id or a nil register
// func.
//
// Callable both at compile-time init() (the 9 business plugins, P5-2 before)
// and at runtime inside a plugin's Start (P5-2 after — the frontend scope
// starts plugins in registration order, each Start re-enters WithPlugin with
// its own id; the previous id is restored on exit). Package init() is
// single-threaded, but the guard lock keeps the state sound even when a scope
// body spawns goroutines or Start calls run interleaved with WASM loading.
func WithPlugin(id, name string, register func()) {
	if id == "" || register == nil {
		panic("WithPlugin: empty id or nil register func")
	}
	pluginMu.Lock()
	if _, ok := pluginInfoByID[id]; !ok {
		info := &PluginInfo{ID: id, Name: name}
		pluginInfoByID[id] = info
		pluginInfos = append(pluginInfos, info)
	}
	prev := currentPluginID
	currentPluginID = id
	pluginMu.Unlock()

	defer func() {
		pluginMu.Lock()
		currentPluginID = prev
		pluginMu.Unlock()
	}()
	register()
}

// recordPluginMenuKey attributes a registered menu provider key to the current
// plugin scope. Registrations outside any scope (empty current id, e.g. the
// built-in bootstrap or test binaries) are not attributed to any plugin.
func recordPluginMenuKey(key string) {
	pluginMu.Lock()
	defer pluginMu.Unlock()
	if info := pluginInfoByID[currentPluginID]; info != nil {
		info.MenuKeys = append(info.MenuKeys, key)
	}
}

// recordPluginPageKey is the page-provider analog of recordPluginMenuKey.
func recordPluginPageKey(key string) {
	pluginMu.Lock()
	defer pluginMu.Unlock()
	if info := pluginInfoByID[currentPluginID]; info != nil {
		info.PageKeys = append(info.PageKeys, key)
	}
}

// recordPluginMainMenuItemKey attributes a registered main-menu entry key to
// the current plugin scope.
func recordPluginMainMenuItemKey(key string) {
	pluginMu.Lock()
	defer pluginMu.Unlock()
	if info := pluginInfoByID[currentPluginID]; info != nil {
		info.MainMenuItems = append(info.MainMenuItems, key)
	}
}

// recordPluginCommandKey attributes a registered track-B command key to the
// current plugin scope. Registrations outside any scope (empty current id) are
// not attributed to any plugin. It mirrors recordPluginMenuKey.
func recordPluginCommandKey(key string) {
	pluginMu.Lock()
	defer pluginMu.Unlock()
	if info := pluginInfoByID[currentPluginID]; info != nil {
		info.CommandKeys = append(info.CommandKeys, key)
	}
}

// recordPluginStartupHook attributes a registered startup hook to the plugin
// id that was current at registration time (the id is captured by
// RegisterStartupHook before the record call).
func recordPluginStartupHook(pluginID string) {
	if pluginID == "" {
		return
	}
	pluginMu.Lock()
	defer pluginMu.Unlock()
	if info := pluginInfoByID[pluginID]; info != nil {
		info.StartupHooks++
	}
}

// PluginInfos returns a snapshot copy of the declared plugins in declaration
// order (callers may mutate the returned slices freely).
func PluginInfos() []PluginInfo {
	pluginMu.Lock()
	defer pluginMu.Unlock()
	infos := make([]PluginInfo, 0, len(pluginInfos))
	for _, info := range pluginInfos {
		infos = append(infos, PluginInfo{
			ID:            info.ID,
			Name:          info.Name,
			MenuKeys:      append([]string(nil), info.MenuKeys...),
			PageKeys:      append([]string(nil), info.PageKeys...),
			MainMenuItems: append([]string(nil), info.MainMenuItems...),
			CommandKeys:   append([]string(nil), info.CommandKeys...),
			StartupHooks:  info.StartupHooks,
		})
	}
	return infos
}

// IsPluginEnabled reports whether the plugin id is enabled under the current
// [plugins] config. An id that was never declared (or has no plugin scope) is
// treated as enabled; a nil configs.AppConfig (e.g. tests without a config) is
// also treated as enabled. It delegates to the configs-layer helper so the
// engine (which must not import ui) can gate startup hooks with the same rule.
func IsPluginEnabled(id string) bool {
	return configs.IsPluginEnabled(id)
}
