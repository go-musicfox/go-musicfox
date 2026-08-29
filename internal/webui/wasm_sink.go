package webui

import (
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
)

// webuiWasmSink registers WASM plugin commands into the frontend registry.
// Commands are REPLACED (frontend.ReplaceCommand) rather than registered: a
// reloaded plugin re-registers its keys without a duplicate-key panic and keeps
// its registration-order position. Unlike the TUI sink it needs no WithPlugin
// scope: CommandsOf has already stamped PluginID, so the [plugins] disabled
// config gates them directly, and WebUI does not depend on the ui package.
type webuiWasmSink struct{}

func (webuiWasmSink) RegisterCommands(p *wasm.Plugin, cmds []frontend.Command) error {
	for _, cmd := range cmds {
		frontend.ReplaceCommand(cmd)
	}
	return nil
}
