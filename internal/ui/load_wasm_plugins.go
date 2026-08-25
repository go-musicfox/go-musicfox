package ui

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// loadWasmPlugins discovers, loads and registers WASM plugins at startup. The
// registration is attributed to each plugin via WithPlugin so the [plugins]
// disabled config works for WASM plugins by manifest id.
func (n *Netease) loadWasmPlugins() {
	if n.wasmManager == nil {
		return
	}
	dir := configs.AppConfig.Plugins.WasmDir
	if dir == "" {
		dir = filepath.Join(app.ConfigDir(), "wasm-plugins")
	}
	errs := n.wasmManager.LoadDir(context.Background(), dir)
	for _, err := range errs {
		slog.Error("wasm plugin load failed", slogx.Error(err))
	}
	for _, p := range n.wasmManager.Plugins() {
		registerWasmPlugin(p)
	}
}

// registerWasmPlugin registers every manifest-declared menu of a WASM plugin as
// a WasmPluginMenu provider and main-menu entry, attributed to the plugin id via
// WithPlugin so the [plugins] disabled config applies. Recover-isolated: one bad
// plugin (duplicate menu key, bad after anchor, ...) must not crash startup.
func registerWasmPlugin(p *wasm.Plugin) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("wasm plugin registration panicked", slogx.Error(r))
		}
	}()
	WithPlugin(p.ID, p.Name, func() {
		for _, decl := range p.Menus {
			RegisterMenu(decl.Key, func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
				return &WasmPluginMenu{BaseMenu: base, plugin: p, decl: decl}, nil
			})
			if decl.After == "" {
				RegisterMainMenuItem(decl.Key, decl.Title)
			} else {
				RegisterMainMenuItemAfter(decl.Key, decl.Title, decl.After, nil)
			}
		}
	})
}
