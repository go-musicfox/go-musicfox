package ui

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// tuiWasmSink registers WASM plugin commands into the TUI: every command is
// registered inside a WithPlugin(p.ID, p.Name, ...) scope so the manifest
// plugin id is attributed (CommandKeys recording) and [plugins] disabled can
// gate the command (CommandsOf already stamps PluginID, so the stamp is
// idempotent). Commands are REPLACED (frontend.ReplaceCommand) rather than
// registered: a reloaded plugin re-registers its keys without a duplicate-key
// panic and keeps its registration-order position. The WithPlugin scope records
// the key under the plugin's CommandKeys (the mirror of ui.RegisterCommand).
type tuiWasmSink struct{}

func (tuiWasmSink) RegisterCommands(p *wasm.Plugin, cmds []frontend.Command) error {
	WithPlugin(p.ID, p.Name, func() {
		for _, cmd := range cmds {
			frontend.ReplaceCommand(cmd)
			recordPluginCommandKey(cmd.Key)
		}
	})
	return nil
}

// loadWasmPlugins discovers, loads and mounts WASM plugins into the wasm
// sub-scope (P6). NewFrontendScope registers the sub-scope's ManagerPlugin and
// the frontend scope Start provides the manager into the app context; this
// mounts one wasmPlugin adapter per loaded plugin directory via
// wasm.LoadIntoScope, which AddAndStarts them on the already-started child
// scope. The wasm sub-scope is a child of the frontend scope, so the manager
// and plugin instances are closed by the frontend scope's Dispose in CloseHook
// — no separate manager cleanup is kept on the shell. Recover-isolated: an
// unexpected panic (loader bug, ...) is logged, never fatal.
func (n *Netease) loadWasmPlugins(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("wasm plugin load panicked", slogx.Error(r))
		}
	}()
	if n.wasmScope == nil || n.ctx == nil {
		return
	}
	dir := configs.AppConfig.Plugins.WasmDir
	if dir == "" {
		dir = filepath.Join(app.ConfigDir(), "wasm-plugins")
	}
	_, errs := wasm.LoadIntoScope(ctx, n.ctx, n.wasmScope, dir, tuiWasmSink{})
	for _, err := range errs {
		slog.Error("wasm plugin load failed", slogx.Error(err))
	}
}
