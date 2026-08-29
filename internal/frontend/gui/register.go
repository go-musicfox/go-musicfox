// Package gui provides the Wails v2 native window frontend (ID "wails").
//
// The window loads the WebUI page (reusing webui.Server + the Backend
// abstraction from S5-2); non-Bind mode: the page talks to the backend over
// plain HTTP/WS and never depends on the Wails runtime/IPC (wails issue
// #4686: pages loaded from an external URL cannot reach the Wails runtime).
//
// GUI-1 spike conclusion (see spike.md): Wails v2.15 is compileable on the
// current toolchain, GUI-B (navigating the window to an external
// http://127.0.0.1:<port> URL) is feasible via a minimal AssetServer.Handler
// + runtime.WindowExecJS from OnDomReady. The wails dependency is kept in
// go.mod by wails_spike.go (build tag) until GUI-2 wires it into Run.
package gui

import (
	"context"
	"errors"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// guiFrontend is the Wails v2 native window frontend (Frontend interface
// implementation). The window loads the WebUI page (reusing webui.Server +
// Backend abstraction); non-Bind mode: the page communicates over HTTP/WS
// and does not depend on the Wails runtime.
type guiFrontend struct{}

func (guiFrontend) ID() string   { return "wails" }
func (guiFrontend) Name() string { return "GUI" }

func (guiFrontend) Run(ctx context.Context, opts frontend.LaunchOptions) error {
	// GUI-1 placeholder: the spike conclusions are locked in spike.md; GUI-2
	// implements the full window integration (engine + server + Wails window).
	return errors.New("wails frontend not wired yet")
}

func init() {
	frontend.Register(guiFrontend{})
}
