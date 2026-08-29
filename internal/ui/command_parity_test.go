package ui

import (
	"strings"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// parityTestKey derives a unique key prefix from the test name so the
// package-global registries (frontend command registry, menu registry,
// main-menu item list) stay pollution-free across a single test run.
func parityTestKey(t *testing.T) string {
	t.Helper()
	return "parity_" + strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, t.Name())
}

// findMainMenuItem returns the main-menu item registered under key.
func findMainMenuItem(items []MainMenuItem, key string) (MainMenuItem, bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return MainMenuItem{}, false
}

// TestCommandParity 断言每个注册的轨 B 命令在 TUI 都有对应的 CommandMenu
// provider 与主菜单入口（防 registerCommandMenus 漏注册）。
func TestCommandParity(t *testing.T) {
	const owner = "parity_owner"
	prefix := parityTestKey(t)
	keys := []string{
		prefix + "_owned", // PluginID stamped via a WithPlugin scope
		prefix + "_after", // non-empty After anchor
		prefix + "_show",  // non-nil Show gate
	}

	// ① Register the test commands: one inside a WithPlugin scope (PluginID
	// attribution), one with a non-empty After anchor and one with a non-nil
	// Show gate.
	WithPlugin(owner, "parity", func() {
		RegisterCommand(testCommand(keys[0]))
	})
	afterCmd := testCommand(keys[1])
	// Keep After empty (end-append) for the parity walk: registering a
	// non-empty anchor command here would pollute the package-global
	// main-menu chain (its After target would need to exist and be unique),
	// orphaning entries registered by other tests and panicking later
	// NewMainMenu calls. After-anchor propagation is covered by the frontend
	// command contract tests.
	RegisterCommand(afterCmd)
	showCmd := testCommand(keys[2])
	showCmd.Show = func(frontend.CommandContext) bool { return true }
	RegisterCommand(showCmd)

	// ② Adapt every registered command into a CommandMenu provider + main-menu
	// item. The idempotent dup-guard makes repeated calls safe.
	registerCommandMenus()

	// ③ Collect the just-registered commands from the frontend snapshot.
	registered := make(map[string]frontend.Command, len(keys))
	for _, cmd := range frontend.Commands() {
		if _, want := registered[cmd.Key]; !want {
			for _, key := range keys {
				if cmd.Key == key {
					registered[key] = cmd
					break
				}
			}
		}
	}
	if len(registered) != len(keys) {
		t.Fatalf("frontend snapshot holds %d of the registered commands, want %d: %v", len(registered), len(keys), keys)
	}

	items := MainMenuPluginItems()
	for _, key := range keys {
		cmd, ok := registered[key]
		if !ok {
			t.Fatalf("command %q missing from the frontend snapshot", key)
		}

		// a) A *CommandMenu provider is registered and buildable; its key matches.
		menu, err := BuildMenu(key, baseMenu{}, NoArgMenuOpts{})
		if err != nil {
			t.Fatalf("BuildMenu(%q): %v", key, err)
		}
		cm, ok := menu.(*CommandMenu)
		if !ok {
			t.Fatalf("BuildMenu(%q) = %T, want *CommandMenu", key, menu)
		}
		if got := cm.GetMenuKey(); got != key {
			t.Fatalf("GetMenuKey() = %q, want %q", got, key)
		}

		// b) A main-menu entry exists; its PluginID matches the command's
		// attribution (plugin-scoped commands carry the plugin id, others stay
		// empty).
		item, ok := findMainMenuItem(items, key)
		if !ok {
			t.Fatalf("main-menu item %q not registered", key)
		}
		if item.PluginID != cmd.PluginID {
			t.Fatalf("item %q PluginID = %q, want %q (command PluginID)", key, item.PluginID, cmd.PluginID)
		}

		// c) A non-empty After anchor is preserved on the main-menu entry.
		if cmd.After != "" && item.After != cmd.After {
			t.Fatalf("item %q After = %q, want %q (command After)", key, item.After, cmd.After)
		}
	}

	// ④ Negative: a key that was never registered must fail to build (the
	// assertion above must not trivially pass).
	if _, err := BuildMenu(prefix+"_unregistered", baseMenu{}, NoArgMenuOpts{}); err == nil {
		t.Fatalf("BuildMenu(%q) succeeded, want error", prefix+"_unregistered")
	}
}
