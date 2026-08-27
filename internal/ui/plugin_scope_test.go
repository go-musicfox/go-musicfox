package ui

import (
	"reflect"
	"sort"
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
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
// registerUIExtraServices call in NewNetease.
func TestFrontendScopeStartRegistersUIExtraServices(t *testing.T) {
	ctx := &framework.Context{}
	scope := NewFrontendScope(&core.Engine{}, testNetease())
	if err := scope.Start(ctx); err != nil {
		t.Fatalf("frontend scope Start() error = %v", err)
	}

	got := ctx.Names()
	sort.Strings(got)
	want := []string{ServiceCoverRenderer, ServiceMenuRegistry, ServicePageRegistry}
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
