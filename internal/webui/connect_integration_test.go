package webui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/headless"
)

// startDaemonWithPlugin starts a DaemonPlugin-backed headless daemon on the
// DialAddr-resolved address (the webui test root's unix socket), so a
// headless.DialSubscribe client receives live events the way the production
// daemon streams them (emitter → DaemonPlugin → broadcastEvent → socket). The
// scope and server are cleaned up on test end.
func startDaemonWithPlugin(t *testing.T, engine *core.Engine) *headless.Server {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("daemon integration tests use a unix socket; the Windows TCP path is covered by the manual smoke checklist")
	}
	server := headless.NewServerWithAddr(engine, "unix", headless.ListenAddr())
	scope := framework.NewScope()
	if err := scope.Add(headless.NewDaemonPlugin(server)); err != nil {
		t.Fatalf("scope.Add(daemonPlugin) error = %v", err)
	}
	if err := scope.Start(engine.Ctx()); err != nil {
		t.Fatalf("scope.Start error = %v", err)
	}
	t.Cleanup(func() { _ = scope.Dispose() })
	t.Cleanup(server.Close)
	waitDaemonReady(t)
	return server
}

// commandKeySet snapshots the keys of the in-process frontend command registry.
func commandKeySet() map[string]bool {
	set := make(map[string]bool)
	for _, cmd := range frontend.Commands() {
		set[cmd.Key] = true
	}
	return set
}

