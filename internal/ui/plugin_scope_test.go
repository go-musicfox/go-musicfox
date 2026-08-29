package ui

import (
	"reflect"
	"sort"
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
)

// startWithPluginPlugin is a P5-2-shaped test plugin: its Start re-enters
// ui.WithPlugin and registers contributions, exactly like a migrated business
// plugin will. It proves the WithPlugin attribution stamp works at runtime
// (inside a scope Start) and that the record* helpers stay side-effect-free
// against the PluginInfos snapshot.
type startWithPluginPlugin struct {
	framework.NoopPlugin
	pluginID string
	name     string
	menuKey  string
	pageKey  string
}

func (p *startWithPluginPlugin) Start(*framework.Context) error {
	WithPlugin(p.pluginID, p.name, func() {
		RegisterMenu(p.menuKey, func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
			return &testCheckUpdateMenu{baseMenu: base}, nil
		})
		RegisterPage(p.pageKey, func(opts LoginPageOpts) (model.Page, error) {
			return &LoginPage{}, nil
		})
		RegisterStartupHook(func() {})
	})
	return nil
}

// TestFrontendScopeStartRegistersUIExtraServices proves the frontend scope's
// uiServicesPlugin registers exactly the three TUI-only services when the
// scope starts — the scoped replacement of the former direct
// registerUIExtraServices call in NewNetease — plus the WASM sub-scope's
// manager service (P6: NewFrontendScope builds the wasm child scope holding
// wasm.ManagerPlugin, whose Start provides ServiceWasmManager).
func TestFrontendScopeStartRegistersUIExtraServices(t *testing.T) {
	ctx := &framework.Context{}
	scope := NewFrontendScope(&core.Engine{}, testNetease())
	if err := scope.Start(ctx); err != nil {
		t.Fatalf("frontend scope Start() error = %v", err)
	}

	got := ctx.Names()
	sort.Strings(got)
	want := []string{ServiceCoverRenderer, ServiceMenuRegistry, ServicePageRegistry, wasm.ServiceWasmManager}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("frontend scope registered services = %v, want %v", got, want)
	}

	if err := scope.Dispose(); err != nil {
		t.Fatalf("frontend scope Dispose() error = %v", err)
	}
}

// TestFrontendScopeStartWithPluginAttribution proves the WithPlugin Start-path
// contract (P5-2 shape): a plugin calling WithPlugin from its Start gets its
// registrations recorded under its PluginInfo, and the PluginInfos snapshot
// keeps working after a runtime declaration.
func TestFrontendScopeStartWithPluginAttribution(t *testing.T) {
	const (
		pluginID = "frontend_scope_owner"
		menuKey  = "frontend_scope_menu"
		pageKey  = "frontend_scope_page"
	)
	scope := framework.NewScope()
	if err := scope.Add(&startWithPluginPlugin{
		pluginID: pluginID,
		name:     "Start 归属",
		menuKey:  menuKey,
		pageKey:  pageKey,
	}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := scope.Start(&framework.Context{}); err != nil {
		t.Fatalf("scope Start() error = %v", err)
	}

	info := pluginInfoSnapshot(t, pluginID)
	if info == nil {
		t.Fatalf("plugin %q not declared after scope Start", pluginID)
	}
	if info.Name != "Start 归属" {
		t.Fatalf("PluginInfo name = %q, want 首次声明名", info.Name)
	}
	if !containsString(info.MenuKeys, menuKey) {
		t.Fatalf("MenuKeys = %v, want to contain %q", info.MenuKeys, menuKey)
	}
	if !containsString(info.PageKeys, pageKey) {
		t.Fatalf("PageKeys = %v, want to contain %q", info.PageKeys, pageKey)
	}
	if info.StartupHooks != 1 {
		t.Fatalf("StartupHooks = %d, want 1", info.StartupHooks)
	}

	// The snapshot keeps working after a runtime (Start-time) WithPlugin call.
	if snapshot := PluginInfos(); len(snapshot) == 0 {
		t.Fatal("PluginInfos() empty after runtime WithPlugin registration")
	}

	if err := scope.Dispose(); err != nil {
		t.Fatalf("scope Dispose() error = %v", err)
	}
}

// TestNewFrontendScopeValidation proves NewFrontendScope rejects programmer
// errors (nil engine / nil Netease) with a panic, matching the codebase's
// fail-loud bootstrap style.
func TestNewFrontendScopeValidation(t *testing.T) {
	assertPanics(t, func() { NewFrontendScope(nil, testNetease()) })
	assertPanics(t, func() { NewFrontendScope(&core.Engine{}, nil) })
}

// scopeIdentityPlugin is a P8-shaped plugin: it implements framework.
// PluginIdentity AND re-enters ui.WithPlugin in Start (exactly like the 9
// business plugins, which NewFrontendScope additionally wraps in
// identifiedPlugin). It embeds PluginBase so AddWithEnabled records the
// enabled flag, and exposes it via IsEnabled (the same contract
// identifiedPlugin and wasmPlugin satisfy).
type scopeIdentityPlugin struct {
	framework.NoopPlugin
	framework.PluginBase
	id   string
	name string
	key  string
}

func (p *scopeIdentityPlugin) PluginID() string   { return p.id }
func (p *scopeIdentityPlugin) PluginName() string { return p.name }
func (p *scopeIdentityPlugin) IsEnabled() bool    { return p.Enabled }

func (p *scopeIdentityPlugin) Start(*framework.Context) error {
	WithPlugin(p.id, p.name, func() {
		RegisterMenu(p.key, func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
			return &testCheckUpdateMenu{baseMenu: base}, nil
		})
	})
	return nil
}

