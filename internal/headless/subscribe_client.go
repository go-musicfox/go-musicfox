package headless

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
)

// subscribeEventQueueSize bounds the queue between the read loop and the
// Events channel consumers. When full, the newest frame is dropped — the
// slow-subscriber philosophy from the daemon/WebUI side (a slow consumer only
// loses its own frames, never the connection's read loop).
const subscribeEventQueueSize = 64

// SubscribeClient connects to a running headless daemon's control channel and
// holds a subscription session (P7, docs/plugin_ecosystem.md §3.4): the first
// request is "subscribe" (snapshot frame first, then an ack), after which
// event frames interleave with request/response command handling on the same
// connection — the daemon's per-connection write lock guarantees frames never
// interleave. It is the long-lived, event-streaming upgrade of CtrlClient
// (which is kept unchanged for musicfox ctrl's one-shot calls).
//
// The client is consumed by the WebUI connect mode (internal/webui): control
// commands go through Call, status goes through the cached Snapshot, and the
// event stream is consumed from Events.
type SubscribeClient struct {
	network string
	addr    string

	// mu serializes writes (subscribe/Call) so JSON frames never interleave on
	// the wire.
	mu   sync.Mutex
	conn net.Conn
	enc  *json.Encoder
	dec  *json.Decoder

	idSeq atomic.Int64

	// events delivers side frames (the snapshot frame once, then event
	// frames). It is closed by the read loop when it exits (Close or a daemon
	// disconnect), so no sender can ever race a closed channel.
	events chan []byte

	// snapshot caches the most recent snapshot frame's data (daemon status
	// fields + trimmed playlist).
	snapshotMu sync.RWMutex
	snapshot   map[string]any

	// pending correlates in-flight requests by ID: Call registers a channel
	// and the read loop delivers the matching Response.
	pendingMu sync.Mutex
	pending   map[int64]chan *core.Response

	closed    chan struct{} // closed by shutdown: wakes pending callers + Close waits on readDone
	closeOnce sync.Once
	readDone  chan struct{} // closed when the read loop exits (Close blocks on it)
}

// DialSubscribe resolves the daemon control address (DialAddr), probes it for a
// live daemon and sends the "subscribe" request with events (wire names, see
// the core.Ev* constants). It blocks until the daemon acknowledges the
// subscription, so on success the snapshot is cached and the event stream is
// live. A missing or unresponsive daemon yields the same "headless daemon is
// not running" error as CtrlClient.Dial.
func DialSubscribe(events []string) (*SubscribeClient, error) {
	network, addr := DialAddr()
	if network == "" || addr == "" {
		return nil, errors.New("headless daemon is not running")
	}
	return dialSubscribeAddr(network, addr, events)
}

// dialSubscribeAddr is DialSubscribe over an explicit address (used by the
// tests to point at a temp socket).
func dialSubscribeAddr(network, addr string, events []string) (*SubscribeClient, error) {
	conn, err := net.DialTimeout(network, addr, time.Second)
	if err != nil {
		return nil, errors.New("headless daemon is not running")
	}

	c := &SubscribeClient{
		network:  network,
		addr:     addr,
		conn:     conn,
		enc:      json.NewEncoder(conn),
		dec:      json.NewDecoder(conn),
		events:   make(chan []byte, subscribeEventQueueSize),
		pending:  make(map[int64]chan *core.Response),
		closed:   make(chan struct{}),
		readDone: make(chan struct{}),
	}
	go c.readLoop()

	// Register the subscribe request's own ID so the read loop correlates the
	// ack, then wait for it (bounded by the same 3s budget as CtrlClient.Call).
	ackCh := make(chan *core.Response, 1)
	reqID := c.idSeq.Add(1)
	c.pendingMu.Lock()
	c.pending[reqID] = ackCh
	c.pendingMu.Unlock()

	if err := c.writeRequest(core.Request{
		V:    core.ProtocolVersion,
		ID:   reqID,
		Cmd:  "subscribe",
		Args: map[string]any{"events": events},
	}); err != nil {
		c.shutdown()
		return nil, err
	}

	select {
	case resp, ok := <-ackCh:
		if !ok || resp == nil {
			// Connection died before the ack arrived (shutdown closed the
			// channel): treat as no daemon.
			c.Close()
			return nil, errors.New("headless daemon is not running")
		}
		if !resp.Ok {
			c.Close()
			return nil, errors.New(resp.Error)
		}
	case <-c.closed:
		c.Close()
		return nil, errors.New("headless daemon is not running")
	case <-time.After(3 * time.Second):
		c.Close()
		return nil, errors.New("headless daemon subscribe timeout")
	}
	return c, nil
}

