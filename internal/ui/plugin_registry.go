package ui

import (
	"sync"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/framework"
)

// --- Plugin attribution registry (plugin configurable enable/disable) ---
//
// Plugins declare their identity and wrap all their registrations in
// ui.WithPlugin(id, name, func() { ... }) so the shell can attribute the
// registrations made inside the scope to that plugin.
//
// Since P5-2 the 9 business plugins register from their plugin Start inside
// the frontend scope: a plugin disabled in [plugins] never starts
// (AddWithEnabled(..., false)), so its menus/pages/main-menu items/startup
// hooks are never registered — "disabled = nonexistent". The attribution
// stamp (currentPluginID) is a plain set/unset guard that works identically
// in both the former init() window and the current Start window, and the
// record* helpers below are lock-guarded against a snapshot, so a runtime
// (Start-time) call has no side effects on PluginInfos().
//
// Residual consumption-point filtering stays for registrations that happen
// unconditionally even when their plugin is disabled: WASM command-menu items
// (adapted by registerCommandMenus) are registered for every loaded manifest
// and filtered by IsPluginEnabled in NewMainMenu / commandActionCmd, and
// framework.RunStartupHooks still gates package-registered hooks by plugin id.

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

	// activeFrontendScope is the TUI frontend scope, set by NewFrontendScope.
	// PluginInfos() collects the plugin set from it (the single source of truth
	// for "which plugins are actually mounted and enabled"); when nil (test
	// binaries, standalone WithPlugin usage) the WithPlugin declaration records
	// are used directly.
	activeFrontendScope *framework.Scope
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
// Callable at compile-time init() (e.g. a not-yet-migrated plugin or the ui
// test binary's test-doubles) and at runtime inside a plugin's Start (the 9
// business plugins, P5-2 — the frontend scope starts plugins in registration
// order, each Start re-enters WithPlugin with its own id; the previous id is
// restored on exit). Package init() is single-threaded, but the guard lock
// keeps the state sound even when a scope body spawns goroutines or Start
// calls run interleaved with WASM loading.
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
// not attributed to any plugin. It mirrors recordPluginMenuKey, and additionally
// dedupes: a hot reload (P8) re-registers the same keys through the wasm sink's
// Replace path, which would otherwise append the key once per generation.
func recordPluginCommandKey(key string) {
	pluginMu.Lock()
	defer pluginMu.Unlock()
	if info := pluginInfoByID[currentPluginID]; info != nil {
		if containsKey(info.CommandKeys, key) {
			return
		}
		info.CommandKeys = append(info.CommandKeys, key)
	}
}

