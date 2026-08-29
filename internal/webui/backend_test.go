package webui

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/headless"
)

// startRawDaemon starts a headless control server bound to the address that
// headless.DialAddr() resolves in the test environment (TestMain's
// MUSICFOX_ROOT scopes the default unix socket under the webui test root), so
// headless.DialSubscribe reaches it through the plain exported API. It is the
// raw server (no DaemonPlugin): enough for the control-plane assertions in this
// file; the event plane is exercised by connect_integration_test.go.
//
// Windows keeps its TCP port-file path as a manual smoke item (roadmap S5
// §2.4), so the unix-socket daemon tests skip there.
func startRawDaemon(t *testing.T, engine *core.Engine) *headless.Server {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("daemon integration tests use a unix socket; the Windows TCP path is covered by the manual smoke checklist")
	}
	server := headless.NewServerWithAddr(engine, "unix", headless.ListenAddr())
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = server.Serve(ctx) }()
	t.Cleanup(server.Close)
	waitDaemonReady(t)
	return server
}

// waitDaemonReady polls until the daemon accepts connections at the address
// headless.DialAddr() resolves to (the same resolution DialSubscribe uses).
func waitDaemonReady(t *testing.T) {
	t.Helper()
	network, addr := headless.DialAddr()
	if network == "" || addr == "" {
		t.Fatal("headless.DialAddr() empty; daemon address unavailable")
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err := net.Dial(network, addr)
		if err == nil {
			_ = conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("daemon not ready at %s: %v", addr, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// remoteBackendForTest dials the daemon through headless.DialSubscribe (the
// exported connect-mode entry) and wraps it in a remoteBackend.
func remoteBackendForTest(t *testing.T) *remoteBackend {
	t.Helper()
	client, err := headless.DialSubscribe(eventWireNames())
	if err != nil {
		t.Fatalf("DialSubscribe error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return newRemoteBackend(client)
}

// songSnapshot is a comparable projection of a song record taken from either a
// status map or a CommandContext; numbers are normalized to float64 so the
// JSON round-trip (int64 → float64) of the remote path compares equal to the
// in-process int64 of the local path.
type songSnapshot struct {
	id     float64
	name   string
	artist string
	album  string
}

func songFromStatus(song map[string]any) songSnapshot {
	out := songSnapshot{}
	if id, ok := numValue(song["id"]); ok {
		out.id = id
	}
	out.name, _ = song["name"].(string)
	out.artist, _ = song["artist"].(string)
	out.album, _ = song["album"].(string)
	return out
}

func songFromContext(info *frontend.SongInfo) songSnapshot {
	if info == nil {
		return songSnapshot{}
	}
	return songSnapshot{id: float64(info.ID), name: info.Name, artist: info.Artist, album: info.Album}
}

// numValue converts any JSON-number-ish value to float64.
func numValue(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

// assertBackendSemantics is the shared behavior contract for both Backend
// implementations (local engine vs remote daemon client): Ready is up, the
// "status" snapshot is well-formed, the trimmed playlist shape holds and
// CommandContext stays internally consistent with the status snapshot.
func assertBackendSemantics(t *testing.T, b Backend) {
	t.Helper()
	if !b.Ready() {
		t.Fatal("backend not ready")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := b.Dispatch(ctx, "status", nil)
	if err != nil {
		t.Fatalf("Dispatch(status) error = %v", err)
	}
	status, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("Dispatch(status) data = %T, want map[string]any", data)
	}
	for _, key := range []string{"playing", "state", "song", "positionSeconds", "durationSeconds", "volume", "mode", "playlistLen", "user"} {
		if _, ok := status[key]; !ok {
			t.Fatalf("status missing key %q (keys: %v)", key, status)
		}
	}

	playlist := b.Playlist()
	if playlist == nil {
		t.Fatal("Playlist() = nil, want non-nil slice")
	}
	for _, item := range playlist {
		for _, key := range []string{"id", "name", "artist", "album"} {
			if _, ok := item[key]; !ok {
				t.Fatalf("playlist item missing key %q: %v", key, item)
			}
		}
	}

	// CommandContext fields derive from the same state as the status snapshot.
	cc := b.CommandContext()
	playing, _ := status["playing"].(bool)
	if cc.Playing != playing {
		t.Fatalf("CommandContext.Playing = %v, want status playing = %v", cc.Playing, playing)
	}
	user, _ := status["user"].(string)
	if cc.UserName != user {
		t.Fatalf("CommandContext.UserName = %q, want status user = %q", cc.UserName, user)
	}
	song, _ := status["song"].(map[string]any)
	if got, want := songFromContext(cc.Song), songFromStatus(song); got != want {
		t.Fatalf("CommandContext.Song = %+v, want status song = %+v", got, want)
	}
}

// statusSnapshot is the comparable projection of a "status" dispatch used to
// prove the two backends agree with each other (number types normalized).
type statusSnapshot struct {
	playing     bool
	user        string
	song        songSnapshot
	playlistLen float64
}

func collectStatusSnapshot(t *testing.T, b Backend) statusSnapshot {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := b.Dispatch(ctx, "status", nil)
	if err != nil {
		t.Fatalf("Dispatch(status) error = %v", err)
	}
	status, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("Dispatch(status) data = %T, want map[string]any", data)
	}
	out := statusSnapshot{}
	out.playing, _ = status["playing"].(bool)
	out.user, _ = status["user"].(string)
	if song, ok := status["song"].(map[string]any); ok {
		out.song = songFromStatus(song)
	}
	if n, ok := numValue(status["playlistLen"]); ok {
		out.playlistLen = n
	}
	return out
}

// TestBackendSemanticsConsistency runs the shared assertion set against both
// Backend implementations on the same engine: the localBackend reads the
// engine directly while the remoteBackend dials a headless daemon bound to
// that same engine, so the status/playlist/context surface must agree. It also
// cross-checks the two implementations against each other on the stable status
// fields (a fresh engine never changes under the test).
func TestBackendSemanticsConsistency(t *testing.T) {
	engine := newTestEngine(t)

	local := &localBackend{engine: engine}
	assertBackendSemantics(t, local)

	startRawDaemon(t, engine)
	remote := remoteBackendForTest(t)
	assertBackendSemantics(t, remote)

	if localSnap, remoteSnap := collectStatusSnapshot(t, local), collectStatusSnapshot(t, remote); localSnap != remoteSnap {
		t.Fatalf("status snapshots differ: local = %+v, remote = %+v", localSnap, remoteSnap)
	}
}

// TestBackendAuthDisabledEndpointsReachable verifies the Auth=false server
// shape (the GUI AssetServer scheme where the cookie exchange is impossible):
// /api/status and /api/lyrics answer 200 with no session cookie and a
// WebSocket handshake succeeds without one either, still sending the snapshot
// frame first.
func TestBackendAuthDisabledEndpointsReachable(t *testing.T) {
	server := NewServerWithOptions(&localBackend{engine: newTestEngine(t)}, ServerOptions{Auth: false})
	ts := httptest.NewServer(server.mux)
	t.Cleanup(ts.Close)
	t.Cleanup(func() { _ = server.Close() })

	resp := doAuthedRequest(t, ts, "/api/status", "", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/status (no cookie) = %d, want 200", resp.StatusCode)
	}

	resp = doAuthedRequest(t, ts, "/api/lyrics", "", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/lyrics (no cookie) = %d, want 200", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, wsURL(ts), nil)
	if err != nil {
		t.Fatalf("websocket dial without cookie = %v, want success", err)
	}
	t.Cleanup(func() { _ = c.Close(websocket.StatusNormalClosure, "test done") })

	snapshot := wsReadRaw(t, c)
	var frame struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(snapshot, &frame); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if frame.Type != "snapshot" {
		t.Fatalf("first WS frame type = %q, want snapshot", frame.Type)
	}
}
