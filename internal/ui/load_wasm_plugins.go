package ui

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// tuiWasmSink registers WASM plugin commands into the TUI: every command is
// registered inside a WithPlugin(p.ID, p.Name, ...) scope so ui.RegisterCommand
// stamps the manifest plugin id and [plugins] disabled can gate the command.
// Panics (duplicate command key, ...) are recovered and surfaced as an error so
// one bad plugin does not crash startup.
type tuiWasmSink struct{}

func (tuiWasmSink) RegisterCommands(p *wasm.Plugin, cmds []frontend.Command) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("wasm: register commands for plugin %q panicked: %v", p.ID, r)
		}
	}()
	WithPlugin(p.ID, p.Name, func() {
		for _, cmd := range cmds {
			RegisterCommand(cmd)
		}
	})
	return nil
}

// loadWasmPlugins discovers, loads and registers WASM plugins at startup. The
// registration is attributed to each plugin via tuiWasmSink / WithPlugin so the
// [plugins] disabled config works for WASM plugins by manifest id. The Manager
// is kept on the shell so CloseHook can release the instances. Recover-isolated:
// an unexpected panic (loader bug, ...) is logged, never fatal.
func (n *Netease) loadWasmPlugins(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("wasm plugin load panicked", slogx.Error(r))
		}
	}()
	if n.wasmManager != nil {
		return
	}
	dir := configs.AppConfig.Plugins.WasmDir
	if dir == "" {
		dir = filepath.Join(app.ConfigDir(), "wasm-plugins")
	}
	mgr, errs := wasm.LoadAndRegister(ctx, dir, tuiWasmSink{})
	if mgr == nil {
		return
	}
	n.wasmManager = mgr
	for _, err := range errs {
		slog.Error("wasm plugin load failed", slogx.Error(err))
	}
}
