package headless

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
)

// startDaemonForTest starts a DaemonPlugin-backed server on a temp socket and
// returns the server. The plugin scope and server are cleaned up on test end
// (the scope Dispose unsubscribes the shared emitter, the server close drops
// the subscription connections).
func startDaemonForTest(t *testing.T, engine *core.Engine, sock string) *Server {
	t.Helper()
	server := NewServerWithAddr(engine, "unix", sock)
	scope := framework.NewScope()
	if err := scope.Add(NewDaemonPlugin(server)); err != nil {
		t.Fatalf("scope.Add(daemon): %v", err)
	}
	if err := scope.Start(engine.Ctx()); err != nil {
		t.Fatalf("scope.Start: %v", err)
	}
	t.Cleanup(func() { _ = scope.Dispose() })
	t.Cleanup(server.Close)
	waitForSocket(t, sock)
	return server
}

// subscribeClient dials the daemon and sends a subscribe request, returning
// the open connection and a buffered reader positioned at the first frame
// (the snapshot).
func subscribeClient(t *testing.T, sock string, events []string) (net.Conn, *bufio.Reader) {
	t.Helper()
	conn, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	req := map[string]any{
		"v":    core.ProtocolVersion,
		"id":   1,
		"cmd":  "subscribe",
		"args": map[string]any{"events": events},
	}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		t.Fatalf("encode subscribe: %v", err)
	}
	return conn, bufio.NewReader(conn)
}

// sendSubCmd writes a request on an open subscription connection.
func sendSubCmd(t *testing.T, conn net.Conn, id int64, cmd string, args map[string]any) {
	t.Helper()
	req := map[string]any{"v": core.ProtocolVersion, "id": id, "cmd": cmd, "args": args}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		t.Fatalf("encode %s: %v", cmd, err)
	}
}

// readSubFrame reads the next newline-delimited JSON frame from a subscription
// stream (the headless wire protocol frames responses and events as
// newline-delimited JSON).
func readSubFrame(t *testing.T, r *bufio.Reader) map[string]any {
	t.Helper()
	line, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}
	var frame map[string]any
	if err := json.Unmarshal(line, &frame); err != nil {
		t.Fatalf("unmarshal frame %q: %v", line, err)
	}
	return frame
}

// expectNoSubFrame asserts no frame arrives within d on a subscription stream
// (used to verify subscription filtering and unsubscribe). It sets a read
// deadline for the wait and clears it afterwards.
func expectNoSubFrame(t *testing.T, conn net.Conn, r *bufio.Reader, d time.Duration, what string) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(d))
	if line, err := r.ReadBytes('\n'); err == nil {
		t.Fatalf("%s: expected no frame, got %q", what, line)
	}
	_ = conn.SetReadDeadline(time.Time{})
}

