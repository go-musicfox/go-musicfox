package framework

import (
	"reflect"
	"testing"
)

func TestRegisterStartupHookRejectsNil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("RegisterStartupHook(nil) did not panic")
		}
	}()
	RegisterStartupHook("test", nil)
}

func TestRunStartupHooksOrderAndPanicIsolation(t *testing.T) {
	var order []string
	RegisterStartupHook("", func() { order = append(order, "first") })
	RegisterStartupHook("", func() { panic("boom") })
	RegisterStartupHook("", func() { order = append(order, "third") })

	RunStartupHooks(nil)

	if len(order) != 2 || order[0] != "first" || order[1] != "third" {
		t.Fatalf("startup hooks ran in order %v, want [first third]", order)
	}
}

func TestRunStartupHooksEnabledGating(t *testing.T) {
	var ran []string
	RegisterStartupHook("always", func() { ran = append(ran, "always") })
	RegisterStartupHook("disabled", func() { ran = append(ran, "disabled") })

	enabled := func(pluginID string) bool { return pluginID != "disabled" }
	RunStartupHooks(enabled)

	if len(ran) != 1 || ran[0] != "always" {
		t.Fatalf("gated hooks ran = %v, want [always]", ran)
	}
}

func TestStartupHooksSnapshot(t *testing.T) {
	before := len(StartupHooks())
	RegisterStartupHook("snap", func() {})
	snapshot := StartupHooks()
	if len(snapshot) != before+1 {
		t.Fatalf("snapshot len = %d, want %d", len(snapshot), before+1)
	}
	if snapshot[len(snapshot)-1].PluginID != "snap" || snapshot[len(snapshot)-1].Fn == nil {
		t.Fatalf("snapshot entry = %+v, want plugin id snap with non-nil fn", snapshot[len(snapshot)-1])
	}
	// The snapshot must not alias the live registry: mutating an element must
	// not affect the live entries.
	origID := StartupHooks()[0].PluginID
	snapshot[0].PluginID = "mutated"
	if hooks := StartupHooks(); hooks[0].PluginID != origID {
		t.Fatal("StartupHooks() snapshot aliases the live registry")
	}
}

// --- scope-collected startup hooks ---

// hookPlugin embeds NoopPlugin for the lifecycle and contributes a startup
// hook via StartupHookContributor.
type hookPlugin struct {
	NoopPlugin
	name string
	fn   func()
}

func (p *hookPlugin) StartupHook() func() { return p.fn }

func TestScopeStartupHooksCollectsContributors(t *testing.T) {
	scope := NewScope()
	var ran []string
	_ = scope.Add(&hookPlugin{name: "enabled", fn: func() { ran = append(ran, "enabled") }})
	_ = scope.AddWithEnabled(&hookPlugin{name: "disabled", fn: func() { ran = append(ran, "disabled") }}, false)
	_ = scope.Add(&trackingPlugin{name: "no-hook", history: &[]string{}})

	hooks := scope.StartupHooks()
	if len(hooks) != 1 {
		t.Fatalf("StartupHooks() len = %d, want 1", len(hooks))
	}
	for _, fn := range hooks {
		fn()
	}
	if len(ran) != 1 || ran[0] != "enabled" {
		t.Fatalf("ran = %v, want [enabled] (disabled plugin's hook must be skipped)", ran)
	}
}

func TestScopeStartupHooksIncludesScopeRegisteredAndChildHooks(t *testing.T) {
	scope := NewScope()
	child := scope.NewScope()
	var ran []string
	RegisterStartupHookWithScope(scope, "", func() { ran = append(ran, "scope") })
	RegisterStartupHookWithScope(child, "", func() { ran = append(ran, "child") })
	_ = scope.Add(&hookPlugin{name: "p", fn: func() { ran = append(ran, "plugin") }})

	hooks := scope.StartupHooks()
	if len(hooks) != 3 {
		t.Fatalf("StartupHooks() len = %d, want 3", len(hooks))
	}
	for _, fn := range hooks {
		fn()
	}
	if !reflect.DeepEqual(ran, []string{"scope", "plugin", "child"}) {
		t.Fatalf("ran = %v, want [scope plugin child]", ran)
	}
}

func TestRegisterStartupHookWithScopeRejectsNil(t *testing.T) {
	for name, fn := range map[string]func(){
		"nil scope": func() { RegisterStartupHookWithScope(nil, "", func() {}) },
		"nil hook":  func() { RegisterStartupHookWithScope(NewScope(), "", nil) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s did not panic", name)
				}
			}()
			fn()
		})
	}
}
