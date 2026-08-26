package webui

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// The beep speaker can only be initialized once per process, so all WS tests
// share ONE engine with process-lifetime config/storage setup — the same
// pattern as the headless control-channel tests. TestMain also covers the
// package's other tests (security_test.go exercises auth only and never
// touches the engine).
var (
	sharedEngineOnce sync.Once
	sharedEngineVal  *core.Engine

	testRoot string
)

func TestMain(m *testing.M) {
	var err error
	testRoot, err = os.MkdirTemp("", "musicfox-webui-test-*")
	if err != nil {
		panic(err)
	}

	prevRoot := os.Getenv("MUSICFOX_ROOT")
	prevConfig := configs.AppConfig
	prevDB := storage.DBManager

	// The first app path access bootstraps the path manager from MUSICFOX_ROOT,
	// so any db/cache files land in the temp root, never in user dirs.
	_ = os.Setenv("MUSICFOX_ROOT", testRoot)
	configs.AppConfig = &configs.Config{Player: configs.PlayerConfig{Engine: types.BeepPlayer}}
	storage.DBManager = new(storage.LocalDBManager)

	code := m.Run()

	storage.DBManager = prevDB
	configs.AppConfig = prevConfig
	if prevRoot == "" {
		_ = os.Unsetenv("MUSICFOX_ROOT")
	} else {
		_ = os.Setenv("MUSICFOX_ROOT", prevRoot)
	}
	_ = os.RemoveAll(testRoot)
	os.Exit(code)
}

// newTestEngine returns the shared real core engine (no Startup). Its
// engine.Player() is usable immediately after construction.
func newTestEngine(t *testing.T) *core.Engine {
	t.Helper()
	sharedEngineOnce.Do(func() {
		sharedEngineVal = core.NewEngine(core.EngineOptions{})
	})
	return sharedEngineVal
}

// newWSServer builds a Server bound to the shared engine and serves its real
// mux over httptest.
func newWSServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	s := NewServer(newTestEngine(t))
	ts := httptest.NewServer(s.mux)
	t.Cleanup(ts.Close)
	return s, ts
}

// wsURL converts the httptest http URL into the matching ws URL.
func wsURL(ts *httptest.Server) string {
	return "ws://" + strings.TrimPrefix(ts.URL, "http://") + "/ws"
}

// wsDialAuthed walks the real token-exchange handshake (GET /token → HttpOnly
// session cookie), opens a WebSocket connection carrying that cookie and drains
// the initial snapshot frame (the server always sends it first, before any
// command response).
func wsDialAuthed(t *testing.T, s *Server, ts *httptest.Server) (*websocket.Conn, []byte) {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New: %v", err)
	}
	client := &http.Client{Jar: jar}
	resp, err := client.Get(ts.URL + "/token?token=" + s.token)
	if err != nil {
		t.Fatalf("GET /token: %v", err)
	}
	_ = resp.Body.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, wsURL(ts), &websocket.DialOptions{HTTPClient: client})
	if err != nil {
		t.Fatalf("websocket dial: %v", err)
	}
	t.Cleanup(func() { _ = c.Close(websocket.StatusNormalClosure, "test done") })
	snapshot := wsReadRaw(t, c)
	return c, snapshot
}

// wsReadRaw reads the next WebSocket message as raw bytes.
func wsReadRaw(t *testing.T, c *websocket.Conn) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("ws read: %v", err)
	}
	return data
}

// wsTryReadRaw attempts to read the next message within d; ok=false when no
// frame arrives in time (used to assert a frame was NOT sent, e.g. throttled).
func wsTryReadRaw(t *testing.T, c *websocket.Conn, d time.Duration) (raw []byte, ok bool) {
	t.Helper()
	type result struct {
		data []byte
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		_, data, err := c.Read(context.Background())
		ch <- result{data, err}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			return nil, false
		}
		return r.data, true
	case <-time.After(d):
		return nil, false
	}
}