// waitForConnReaped polls until the server has removed the connection with the
// given id from its live set (the read loop exits after the client closed).
func waitForConnReaped(t *testing.T, s *Server, id int64) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		s.connsMu.Lock()
		_, ok := s.conns[id]
		s.connsMu.Unlock()
		if !ok {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("connection %d was not reaped after client close", id)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// emitDaemonEvent emits an event on the engine's core bus the way the player
// does (the payload already carries the frame data shape).
func emitDaemonEvent(t *testing.T, s *Server, name string, payload any) {
	t.Helper()
	emitter, ok := framework.ServiceOf[*framework.EventEmitter](s.engine.Ctx(), core.ServiceEventBus)
	if !ok {
		t.Fatal("eventBus not resolved from engine ctx")
	}
	if err := emitter.Emit(s.engine.Ctx(), name, payload); err != nil {
		t.Fatalf("emit %s: %v", name, err)
	}
}

// TestDaemonSubscriptionSession exercises the P7 dual-capability control
// channel: the legacy musicfox ctrl one-shot path stays intact, a subscribe
// connection receives the snapshot first and then a subscription-filtered
// event stream, unsubscribe stops delivery, and a disconnected subscriber is
// reaped without breaking the broadcast path.
func TestDaemonSubscriptionSession(t *testing.T) {
	engine := newTestEngine(t)
	sock := shortSockPath(t)
	server := startDaemonForTest(t, engine, sock)

	// Legacy musicfox ctrl compatibility on the same daemon: one-shot request →
	// one-shot response, connection closed after (CtrlClient zero changes).
	client := testCtrlClient(t, sock)
	resp, err := client.Call(context.Background(), "status", nil)
	if err != nil {
		t.Fatalf("legacy Call(status) error = %v", err)
	}
	if !resp.Ok {
		t.Fatalf("legacy Call(status) ok = false, error = %q", resp.Error)
	}

	// Subscription session: subscribe to song_changed only.
	conn, r := subscribeClient(t, sock, []string{core.EvSongChanged})

	// Snapshot must be the first frame, carrying the status fields and the
	// trimmed playlist.
	snap := readSubFrame(t, r)
	if snap["type"] != "snapshot" {
		t.Fatalf("first frame type = %q, want snapshot (frame=%v)", snap["type"], snap)
	}
	snapData, _ := snap["data"].(map[string]any)
	for _, key := range []string{"state", "playlist"} {
		if _, ok := snapData[key]; !ok {
			t.Fatalf("snapshot missing key %q (data=%v)", key, snapData)
		}
	}

	// Then the subscribe ack.
	ack := readSubFrame(t, r)
	if ack["ok"] != true {
		t.Fatalf("subscribe ack ok = %v, error = %v", ack["ok"], ack["error"])
	}

	// A subscribed event reaches the connection with the wire name echoed.
	emitDaemonEvent(t, server, core.EvSongChanged, map[string]any{"id": 42, "name": "订阅歌曲"})
	ev := readSubFrame(t, r)
	if ev["type"] != "event" {
		t.Fatalf("frame type = %q, want event (frame=%v)", ev["type"], ev)
	}
	if ev["event"] != core.EvSongChanged {
		t.Fatalf("event name = %q, want %q", ev["event"], core.EvSongChanged)
	}

	// An unsubscribed event is filtered out by the subscription set.
	emitDaemonEvent(t, server, core.EvStateChanged, map[string]any{"state": "paused"})
	expectNoSubFrame(t, conn, r, 200*time.Millisecond, "state_changed (unsubscribed)")

	// unsubscribe removes the subscription.
	sendSubCmd(t, conn, 2, "unsubscribe", map[string]any{"events": []string{core.EvSongChanged}})
	uack := readSubFrame(t, r)
	if uack["ok"] != true {
		t.Fatalf("unsubscribe ack ok = %v, error = %v", uack["ok"], uack["error"])
	}

	emitDaemonEvent(t, server, core.EvSongChanged, map[string]any{"id": 43, "name": "不应送达"})
	expectNoSubFrame(t, conn, r, 200*time.Millisecond, "song_changed after unsubscribe")

	// A second subscriber to a different event; after it disconnects the frame
	// targeting it must be dropped (never delivered to conn1, never panicking
	// the broadcast path).
	conn2, r2 := subscribeClient(t, sock, []string{core.EvStateChanged})
	_ = readSubFrame(t, r2) // snapshot
	_ = readSubFrame(t, r2) // ack
	_ = conn2.Close()
	waitForConnReaped(t, server, 2)

	emitDaemonEvent(t, server, core.EvStateChanged, map[string]any{"state": "playing"})
	expectNoSubFrame(t, conn, r, 200*time.Millisecond, "state_changed (targets reaped conn)")

	// The long-lived connection is still usable after all of the above: a
	// request/response round-trip works alongside the event stream.
	sendSubCmd(t, conn, 3, "status", nil)
	statusAck := readSubFrame(t, r)
	if statusAck["ok"] != true {
		t.Fatalf("status on subscription connection ok = %v, error = %v", statusAck["ok"], statusAck["error"])
	}
}

// TestDaemonSubscribeRequiresEvents verifies the subscribe validation path: a
// subscribe request without a non-empty events array is answered with an error
// ack (unknown/invalid commands never reach the Dispatcher).
func TestDaemonSubscribeRequiresEvents(t *testing.T) {
	engine := newTestEngine(t)
	sock := shortSockPath(t)
	startDaemonForTest(t, engine, sock)

	conn, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()

	req := map[string]any{
		"v":    core.ProtocolVersion,
		"id":   1,
		"cmd":  "subscribe",
		"args": map[string]any{},
	}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		t.Fatalf("encode subscribe: %v", err)
	}

	var resp core.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatalf("decode subscribe error ack: %v", err)
	}
	if resp.Ok {
		t.Fatal("subscribe with empty events should fail")
	}
	if resp.Error == "" {
		t.Fatal("subscribe error ack missing error message")
	}
}
