// P7 generic event sink (docs/plugin_ecosystem.md §3.4): the connection-set +
// write-lock + non-blocking broadcast mechanism extracted from the WebUI
// broadcaster so the headless daemon and the WebUI share one implementation.
//
// This file keeps the frontend package's zero-business-dependency invariant:
// standard library only, no internal/* imports and no third-party business
// libraries.
package frontend

import (
	"encoding/json"
	"log/slog"
	"sync"
)

// EventSinkConn is a connection registered with an EventSink. Write must be
// safe for concurrent calls (implementations serialize with their own write
// lock and own the transport framing, e.g. a WebSocket message or a newline
// terminator): the sink invokes it from per-connection goroutines. A failed
// write only drops that connection's frame.
type EventSinkConn interface {
	Write(payload []byte) error
}

// EventSink fans event payloads out to every registered connection. The
// connection set is snapshotted under the lock and each connection is written
// by its own goroutine (serialized per connection by the conn's own lock), so
// a slow or broken connection only loses its own frames and never stalls the
// others or the event emitter callback.
type EventSink struct {
	mu    sync.Mutex
	conns map[int64]EventSinkConn
}

// NewEventSink creates an empty event sink.
func NewEventSink() *EventSink {
	return &EventSink{conns: make(map[int64]EventSinkConn)}
}

// Add registers a connection for event delivery.
func (s *EventSink) Add(id int64, c EventSinkConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[id] = c
}

// Remove drops a connection from the delivery set (idempotent).
func (s *EventSink) Remove(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, id)
}

// Broadcast delivers payload to every registered connection.
func (s *EventSink) Broadcast(payload []byte) {
	s.BroadcastFiltered(payload, func(int64) bool { return true })
}

// BroadcastFiltered delivers payload to the registered connections for which
// keep(id) returns true. keep is evaluated under the connection-set lock, so
// it must be cheap and must not call back into the sink (callers that filter
// by their own state should snapshot it first, as the headless daemon's
// subscription set does).
func (s *EventSink) BroadcastFiltered(payload []byte, keep func(id int64) bool) {
	s.mu.Lock()
	conns := make([]EventSinkConn, 0, len(s.conns))
	for id, c := range s.conns {
		if keep(id) {
			conns = append(conns, c)
		}
	}
	s.mu.Unlock()

	for _, c := range conns {
		c := c
		go func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("frontend eventsink: connection write panic", slog.Any("panic", r))
				}
			}()
			if err := c.Write(payload); err != nil {
				// A dead/slow connection only loses its own frame.
				slog.Debug("frontend eventsink: write failed", slog.Any("err", err))
			}
		}()
	}
}

// Reset drops every registered connection. Close calls it after terminating
// the connections so no further broadcast targets a dead socket.
func (s *EventSink) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns = make(map[int64]EventSinkConn)
}

// EventFrame marshals an event frame:
// {"type":"event","event":"<name>","data":{...}} — the shared wire shape for
// both the WebUI and the headless daemon event streams. It returns nil when
// data cannot be marshaled (unreachable for the JSON-primitive payloads the
// core event bus emits).
func EventFrame(name string, data any) []byte {
	payload, err := json.Marshal(map[string]any{"type": "event", "event": name, "data": data})
	if err != nil {
		return nil
	}
	return payload
}
