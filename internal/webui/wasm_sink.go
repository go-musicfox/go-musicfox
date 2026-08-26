package webui

import (
	"fmt"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
)

// webuiWasmSink registers WASM plugin commands into the frontend registry.
// Unlike the TUI sink it needs no WithPlugin scope: CommandsOf has already
// stamped PluginID on each command, so the [plugins] disabled config gates
// them directly, and WebUI does not depend on the ui package. Panics
// (duplicate command key, ...) are recovered and surfaced as an error so one
// bad plugin does not crash startup.
type webuiWasmSink struct{}

func (webuiWasmSink) RegisterCommands(p *wasm.Plugin, cmds []frontend.Command) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("wasm: register commands for plugin %q panicked: %v", p.ID, r)
		}
	}()
	for _, cmd := range cmds {
		frontend.RegisterCommand(cmd)
	}
	return nil
}