// Call executes a control command on the subscription connection and waits for
// the ID-correlated response, bounded by a 3s deadline (or an earlier ctx
// deadline — mirroring CtrlClient). "quit" is rejected here: the connect-mode
// transport-layer semantics deliberately never forward it to the daemon
// (D-S5-2).
func (c *SubscribeClient) Call(ctx context.Context, cmd string, args map[string]any) (*core.Response, error) {
	if cmd == "quit" {
		return nil, errors.New("quit is not forwarded by the subscribe client")
	}

	ch := make(chan *core.Response, 1)
	id := c.idSeq.Add(1)

	c.pendingMu.Lock()
	if c.Closed() {
		c.pendingMu.Unlock()
		return nil, errors.New("headless daemon is not running")
	}
	c.pending[id] = ch
	c.pendingMu.Unlock()

	if err := c.writeRequest(core.Request{V: core.ProtocolVersion, ID: id, Cmd: cmd, Args: args}); err != nil {
		c.dropPending(id)
		return nil, err
	}

	deadline := time.Now().Add(3 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case resp, ok := <-ch:
		if !ok {
			// The connection died under us: shutdown closed the channel.
			return nil, errors.New("headless daemon is not running")
		}
		return resp, nil
	case <-timer.C:
		c.dropPending(id)
		return nil, errors.New("headless daemon call timeout")
	case <-ctx.Done():
		c.dropPending(id)
		return nil, ctx.Err()
	}
}

// Snapshot returns a copy of the most recent snapshot data (daemon status +
// trimmed playlist), or nil before the snapshot frame arrives (e.g. during the
// handshake).
func (c *SubscribeClient) Snapshot() map[string]any {
	c.snapshotMu.RLock()
	defer c.snapshotMu.RUnlock()
	if c.snapshot == nil {
		return nil
	}
	out := make(map[string]any, len(c.snapshot))
	for k, v := range c.snapshot {
		out[k] = v
	}
	return out
}

// Events returns the side-frame delivery channel. The snapshot frame is
// delivered once, followed by event frames; the channel closes when the client
// is closed or the daemon disconnects.
func (c *SubscribeClient) Events() <-chan []byte {
	return c.events
}

// Closed reports whether the client has been torn down (Close or a daemon
// disconnect). Backends use it to degrade auxiliary endpoints gracefully.
func (c *SubscribeClient) Closed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

// Close tears the subscription down: it wakes every pending Caller (they
// return an error), closes the Events channel and releases the connection. It
// is idempotent and safe for concurrent use.
func (c *SubscribeClient) Close() error {
	c.shutdown()
	<-c.readDone
	return nil
}

// shutdown tears down the connection and wakes every pending caller. It is
// called by Close (external teardown) and by the read loop (daemon disconnect);
// it never closes the Events channel — the read loop owns that (deferred
// close), so no sender can race a closed channel.
func (c *SubscribeClient) shutdown() {
	c.closeOnce.Do(func() {
		close(c.closed)
		_ = c.conn.Close()
		c.pendingMu.Lock()
		for id, ch := range c.pending {
			delete(c.pending, id)
			close(ch)
		}
		c.pendingMu.Unlock()
	})
}

// writeRequest marshals req and writes it as a newline-delimited frame under
// the write lock (so it can never interleave with a concurrent Call write).
func (c *SubscribeClient) writeRequest(req core.Request) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.enc.Encode(req)
}

// dropPending removes a timed-out/abandoned request from the correlation map.
// A response arriving for it afterwards is silently discarded (the caller no
// longer waits).
func (c *SubscribeClient) dropPending(id int64) {
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()
	delete(c.pending, id)
}

// readLoop decodes frames from the connection until it breaks (Close or a
// daemon disconnect), then wakes pending callers and closes the Events
// channel.
func (c *SubscribeClient) readLoop() {
	defer close(c.events)
	defer close(c.readDone)
	for {
		var raw json.RawMessage
		if err := c.dec.Decode(&raw); err != nil {
			// Connection closed (Close) or daemon gone: wake pending callers
			// and drop the subscription. Events consumers observe the closed
			// channel.
			c.shutdown()
			return
		}
		c.handleFrame(raw)
	}
}

// handleFrame dispatches one decoded frame: side frames (snapshot/event, they
// carry a "type" field) go to handleSideFrame; Response frames (they carry
// "id"/"v"/"ok") are correlated with a pending Call by ID.
func (c *SubscribeClient) handleFrame(raw []byte) {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return
	}
	if probe.Type != "" {
		c.handleSideFrame(raw)
		return
	}

	var resp core.Response
	if err := json.Unmarshal(raw, &resp); err != nil {
		return
	}
	c.pendingMu.Lock()
	ch, ok := c.pending[resp.ID]
	if ok {
		delete(c.pending, resp.ID)
		select {
		case ch <- &resp:
		default:
		}
	}
	c.pendingMu.Unlock()
}

// handleSideFrame handles a snapshot or event frame. The snapshot's data is
// cached (Snapshot reads it) and the raw frame is delivered to the Events
// channel; the frame is dropped when the channel is full (the consumer is
// slower than the daemon's stream — it only loses its own frames).
func (c *SubscribeClient) handleSideFrame(raw []byte) {
	var frame struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &frame); err != nil {
		return
	}
	if frame.Type == "snapshot" {
		var data map[string]any
		if err := json.Unmarshal(frame.Data, &data); err != nil {
			return
		}
		c.snapshotMu.Lock()
		c.snapshot = data
		c.snapshotMu.Unlock()
	}
	select {
	case c.events <- raw:
	default:
		slog.Debug("headless subscribe: event channel full, dropping frame", slog.String("type", frame.Type))
	}
}