// TestConnectIntegration exercises the full connect-mode chain against a real
// headless daemon: headless server (unix socket) → SubscribeClient →
// remoteBackend → connect Server (Auth=true) served over httptest. It locks in
// the functional boundary table (connect.go header): daemon-backed status and
// events reach the browser, the command surface contributes nothing, album art
// is 404, lyrics are the empty structure and QR login cannot complete (503).
func TestConnectIntegration(t *testing.T) {
	engine := newTestEngine(t)
	startDaemonWithPlugin(t, engine)

	// The connect-mode client over the daemon subscription (the same exported
	// entry connectRun uses).
	client, err := headless.DialSubscribe(eventWireNames())
	if err != nil {
		t.Fatalf("DialSubscribe error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	backend := newRemoteBackend(client)

	// The connect-mode Server over the remote backend, auth enabled.
	server := NewServerWithOptions(backend, ServerOptions{Auth: true})
	ts := httptest.NewServer(server.mux)
	t.Cleanup(ts.Close)
	t.Cleanup(func() { _ = server.Close() })

	// /api/status reflects the daemon's state: the HTTP response and a direct
	// backend dispatch (same client → daemon) agree on the stable fields, and
	// the fresh daemon engine reports an empty playlist.
	resp := doAuthedRequest(t, ts, "/api/status", server.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/status = %d, want 200", resp.StatusCode)
	}
	var statusBody struct {
		OK   bool           `json:"ok"`
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&statusBody); err != nil {
		t.Fatalf("decode /api/status body: %v", err)
	}
	if !statusBody.OK {
		t.Fatal("ok = false, want true")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	direct, err := backend.Dispatch(ctx, "status", nil)
	if err != nil {
		t.Fatalf("backend.Dispatch(status) error = %v", err)
	}
	directStatus, _ := direct.(map[string]any)
	if statusBody.Data["state"] != directStatus["state"] {
		t.Fatalf("HTTP status state = %v, direct = %v", statusBody.Data["state"], directStatus["state"])
	}
	if n, ok := numValue(statusBody.Data["playlistLen"]); !ok || n != 0 {
		t.Fatalf("HTTP status playlistLen = %v, want 0 (fresh daemon engine)", statusBody.Data["playlistLen"])
	}

	// WS: the snapshot frame arrives first (rebuilt by the connect server, not
	// re-broadcast from the daemon), then a control round-trip forwards to the
	// daemon, then a live song_changed event pushed by the daemon arrives.
	ws, wsSnapshot := wsDialAuthed(t, server, ts)
	var snap struct {
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(wsSnapshot, &snap); err != nil {
		t.Fatalf("unmarshal WS snapshot: %v", err)
	}
	if snap.Type != "snapshot" {
		t.Fatalf("first WS frame type = %q, want snapshot", snap.Type)
	}
	for _, key := range []string{"playing", "state", "song", "playlist"} {
		if _, ok := snap.Data[key]; !ok {
			t.Fatalf("WS snapshot missing key %q (keys: %v)", key, snap.Data)
		}
	}

	wsWrite(t, ws, &core.Request{V: core.ProtocolVersion, ID: 7, Cmd: "status"})
	if resp := wsRead(t, ws); !resp.Ok {
		t.Fatalf("WS status ok = false, error = %q", resp.Error)
	}

	// Emit the song-changed event on the daemon engine's event bus exactly the
	// way core does (payload carries the frame data shape); it must travel
	// daemon → SubscribeClient → remoteBackend → WS.
	emitter, ok := framework.ServiceOf[*framework.EventEmitter](engine.Ctx(), core.ServiceEventBus)
	if !ok {
		t.Fatal("eventBus not resolved from engine ctx")
	}
	if err := emitter.Emit(engine.Ctx(), core.EvSongChanged, map[string]any{
		"id":              int64(42),
		"name":            "测试歌曲",
		"artist":          "歌手A",
		"album":           "专辑X",
		"picUrl":          "https://p1.music.126.net/cover.jpg",
		"durationSeconds": 180.0,
	}); err != nil {
		t.Fatalf("emit song_changed: %v", err)
	}
	event, data := readEventFrame(t, ws)
	if event != "song_changed" {
		t.Fatalf("event = %q, want song_changed", event)
	}
	if id, _ := data["id"].(float64); id != 42 {
		t.Fatalf("song id = %v, want 42", data["id"])
	}
	if data["name"] != "测试歌曲" {
		t.Fatalf("song name = %v, want 测试歌曲", data["name"])
	}

	// /api/commands: the connect server contributes nothing. This test binary
	// shares the package-global registry with commands_test.go's t6test-*
	// registrations, so assert the meaningful invariant instead of a literal
	// empty list: every returned item pre-existed in the local process registry
	// (a real connect process loads no WASM scope and has none).
	preExisting := commandKeySet()
	resp = doAuthedRequest(t, ts, "/api/commands", server.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/commands = %d, want 200", resp.StatusCode)
	}
	var cmdBody struct {
		OK   bool          `json:"ok"`
		Data []commandItem `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cmdBody); err != nil {
		t.Fatalf("decode /api/commands body: %v", err)
	}
	if !cmdBody.OK {
		t.Fatal("commands ok = false, want true")
	}
	for _, item := range cmdBody.Data {
		if !preExisting[item.Key] {
			t.Fatalf("connect-mode /api/commands exposes non-local command %q", item.Key)
		}
	}

	// /api/albumart: connect mode has no PicUrl (the daemon status snapshot
	// carries none), so the endpoint answers 404.
	resp = doAuthedRequest(t, ts, "/api/albumart", server.token, "", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET /api/albumart = %d, want 404", resp.StatusCode)
	}

	// /api/lyrics: the empty structure (200, not 500).
	resp = doAuthedRequest(t, ts, "/api/lyrics", server.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/lyrics = %d, want 200", resp.StatusCode)
	}
	var lyrics map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&lyrics); err != nil {
		t.Fatalf("decode /api/lyrics body: %v", err)
	}
	for _, key := range []string{"fragments", "translatedFragments", "currentIndex", "offsetMs"} {
		if _, ok := lyrics[key]; !ok {
			t.Fatalf("lyrics body missing key %q: %v", key, lyrics)
		}
	}
	if frags, ok := lyrics["fragments"].([]any); ok && len(frags) != 0 {
		t.Fatalf("fragments = %v, want empty", frags)
	}

	// /api/login/qr/status: login cannot complete without a local engine — the
	// confirmed-scan (803) path answers 503 (D-S5-2). The key/image endpoints
	// need no engine and are not gated, so they stay reachable (not asserted
	// here as 503).
	prevStatus := qrCheckStatus
	qrCheckStatus = func(_ string, _ http.CookieJar) (float64, []byte, error) {
		return 803, []byte(`{"code":803}`), nil
	}
	t.Cleanup(func() { qrCheckStatus = prevStatus })
	resp = doAuthedRequest(t, ts, "/api/login/qr/status?key=uniKey-abc", server.token, "", "")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("GET /api/login/qr/status (803) = %d, want 503", resp.StatusCode)
	}
}