// containsKey reports whether ss contains s (linear scan; command keys are few).
func containsKey(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
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

// PluginInfos returns a snapshot copy of the active plugins in declaration
// order (callers may mutate the returned slices freely). Since P8 the source is
// the frontend scope when one is active: the plugin set is collected from
// framework.Scope.Plugins() (recursively through child scopes, so WASM plugin
// adapters are included) — plugins must expose their identity (framework.
// PluginIdentity; the 9 business plugins are wrapped with an id-carrying
// decorator, wasmPlugin implements it) and a plugin that never started
// (AddWithEnabled disabled) is excluded. The per-plugin attribution (keys,
// name) comes from the WithPlugin record populated during Start, or from the
// plugin's contributor interfaces when no record exists. WithPlugin records
// not represented in the scope (test doubles, standalone declarations) are
// merged in afterwards so the API stays total without a scope.
func PluginInfos() []PluginInfo {
	pluginMu.Lock()
	defer pluginMu.Unlock()

	var (
		infos   []*PluginInfo
		covered = map[string]bool{}
	)
	if activeFrontendScope != nil {
		for _, info := range collectScopePluginInfos(activeFrontendScope) {
			infos = append(infos, info)
			covered[info.ID] = true
		}
	}
	for _, info := range pluginInfos {
		if covered[info.ID] {
			continue
		}
		infos = append(infos, info)
	}

	snapshot := make([]PluginInfo, 0, len(infos))
	for _, info := range infos {
		snapshot = append(snapshot, PluginInfo{
			ID:            info.ID,
			Name:          info.Name,
			MenuKeys:      append([]string(nil), info.MenuKeys...),
			PageKeys:      append([]string(nil), info.PageKeys...),
			MainMenuItems: append([]string(nil), info.MainMenuItems...),
			CommandKeys:   append([]string(nil), info.CommandKeys...),
			StartupHooks:  info.StartupHooks,
		})
	}
	return snapshot
}

// collectScopePluginInfos walks scope (and its child scopes) for plugins
// implementing framework.PluginIdentity and returns their *PluginInfo in scope
// registration order. A plugin with a WithPlugin record (its Start ran and
// recorded) contributes that record. A plugin without a record (no
// WithPlugin-based attribution) contributes an info derived from its
// contributor interfaces, but only when it is actually enabled: plugins
// registered with AddWithEnabled(..., false) never start and are excluded even
// though they sit in the scope's plugin slice. Plugins without identity
// (service plugins such as uiServicesPlugin / ManagerPlugin) are excluded.
func collectScopePluginInfos(scope *framework.Scope) []*PluginInfo {
	var infos []*PluginInfo
	collect := func(p framework.Plugin) {
		idn, ok := p.(framework.PluginIdentity)
		if !ok {
			return
		}
		id := idn.PluginID()
		if id == "" {
			return
		}
		if rec := pluginInfoByID[id]; rec != nil {
			// Started: its Start (or the wasm sink) ran WithPlugin and recorded
			// the attribution — the record is the precise registration source.
			infos = append(infos, rec)
			return
		}
		// No record: include only when the plugin is actually enabled. The
		// scope's enabled flag (written by AddWithEnabled/AddAndStart through
		// the EnabledSetter mechanism) is authoritative when the plugin exposes
		// it; configs.IsPluginEnabled is the production fallback.
		if eg, ok := p.(pluginEnabledGetter); ok && !eg.IsEnabled() {
			return
		}
		if !configs.IsPluginEnabled(id) {
			return
		}
		infos = append(infos, pluginInfoFromContributors(id, idn.PluginName(), p))
	}
	for _, p := range scope.Plugins() {
		collect(p)
	}
	for _, child := range scope.Children() {
		for _, p := range child.Plugins() {
			collect(p)
		}
	}
	return infos
}

// pluginEnabledGetter is implemented by plugins that expose their scope-enabled
// state (the flag AddWithEnabled/AddAndStart wrote via the EnabledSetter
// mechanism). The ui business-plugin decorator (identifiedPlugin) and the wasm
// adapter (wasm.Plugin... via wasmPlugin) implement it so PluginInfos can
// exclude plugins that never started.
type pluginEnabledGetter interface {
	IsEnabled() bool
}

// pluginInfoFromContributors builds a PluginInfo from a plugin's contributor
// interfaces (framework.MenuContributor etc.) — the fallback for plugins that
// implement identity but register without a WithPlugin scope.
func pluginInfoFromContributors(id, name string, p framework.Plugin) *PluginInfo {
	info := &PluginInfo{ID: id, Name: name}
	if m, ok := p.(framework.MenuContributor); ok {
		info.MenuKeys = append([]string(nil), m.MenuKeys()...)
	}
	if pg, ok := p.(framework.PageContributor); ok {
		info.PageKeys = append([]string(nil), pg.PageKeys()...)
	}
	if mm, ok := p.(framework.MainMenuContributor); ok {
		info.MainMenuItems = append([]string(nil), mm.MainMenuKeys()...)
	}
	if c, ok := p.(framework.CommandContributor); ok {
		info.CommandKeys = append([]string(nil), c.CommandKeys()...)
	}
	if sh, ok := p.(framework.StartupHookContributor); ok && sh.StartupHook() != nil {
		info.StartupHooks = 1
	}
	return info
}

// IsPluginEnabled reports whether the plugin id is enabled under the current
// [plugins] config. An id that was never declared (or has no plugin scope) is
// treated as enabled; a nil configs.AppConfig (e.g. tests without a config) is
// also treated as enabled. It delegates to the configs-layer helper so the
// engine (which must not import ui) can gate startup hooks with the same rule.
func IsPluginEnabled(id string) bool {
	return configs.IsPluginEnabled(id)
}
