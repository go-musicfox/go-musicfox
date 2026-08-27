package framework

import "testing"

// TestRegisterPluginAndPluginConstructors proves RegisterPlugin /
// PluginConstructors round trip: constructors resolve by id and produce a
// working Plugin, and the snapshot is detached from the live registry.
func TestRegisterPluginAndPluginConstructors(t *testing.T) {
	const idA = "registry_test_plugin_a"
	const idB = "registry_test_plugin_b"
	var history []string
	RegisterPlugin(idA, func() Plugin { return &trackingPlugin{name: "a", history: &history} })
	RegisterPlugin(idB, func() Plugin { return NoopPlugin{} })

	constructors := PluginConstructors()

	ctorA, ok := constructors[idA]
	if !ok {
		t.Fatalf("PluginConstructors() missing %q", idA)
	}
	p := ctorA()
	if p == nil {
		t.Fatal("constructor for a returned nil Plugin")
	}
	if _, ok := p.(*trackingPlugin); !ok {
		t.Fatalf("constructor for a returned %T, want *trackingPlugin", p)
	}
	// The constructed plugin is a live instance: Start runs through the scope
	// lifecycle as a real plugin would.
	if err := p.Start(&Context{}); err != nil {
		t.Fatalf("constructed plugin Start() error = %v", err)
	}
	if len(history) != 1 || history[0] != "start:a" {
		t.Fatalf("history = %v, want [start:a]", history)
	}

	if _, ok := constructors[idB]; !ok {
		t.Fatalf("PluginConstructors() missing %q", idB)
	}
}

// TestPluginConstructorsSnapshotDetached proves the map returned by
// PluginConstructors does not alias the live registry.
func TestPluginConstructorsSnapshotDetached(t *testing.T) {
	const id = "registry_test_snapshot"
	RegisterPlugin(id, func() Plugin { return NoopPlugin{} })

	snapshot := PluginConstructors()
	delete(snapshot, id)
	if _, ok := PluginConstructors()[id]; !ok {
		t.Fatal("PluginConstructors() snapshot aliases the live registry")
	}
}

// TestRegisterPluginValidation proves RegisterPlugin panics on programmer
// errors: empty id, nil constructor and duplicate id.
func TestRegisterPluginValidation(t *testing.T) {
	for name, fn := range map[string]func(){
		"empty id":        func() { RegisterPlugin("", func() Plugin { return NoopPlugin{} }) },
		"nil constructor": func() { RegisterPlugin("registry_test_nil", nil) },
		"duplicate id": func() {
			RegisterPlugin("registry_test_dup", func() Plugin { return NoopPlugin{} })
			RegisterPlugin("registry_test_dup", func() Plugin { return NoopPlugin{} })
		},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("expected panic")
				}
			}()
			fn()
		})
	}
}