// TestPluginInfosCollectedFromScope proves the P8 scope-driven PluginInfos():
// with an active frontend scope the plugin set is collected from the scope's
// plugins (identity + WithPlugin record); a plugin registered with
// AddWithEnabled(..., false) — present in the scope slice but never started —
// is excluded.
func TestPluginInfosCollectedFromScope(t *testing.T) {
	const (
		enabledID  = "scope_collect_enabled"
		disabledID = "scope_collect_disabled"
		menuKey    = "scope_collect_menu"
	)
	prevScope := activeFrontendScope
	t.Cleanup(func() { activeFrontendScope = prevScope })

	scope := framework.NewScope()
	if err := scope.AddWithEnabled(&scopeIdentityPlugin{id: enabledID, name: "Scope 启用", key: menuKey}, true); err != nil {
		t.Fatalf("AddWithEnabled(enabled) error = %v", err)
	}
	if err := scope.AddWithEnabled(&scopeIdentityPlugin{id: disabledID, name: "Scope 禁用", key: menuKey + "_disabled"}, false); err != nil {
		t.Fatalf("AddWithEnabled(disabled) error = %v", err)
	}
	if err := scope.Start(&framework.Context{}); err != nil {
		t.Fatalf("scope Start() error = %v", err)
	}
	activeFrontendScope = scope

	infos := PluginInfos()
	var enabledInfo, disabledInfo *PluginInfo
	for i := range infos {
		switch infos[i].ID {
		case enabledID:
			enabledInfo = &infos[i]
		case disabledID:
			disabledInfo = &infos[i]
		}
	}
	if enabledInfo == nil {
		t.Fatalf("enabled plugin %q missing from scope-collected PluginInfos: %+v", enabledID, infos)
	}
	if !containsString(enabledInfo.MenuKeys, menuKey) {
		t.Fatalf("enabled plugin MenuKeys = %v, want to contain %q", enabledInfo.MenuKeys, menuKey)
	}
	if disabledInfo != nil {
		t.Fatalf("disabled plugin %q appeared in PluginInfos (AddWithEnabled false = never started = excluded)", disabledID)
	}
}

// TestDisabledPluginNotStartedRegistersNothing proves the P5 "disabled =
// nonexistent" semantics at the scope boundary: a plugin registered with
// AddWithEnabled(..., false) never starts, so its menu/page contributions are
// NOT registered and BuildMenu fails with the missing-key error — the former
// "disabled plugin key still jumpable" contract is intentionally gone. The
// enabled path starts the same plugin shape and registers the contributions.
func TestDisabledPluginNotStartedRegistersNothing(t *testing.T) {
	const (
		disabledID = "disabled_semantics_plugin"
		menuKey    = "disabled_semantics_menu"
		pageKey    = "disabled_semantics_page"
	)

	// Disabled: Start is skipped → nothing registered, BuildMenu errors.
	scope := framework.NewScope()
	if err := scope.AddWithEnabled(&startWithPluginPlugin{
		pluginID: disabledID,
		name:     "禁用语义",
		menuKey:  menuKey,
		pageKey:  pageKey,
	}, false); err != nil {
		t.Fatalf("AddWithEnabled() error = %v", err)
	}
	if err := scope.Start(&framework.Context{}); err != nil {
		t.Fatalf("scope Start() error = %v", err)
	}
	if (MenuRegistry{}).Registered(menuKey) {
		t.Fatalf("disabled plugin menu %q is registered (Start must have been skipped)", menuKey)
	}
	if (PageRegistry{}).Registered(pageKey) {
		t.Fatalf("disabled plugin page %q is registered (Start must have been skipped)", pageKey)
	}
	if _, err := BuildMenu(menuKey, baseMenu{}, NoArgMenuOpts{}); err == nil {
		t.Fatalf("BuildMenu(%q) succeeded for a disabled plugin, want missing-key error", menuKey)
	}

	// Enabled: Start runs → the contributions are registered.
	const enabledKey = "disabled_semantics_enabled_menu"
	scope2 := framework.NewScope()
	if err := scope2.AddWithEnabled(&startWithPluginPlugin{
		pluginID: disabledID + "_enabled",
		name:     "禁用语义启用",
		menuKey:  enabledKey,
		pageKey:  pageKey + "_enabled",
	}, true); err != nil {
		t.Fatalf("AddWithEnabled() error = %v", err)
	}
	if err := scope2.Start(&framework.Context{}); err != nil {
		t.Fatalf("scope Start() error = %v", err)
	}
	if !(MenuRegistry{}).Registered(enabledKey) {
		t.Fatalf("enabled plugin menu %q not registered after Start", enabledKey)
	}
}
