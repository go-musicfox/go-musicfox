package webui

import (
	"context"
	"fmt"

	"github.com/go-musicfox/go-musicfox/internal/headless"
)

// connectRun runs the WebUI in connect mode: it dials the local headless
// daemon (musicfox --headless), builds a remoteBackend over the subscription
// connection and serves the regular WebUI HTTP/WS surface against it. No
// engine is built and no WASM scope is loaded — playback, state and events all
// come from the daemon.
//
// Functional boundary (D-S5-2, docs/roadmap_s_series.md §2.3 S5-4):
//
//	capability       | standalone           | connect
//	-----------------|----------------------|------------------------------------
//	playback control | local Dispatcher     | forwarded to the daemon (Call) ✅
//	state/snapshot   | local engine         | daemon status + playlist snapshot ✅
//	events           | local event bus      | daemon subscription, wire→frame map ✅
//	/api/commands    | local command list   | empty (daemon registers no commands) |
//	login (QR)       | local                | 503 (no local engine)               |
//	/api/albumart    | local engine         | 404 (snapshot carries no PicUrl)    |
//	/api/lyrics      | local engine         | empty structure                     |
//	WS quit          | shuts itself down    | not forwarded (never reaches daemon)|
//
// These boundaries fall out of the Backend implementations (localBackend vs
// remoteBackend) — the Server itself is mode-agnostic.
func connectRun(ctx context.Context) error {
	client, err := headless.DialSubscribe(eventWireNames())
	if err != nil {
		return fmt.Errorf("connect 模式需要 headless daemon 正在运行: %w", err)
	}
	defer client.Close()

	backend := newRemoteBackend(client)
	server := NewServerWithOptions(backend, ServerOptions{Auth: true})
	return runServer(ctx, server)
}

// eventWireNames returns the core event-bus wire names the WebUI consumes —
// the subscribe set for connect mode (the keys of eventWireToFrame).
func eventWireNames() []string {
	names := make([]string, 0, len(eventWireToFrame))
	for wire := range eventWireToFrame {
		names = append(names, wire)
	}
	return names
}