// wsWrite sends a Request over a WebSocket connection.
func wsWrite(t *testing.T, c *websocket.Conn, req *core.Request) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := wsjson.Write(ctx, c, req); err != nil {
		t.Fatalf("wsjson.Write: %v", err)
	}
}

// wsRead receives the next Response over a WebSocket connection.
func wsRead(t *testing.T, c *websocket.Conn) core.Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var resp core.Response
	if err := wsjson.Read(ctx, c, &resp); err != nil {
		t.Fatalf("wsjson.Read: %v", err)
	}
	return resp
}

// TestWSServerStatus exercises the happy path: a status request is answered
// with ok:true and a data payload over the real mux + dispatcher + engine.
func TestWSServerStatus(t *testing.T) {
	s, ts := newWSServer(t)
	c, _ := wsDialAuthed(t, s, ts)

	wsWrite(t, c, &core.Request{V: core.ProtocolVersion, ID: 1, Cmd: "status"})
	resp := wsRead(t, c)
	if resp.V != core.ProtocolVersion {
		t.Fatalf("response v = %d, want %d", resp.V, core.ProtocolVersion)
	}
	if resp.ID != 1 {
		t.Fatalf("response id = %d, want 1", resp.ID)
	}
	if !resp.Ok {
		t.Fatalf("response ok = false, error = %q", resp.Error)
	}
	if resp.Data == nil {
		t.Fatal("status response has nil data")
	}
	if _, ok := resp.Data.(map[string]any); !ok {
		t.Fatalf("status data = %T, want map[string]any", resp.Data)
	}
}

// TestWSServerUnknownCommand verifies that an unhandled command answers
// ok:false with the dispatcher's error and does not kill the connection.
func TestWSServerUnknownCommand(t *testing.T) {
	s, ts := newWSServer(t)
	c, _ := wsDialAuthed(t, s, ts)

	wsWrite(t, c, &core.Request{V: core.ProtocolVersion, ID: 2, Cmd: "nonsense"})
	resp := wsRead(t, c)
	if resp.ID != 2 {
		t.Fatalf("response id = %d, want 2", resp.ID)
	}
	if resp.Ok {
		t.Fatal("response ok = true, want false")
	}
	if resp.Error == "" {
		t.Fatal("response missing error message")
	}

	// The connection survives an error response: a follow-up still works.
	wsWrite(t, c, &core.Request{V: core.ProtocolVersion, ID: 3, Cmd: "status"})
	if resp := wsRead(t, c); !resp.Ok {
		t.Fatalf("follow-up status ok = false, error = %q", resp.Error)
	}
}

// TestWSServerQuit verifies the transport-layer shutdown: "quit" is answered
// ok:true and then shuts the server down (ShutdownCh closes).
func TestWSServerQuit(t *testing.T) {
	s, ts := newWSServer(t)
	c, _ := wsDialAuthed(t, s, ts)

	wsWrite(t, c, &core.Request{V: core.ProtocolVersion, ID: 5, Cmd: "quit"})
	resp := wsRead(t, c)
	if resp.ID != 5 {
		t.Fatalf("response id = %d, want 5", resp.ID)
	}
	if !resp.Ok {
		t.Fatalf("quit response ok = false, error = %q", resp.Error)
	}

	select {
	case <-s.ShutdownCh():
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down after quit command")
	}
}

// TestWSServerRejectsNoCookie verifies that a handshake without the session
// cookie is rejected before the upgrade (plain HTTP 401).
func TestWSServerRejectsNoCookie(t *testing.T) {
	_, ts := newWSServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, resp, err := websocket.Dial(ctx, wsURL(ts), nil)
	if c != nil {
		_ = c.Close(websocket.StatusNormalClosure, "test done")
	}
	if err == nil {
		t.Fatal("dial succeeded without a session cookie, want 401")
	}
	if resp == nil {
		t.Fatal("dial error returned nil HTTP response")
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("rejection status = %d, want 401", resp.StatusCode)
	}
}
