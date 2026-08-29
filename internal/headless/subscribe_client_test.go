package headless

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// TestSubscribeClientSnapshotAndCall verifies the subscribe handshake over a
// real unix socket: the snapshot frame is cached (and delivered once on
// Events), the ack correlates the subscribe request, and control commands are
// answered on the same long-lived connection with per-request ID correlation.
func TestSubscribeClientSnapshotAndCall(t *testing.T) {
	engine := newTestEngine(t)
	sock := shortSockPath(t)
	server := NewServerWithAddr(engine, "unix", sock)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Serve(ctx) }()
	t.Cleanup(server.Close)
	waitForSocket(t, sock)

	client, err := dialSubscribeAddr("unix", sock, []string{core.EvSongChanged, core.EvStateChanged})
	if err != nil {
		t.Fatalf("dialSubscribeAddr error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	// The snapshot arrives before the ack, so it must be cached once
	// dialSubscribeAddr returns.
	snap := client.Snapshot()
	if snap == nil {
		t.Fatal("Snapshot() = nil, want status data")
	}
	for _, key := range []string{"playing", "state", "song", "playlist"} {
		if _, ok := snap[key]; !ok {
			t.Fatalf("snapshot missing key %q (keys: %v)", key, snap)
		}
	}

	// The snapshot frame is delivered once on Events.
	select {
	case first := <-client.Events():
		var f struct {
			Type string         `json:"type"`
			Data map[string]any `json:"data"`
		}
		if err := json.Unmarshal(first, &f); err != nil {
			t.Fatalf("unmarshal first frame: %v", err)
		}
		if f.Type != "snapshot" {
			t.Fatalf("first frame type = %q, want snapshot", f.Type)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("snapshot frame did not arrive within 3s")
	}

	// A control command is answered with its ID correlation.
	resp, err := client.Call(context.Background(), "status", nil)
	if err != nil {
		t.Fatalf("Call(status) error = %v", err)
	}
	if !resp.Ok {
		t.Fatalf("Call(status) ok = false, error = %q", resp.Error)
	}
	if resp.Data == nil {
		t.Fatal("Call(status) data = nil")
	}

	// An unknown command surfaces ok:false with the daemon error.
	resp, err = client.Call(context.Background(), "nonsense", nil)
	if err != nil {
		t.Fatalf("Call(nonsense) error = %v", err)
	}
	if resp.Ok {
		t.Fatal("Call(nonsense) ok = true, want false")
	}
	if resp.Error == "" {
		t.Fatal("Call(nonsense) missing error message")
	}

	// quit is never forwarded by the subscribe client (D-S5-2).
	if _, err := client.Call(context.Background(), "quit", nil); err == nil {
		t.Fatal("Call(quit) succeeded, want rejection")
	}
}

// TestSubscribeClientEvents verifies subscribed event frames arrive on the
// Events channel carrying the core wire name. The frame is fed through the
// server's broadcastEvent (the daemon-side fan-out the DaemonPlugin drives in
// production).
func TestSubscribeClientEvents(t *testing.T) {
	engine := newTestEngine(t)
	sock := shortSockPath(t)
	server := NewServerWithAddr(engine, "unix", sock)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Serve(ctx) }()
	t.Cleanup(server.Close)
	waitForSocket(t, sock)

	client, err := dialSubscribeAddr("unix", sock, []string{core.EvSongChanged})
	if err != nil {
		t.Fatalf("dialSubscribeAddr error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	// Drain the snapshot frame (always delivered first).
	<-client.Events()

	payload := frontend.EventFrame(core.EvSongChanged, map[string]any{
		"id":              42,
		"name":            "测试歌曲",
		"artist":          "歌手A",
		"album":           "专辑X",
		"picUrl":          "https://p1.music.126.net/cover.jpg",
		"durationSeconds": 180.0,
	})
	server.broadcastEvent(core.EvSongChanged, payload)

	select {
	case frame := <-client.Events():
		var f struct {
			Type  string         `json:"type"`
			Event string         `json:"event"`
			Data  map[string]any `json:"data"`
		}
		if err := json.Unmarshal(frame, &f); err != nil {
			t.Fatalf("unmarshal event frame: %v", err)
		}
		if f.Type != "event" {
			t.Fatalf("frame type = %q, want event", f.Type)
		}
		if f.Event != core.EvSongChanged {
			t.Fatalf("event name = %q, want %q", f.Event, core.EvSongChanged)
		}
		if id, _ := f.Data["id"].(float64); id != 42 {
			t.Fatalf("event data id = %v, want 42", f.Data["id"])
		}
	case <-time.After(3 * time.Second):
		t.Fatal("event frame did not arrive within 3s")
	}

	// An unsubscribed event must NOT be delivered (per-connection filter).
	server.broadcastEvent(core.EvStateChanged, frontend.EventFrame(core.EvStateChanged, map[string]any{"state": "paused"}))
	select {
	case frame := <-client.Events():
		t.Fatalf("received unsubscribed frame: %s", frame)
	case <-time.After(300 * time.Millisecond):
	}
}

// TestSubscribeClientDisconnect verifies the disconnect semantics: when the
// daemon goes away, the client reports Closed, the Events channel closes and a
// subsequent Call fails fast.
func TestSubscribeClientDisconnect(t *testing.T) {
	engine := newTestEngine(t)
	sock := shortSockPath(t)
	server := NewServerWithAddr(engine, "unix", sock)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Serve(ctx) }()
	t.Cleanup(server.Close)
	waitForSocket(t, sock)

	client, err := dialSubscribeAddr("unix", sock, []string{core.EvStateChanged})
	if err != nil {
		t.Fatalf("dialSubscribeAddr error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	// Drain the snapshot frame (always delivered first).
	<-client.Events()

	// Kill the daemon: the client's read loop sees EOF and tears down.
	server.Close()

	deadline := time.Now().Add(3 * time.Second)
	for !client.Closed() {
		if time.Now().After(deadline) {
			t.Fatal("client did not report Closed after daemon shutdown")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// The Events channel closes.
	select {
	case _, ok := <-client.Events():
		if ok {
			t.Fatal("Events channel still open after disconnect")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Events channel did not close within 3s")
	}

	// A subsequent Call fails fast.
	if _, err := client.Call(context.Background(), "status", nil); err == nil {
		t.Fatal("Call after disconnect succeeded, want error")
	}
}

// TestDialSubscribeNoDaemon verifies DialSubscribe reports the no-daemon error
// (the address resolution points at the test root's socket, which never
// exists).
func TestDialSubscribeNoDaemon(t *testing.T) {
	if _, err := DialSubscribe([]string{core.EvStateChanged}); err == nil {
		t.Fatal("DialSubscribe succeeded without a daemon, want error")
	}
}
