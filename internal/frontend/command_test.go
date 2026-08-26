package frontend

import (
	"encoding/json"
	"testing"
)

func validCommand(key string) Command {
	return Command{
		Key:   key,
		Title: "Title " + key,
		Run: func(CommandContext) CommandResult {
			return CommandResult{Action: "toast", Message: "ok"}
		},
	}
}

// resetCommandRegistry clears the package-level command registry so each test
// controls its own registration state.
func resetCommandRegistry() {
	mu.Lock()
	defer mu.Unlock()
	cmdRegistry = map[string]Command{}
	cmdOrder = nil
}

func TestRegisterCommandAndLookup(t *testing.T) {
	resetCommandRegistry()
	defer resetCommandRegistry()

	cmd := validCommand("cmd-a")
	RegisterCommand(cmd)

	got, ok := CommandByKey("cmd-a")
	if !ok {
		t.Fatal("CommandByKey miss after RegisterCommand")
	}
	// reflect.DeepEqual cannot compare func fields, so compare fields
	// explicitly.
	if got.Key != cmd.Key || got.Title != cmd.Title ||
		got.After != cmd.After || got.PluginID != cmd.PluginID {
		t.Fatalf("CommandByKey() = %+v, want %+v", got, cmd)
	}
	if got.Run == nil {
		t.Fatal("CommandByKey() returned a command with nil Run")
	}

	if _, ok := CommandByKey("missing"); ok {
		t.Fatal("CommandByKey hit for unregistered key")
	}
}

func TestCommandsInRegistrationOrder(t *testing.T) {
	resetCommandRegistry()
	defer resetCommandRegistry()

	for _, key := range []string{"order-1", "order-2", "order-3"} {
		RegisterCommand(validCommand(key))
	}

	got := Commands()
	if len(got) != 3 {
		t.Fatalf("Commands() len = %d, want 3", len(got))
	}
	for i, key := range []string{"order-1", "order-2", "order-3"} {
		if got[i].Key != key {
			t.Fatalf("Commands()[%d].Key = %q, want %q", i, got[i].Key, key)
		}
	}
}

func TestCommandsReturnsCopy(t *testing.T) {
	resetCommandRegistry()
	defer resetCommandRegistry()

	RegisterCommand(validCommand("copy-a"))
	RegisterCommand(validCommand("copy-b"))

	cmds := Commands()
	// Mutating the returned slice must not affect the registry.
	cmds[0].Key = "mutated"
	cmds[0].Run = nil
	cmds[1] = Command{}

	got := Commands()
	if len(got) != 2 {
		t.Fatalf("Commands() len after mutation = %d, want 2", len(got))
	}
	if got[0].Key != "copy-a" || got[0].Run == nil {
		t.Fatalf("registry mutated: Commands()[0] = %+v, want key %q with non-nil Run", got[0], "copy-a")
	}
	if got[1].Key != "copy-b" || got[1].Run == nil {
		t.Fatalf("registry mutated: Commands()[1] = %+v, want key %q with non-nil Run", got[1], "copy-b")
	}
}

func TestRegisterCommandEmptyKeyPanics(t *testing.T) {
	resetCommandRegistry()
	defer resetCommandRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty Key")
		}
	}()
	RegisterCommand(Command{
		Title: "no key",
		Run:   func(CommandContext) CommandResult { return CommandResult{} },
	})
}

func TestRegisterCommandNilRunPanics(t *testing.T) {
	resetCommandRegistry()
	defer resetCommandRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil Run")
		}
	}()
	RegisterCommand(Command{Key: "no-run"})
}

func TestRegisterCommandDuplicatePanics(t *testing.T) {
	resetCommandRegistry()
	defer resetCommandRegistry()

	RegisterCommand(validCommand("dup-cmd"))
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate Key")
		}
	}()
	RegisterCommand(validCommand("dup-cmd"))
}

// jsonKeys marshals v and returns the set of top-level JSON object keys.
func jsonKeys(t *testing.T, v any) map[string]bool {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	keys := make(map[string]bool, len(m))
	for k := range m {
		keys[k] = true
	}
	return keys
}

// TestCommandResultJSONAlignsWithWasmResponse guards the JSON shape of
// CommandResult against internal/wasm/contract.go's Response. The frontend
// package must not import internal/wasm (zero business dependencies), so the
// wasm side is pinned here as a literal mirror of wasm.Response. The mapping
// itself lives in internal/wasm/sink.go (callWasm) — it relies on this shape
// parity, so the guard stays as a compile-time-free wire-format check.
func TestCommandResultJSONAlignsWithWasmResponse(t *testing.T) {
	// Mirror of internal/wasm/contract.go Response (ProtocolVersion 1).
	// wasm.Response has no Data field, so we only require CommandResult to be
	// a superset of the wasm keys (common-field comparison).
	wasmResponse := struct {
		Action  string   `json:"action"`
		Title   string   `json:"title,omitempty"`
		Message string   `json:"message,omitempty"`
		Level   string   `json:"level,omitempty"`
		URL     string   `json:"url,omitempty"`
		Command string   `json:"command,omitempty"`
		Args    []string `json:"args,omitempty"`
	}{
		Action: "toast", Title: "x", Message: "y", Level: "info",
		URL: "u", Command: "c", Args: []string{"a"},
	}

	wasmKeys := jsonKeys(t, wasmResponse)
	result := CommandResult{
		Action: "toast", Title: "x", Message: "y", Level: "info",
		URL: "u", Command: "c", Args: []string{"a"},
		Data: map[string]any{"k": "v"},
	}
	resultKeys := jsonKeys(t, result)

	for k := range wasmKeys {
		if !resultKeys[k] {
			t.Errorf("CommandResult JSON missing wasm.Response key %q", k)
		}
	}
}
