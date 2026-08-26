package ui

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// testCommand builds a minimal registerable track-B command with a unique key.
func testCommand(key string) frontend.Command {
	return frontend.Command{
		Key:   key,
		Title: "test " + key,
		Run:   func(frontend.CommandContext) frontend.CommandResult { return frontend.CommandResult{} },
	}
}

// --- RegisterCommand attribution ---

// TestRegisterCommandOutsideScope proves a command registered outside any
// WithPlugin scope gets no plugin attribution: cmd.PluginID stays empty and the
// key is not recorded under any plugin's CommandKeys.
func TestRegisterCommandOutsideScope(t *testing.T) {
	const key = "register_cmd_outside"
	RegisterCommand(testCommand(key))

	cmd, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered in frontend registry", key)
	}
	if cmd.PluginID != "" {
		t.Fatalf("PluginID = %q, want empty for registration outside any scope", cmd.PluginID)
	}
	for _, info := range PluginInfos() {
		if containsString(info.CommandKeys, key) {
			t.Fatalf("unattributed command %q recorded under plugin %q", key, info.ID)
		}
	}
}

// TestRegisterCommandInScope proves a command registered inside a WithPlugin
// scope gets the scope's plugin id stamped onto PluginID and the key recorded
// into that plugin's PluginInfo.CommandKeys.
func TestRegisterCommandInScope(t *testing.T) {
	const (
		pluginID = "register_cmd_owner"
		key      = "register_cmd_attr"
	)
	WithPlugin(pluginID, "命令归属", func() {
		RegisterCommand(testCommand(key))
	})

	cmd, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered in frontend registry", key)
	}
	if cmd.PluginID != pluginID {
		t.Fatalf("PluginID = %q, want %q (stamped from WithPlugin scope)", cmd.PluginID, pluginID)
	}

	info := pluginInfoSnapshot(t, pluginID)
	if info == nil {
		t.Fatalf("plugin %q not declared", pluginID)
	}
	if !containsString(info.CommandKeys, key) {
		t.Fatalf("CommandKeys = %v, want to contain %q", info.CommandKeys, key)
	}
}

// TestRegisterCommandStampedPluginIDIdempotent proves a command whose PluginID
// is already set keeps it even inside a WithPlugin scope of a different id.
func TestRegisterCommandStampedPluginIDIdempotent(t *testing.T) {
	const (
		pluginID = "register_cmd_stamp_owner"
		key      = "register_cmd_prestamped"
	)
	cmd := testCommand(key)
	cmd.PluginID = "prestamped"
	WithPlugin(pluginID, "盖章幂等", func() {
		RegisterCommand(cmd)
	})

	got, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered in frontend registry", key)
	}
	if got.PluginID != "prestamped" {
		t.Fatalf("PluginID = %q, want %q (pre-set id must not be overwritten)", got.PluginID, "prestamped")
	}
}

// TestRegisterCommandDuplicateKeyPanics proves duplicate command keys panic
// (delegated to frontend registry semantics) and the failed registration leaves
// no attribution record behind.
func TestRegisterCommandDuplicateKeyPanics(t *testing.T) {
	const (
		pluginID = "register_cmd_dup_owner"
		key      = "register_cmd_dup"
	)
	WithPlugin(pluginID, "重复键", func() {
		RegisterCommand(testCommand(key))
		assertPanics(t, func() { RegisterCommand(testCommand(key)) })
	})

	info := pluginInfoSnapshot(t, pluginID)
	if info == nil {
		t.Fatalf("plugin %q not declared", pluginID)
	}
	if len(info.CommandKeys) != 1 {
		t.Fatalf("CommandKeys = %v, want exactly 1 entry (failed duplicate must not be recorded)", info.CommandKeys)
	}
}

// TestRegisterCommandDuplicateIDMerges proves declaring the same plugin id
// multiple times merges idempotently onto the first PluginInfo: command keys
// from all scopes are recorded, and no second PluginInfo is created.
func TestRegisterCommandDuplicateIDMerges(t *testing.T) {
	const pluginID = "register_cmd_merge_owner"
	WithPlugin(pluginID, "第一次声明", func() {
		RegisterCommand(testCommand("register_cmd_merge_a"))
	})
	// Same id again: must not panic.
	WithPlugin(pluginID, "第二次声明", func() {
		RegisterCommand(testCommand("register_cmd_merge_b"))
	})

	count := 0
	for _, info := range PluginInfos() {
		if info.ID != pluginID {
			continue
		}
		count++
		if info.Name != "第一次声明" {
			t.Fatalf("name = %q, want first declaration name 第一次声明", info.Name)
		}
		if !containsString(info.CommandKeys, "register_cmd_merge_a") ||
			!containsString(info.CommandKeys, "register_cmd_merge_b") {
			t.Fatalf("CommandKeys = %v, want both scopes' keys recorded", info.CommandKeys)
		}
	}
	if count != 1 {
		t.Fatalf("plugin %q declared %d times, want exactly 1 PluginInfo", pluginID, count)
	}
}

// TestPluginInfosCommandKeysSnapshot proves the PluginInfos() snapshot copies
// CommandKeys: mutating the returned slice must not corrupt the live registry.
func TestPluginInfosCommandKeysSnapshot(t *testing.T) {
	const (
		pluginID = "register_cmd_snapshot_owner"
		key      = "register_cmd_snapshot"
	)
	WithPlugin(pluginID, "快照", func() {
		RegisterCommand(testCommand(key))
	})

	snapshot := PluginInfos()
	for i := range snapshot {
		if snapshot[i].ID == pluginID {
			snapshot[i].CommandKeys = append(snapshot[i].CommandKeys, "polluted")
		}
	}
	info := pluginInfoSnapshot(t, pluginID)
	if info == nil {
		t.Fatalf("plugin %q not declared", pluginID)
	}
	if containsString(info.CommandKeys, "polluted") {
		t.Fatal("PluginInfos() snapshot aliases the live CommandKeys registry")
	}
}
