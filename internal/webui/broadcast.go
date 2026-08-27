package webui

import (
	"context"
	"time"

	"github.com/coder/websocket"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// broadcastWriteTimeout bounds a single per-connection event write. A stuck
// client must never pin its write goroutine forever; the frame is dropped for
// that connection alone and the engine callback is never delayed.
const broadcastWriteTimeout = 5 * time.Second

// Write implements frontend.EventSinkConn: it serializes one event-frame write
// under the connection's write lock (responses and broadcasts can never
// interleave) and bounds it by broadcastWriteTimeout. Transport framing is the
// WebSocket message itself.
func (c *wsConn) Write(payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), broadcastWriteTimeout)
	defer cancel()
	return c.c.Write(ctx, websocket.MessageText, payload)
}

// broadcaster fans engine events out to every live WebSocket connection. It is
// the WebUI-typed wrapper over the shared frontend.EventSink (P7 extraction,
// docs/plugin_ecosystem.md §3.4): the generic sink owns the connection set and
// the non-blocking per-connection goroutine writes, so a slow or broken
// connection only loses its own frames and never stalls the others or the
// engine observer callback.
type broadcaster struct {
	sink *frontend.EventSink
}

// newBroadcaster creates an empty broadcaster.
func newBroadcaster() *broadcaster {
	return &broadcaster{sink: frontend.NewEventSink()}
}

// add registers a connection for event delivery.
func (b *broadcaster) add(c *wsConn) { b.sink.Add(c.id, c) }

// remove drops a connection from the delivery set (idempotent).
func (b *broadcaster) remove(id int64) { b.sink.Remove(id) }

// broadcast sends payload to every connected client without ever blocking the
// caller: the conn set is snapshotted under the lock, then each write runs in
// its own goroutine.
func (b *broadcaster) broadcast(payload []byte) { b.sink.Broadcast(payload) }

// reset drops every registered connection. Close calls it after terminating
// the connections so no further broadcast targets a dead socket.
func (b *broadcaster) reset() { b.sink.Reset() }
