package webui

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/go-musicfox/go-musicfox/utils/errorx"
)

// broadcastWriteTimeout bounds a single per-connection event write. A stuck
// client must never pin its write goroutine forever; the frame is dropped for
// that connection alone and the engine callback is never delayed.
const broadcastWriteTimeout = 5 * time.Second

// broadcaster fans engine events out to every live WebSocket connection. The
// connection set is snapshotted under the lock and each connection is written
// by its own goroutine (serialized per connection via writeMu), so a slow or
// broken connection only loses its own frames and never stalls the others or
// the engine observer callback.
type broadcaster struct {
	mu    sync.Mutex
	conns map[int64]*wsConn
}

// add registers a connection for event delivery.
func (b *broadcaster) add(c *wsConn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.conns[c.id] = c
}

// remove drops a connection from the delivery set (idempotent).
func (b *broadcaster) remove(id int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.conns, id)
}

// broadcast sends payload to every connected client without ever blocking the
// caller: the conn set is snapshotted under the lock, then each write runs in
// its own goroutine.
func (b *broadcaster) broadcast(payload []byte) {
	b.mu.Lock()
	conns := make([]*wsConn, 0, len(b.conns))
	for _, c := range b.conns {
		conns = append(conns, c)
	}
	b.mu.Unlock()

	for _, c := range conns {
		c := c
		errorx.Go(func() {
			c.writeMu.Lock()
			defer c.writeMu.Unlock()
			ctx, cancel := context.WithTimeout(context.Background(), broadcastWriteTimeout)
			defer cancel()
			if err := c.c.Write(ctx, websocket.MessageText, payload); err != nil {
				// A dead/slow connection only loses its own frame.
				slog.Debug("webui broadcast: write failed", slog.Any("err", err))
			}
		}, true)
	}
}

// reset drops every registered connection. Close calls it after terminating
// the connections so no further broadcast targets a dead socket.
func (b *broadcaster) reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.conns = make(map[int64]*wsConn)
}
