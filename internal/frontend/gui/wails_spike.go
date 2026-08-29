//go:build wails_spike

// This file exists only for the GUI-1 spike (see spike.md):
//  1. It keeps the Wails v2 module in go.mod/go.sum — `go mod tidy` acts as if
//     all build tags were enabled, so a package that imports wails must exist
//     or tidy would drop the requirement.
//  2. It is the compile verification target: `go build -tags wails_spike
//     ./internal/frontend/gui` compiles the Wails v2 module (cgo + platform
//     bindings) on the current toolchain (spike item 1).
//
// It is excluded from normal builds by the wails_spike build tag and must be
// DELETED once gui/run.go (GUI-2) imports wails unconditionally.
package gui

import (
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// spikeVerifyWails holds compile-time references to the Wails v2 public API
// surface used by the GUI-B window path. It is never executed.
//
// Reference (wails v2.15.0):
//   - wails.Run(options *options.App) error — wails.go:10-13
//   - options.App.AssetServer *assetserver.Options — pkg/options/options.go:56
//   - assetserver.Options{Assets, Handler, Middleware} — pkg/options/assetserver/options.go
func spikeVerifyWails() {
	var run func(*options.App) error = wails.Run
	var _ = run
	var _ assetserver.Options
}
