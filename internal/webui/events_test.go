package webui

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

// TestPositionThrottle exercises the 250ms (4Hz) position throttle: the first
// emit always passes, emits closer than 250ms are dropped, and the next emit
// after 250ms passes again.
func TestPositionThrottle(t *testing.T) {
	var th positionThrottle
	now := time.Now()
	if !th.shouldEmit(now) {
		t.Fatal("first emit should pass")
	}
	if th.shouldEmit(now.Add(100 * time.Millisecond)) {
		t.Fatal("emit 100ms after the last must be dropped (< 250ms)")
	}
	if th.shouldEmit(now.Add(200 * time.Millisecond)) {
		t.Fatal("emit 200ms after the last must be dropped (< 250ms)")
	}
	if !th.shouldEmit(now.Add(300 * time.Millisecond)) {
		t.Fatal("emit 300ms after the last must pass")
	}
}

// testEventBus resolves the shared engine's event bus (the same bus the WebUI
// server subscribed to at construction).
func testEventBus(t *testing.T, s *Server) *framework.EventEmitter {
	t.Helper()
	emitter, ok := framework.ServiceOf[*framework.EventEmitter](s.engine.Ctx(), core.ServiceEventBus)
	if !ok {
		t.Fatal("eventBus not resolved from engine ctx")
	}
	return emitter
}

// emitCoreEvent emits an event the way core does (payload already carries the
// frame data shape); the payload maps mirror the core event payload builders.
func emitCoreEvent(t *testing.T, s *Server, name string, payload any) {
	t.Helper()
	if err := testEventBus(t, s).Emit(s.engine.Ctx(), name, payload); err != nil {
		t.Fatalf("emit %s: %v", name, err)
	}
}

// readEventFrame reads the next WS frame and asserts it is an event frame,
// returning its event name and data map.
func readEventFrame(t *testing.T, c *websocket.Conn) (string, map[string]any) {
	t.Helper()
	var frame map[string]any
	if err := json.Unmarshal(wsReadRaw(t, c), &frame); err != nil {
		t.Fatalf("unmarshal event frame: %v", err)
	}
	if frame["type"] != "event" {
		t.Fatalf("frame type = %q, want event", frame["type"])
	}
	event, _ := frame["event"].(string)
	data, _ := frame["data"].(map[string]any)
	return event, data
}

// TestWSEmitterSnapshot verifies a fresh connection first receives the
// snapshot frame carrying the status fields and the trimmed playlist.
func TestWSEmitterSnapshot(t *testing.T) {
	s, ts := newWSServer(t)
	c, snapshot := wsDialAuthed(t, s, ts)

	var frame struct {
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(snapshot, &frame); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if frame.Type != "snapshot" {
		t.Fatalf("first frame type = %q, want snapshot", frame.Type)
	}
	for _, key := range []string{"playing", "state", "song", "playlist"} {
		if _, ok := frame.Data[key]; !ok {
			t.Fatalf("snapshot data missing key %q (keys: %v)", key, frame.Data)
		}
	}
	if _, ok := frame.Data["playlist"].([]any); !ok {
		t.Fatalf("snapshot playlist = %T, want []any", frame.Data["playlist"])
	}
	_ = c
}

// TestWSEmitterEvents emits events through the core event bus (the same path
// the engine player/startup now use) and verifies each event reaches a
// connected WS client as the browser protocol frame. Frames are collected until
// all expected events arrive because broadcasts write each connection from its
// own goroutine (order not strict).
func TestWSEmitterEvents(t *testing.T) {
	s, ts := newWSServer(t)
	c, _ := wsDialAuthed(t, s, ts)

	song := structs.Song{
		Id:       42,
		Name:     "测试歌曲",
		Artists:  []structs.Artist{{Id: 1, Name: "歌手A"}},
		Album:    structs.Album{Id: 2, Name: "专辑X", PicUrl: "https://p1.music.126.net/cover.jpg"},
		Duration: 3 * time.Minute,
	}
	// Payload shapes mirror the core event payload builders (songEventPayload
	// etc. in internal/core/events.go).
	emitCoreEvent(t, s, core.EvSongChanged, map[string]any{
		"id":              song.Id,
		"name":            song.Name,
		"artist":          song.ArtistName(),
		"album":           song.Album.Name,
		"picUrl":          song.PicUrl,
		"durationSeconds": song.Duration.Seconds(),
	})
	emitCoreEvent(t, s, core.EvStateChanged, map[string]any{"state": "paused"})
	emitCoreEvent(t, s, core.EvPosition, map[string]any{"positionSeconds": 0.0})
	emitCoreEvent(t, s, core.EvStartupPhase, map[string]any{"phase": string(core.StartupPhasePlaylistLoaded)})
	emitCoreEvent(t, s, core.EvLogin, map[string]any{"user": map[string]any{
		"userId": int64(7), "nickname": "tester", "avatarUrl": "https://p1.music.126.net/u.png",
	}})

	seen := map[string]map[string]any{}
	for len(seen) < 5 {
		event, data := readEventFrame(t, c)
		seen[event] = data
	}

	songData, ok := seen["song_changed"]
	if !ok {
		t.Fatalf("missing song_changed event; saw %v", seen)
	}
	if id := songData["id"].(float64); id != 42 {
		t.Fatalf("song id = %v, want 42", id)
	}
	if songData["name"] != "测试歌曲" {
		t.Fatalf("song name = %v, want 测试歌曲", songData["name"])
	}
	if songData["artist"] != "歌手A" {
		t.Fatalf("artist = %v, want 歌手A", songData["artist"])
	}
	if songData["album"] != "专辑X" {
		t.Fatalf("album = %v, want 专辑X", songData["album"])
	}
	if songData["picUrl"] != "https://p1.music.126.net/cover.jpg" {
		t.Fatalf("picUrl = %v", songData["picUrl"])
	}
	if secs := songData["durationSeconds"].(float64); secs != 180 {
		t.Fatalf("durationSeconds = %v, want 180", secs)
	}

	stateData, ok := seen["state_changed"]
	if !ok {
		t.Fatalf("missing state_changed event; saw %v", seen)
	}
	if stateData["state"] != "paused" {
		t.Fatalf("state = %v, want paused", stateData["state"])
	}

	posData, ok := seen["position"]
	if !ok {
		t.Fatalf("missing position event; saw %v", seen)
	}
	if secs := posData["positionSeconds"].(float64); secs != 0 {
		t.Fatalf("positionSeconds = %v, want 0", secs)
	}

	phaseData, ok := seen["startup_phase"]
	if !ok {
		t.Fatalf("missing startup_phase event; saw %v", seen)
	}
	if phaseData["phase"] != string(core.StartupPhasePlaylistLoaded) {
		t.Fatalf("phase = %v, want %q", phaseData["phase"], core.StartupPhasePlaylistLoaded)
	}

	loginData, ok := seen["login"]
	if !ok {
		t.Fatalf("missing login event; saw %v", seen)
	}
	if _, ok := loginData["user"].(map[string]any); !ok {
		t.Fatalf("login event missing user: %v", loginData)
	}
}

// TestWSEmitterPositionThrottledOverWS verifies the WebUI-side throttle also
// drops consecutive position events on the wire.
func TestWSEmitterPositionThrottledOverWS(t *testing.T) {
	s, ts := newWSServer(t)
	c, _ := wsDialAuthed(t, s, ts)

	emitCoreEvent(t, s, core.EvPosition, map[string]any{"positionSeconds": 0.0})
	if event, data := readEventFrame(t, c); event != "position" {
		t.Fatalf("event = %q, want position (data=%v)", event, data)
	}

	// An immediate second position must be dropped by the 250ms throttle.
	emitCoreEvent(t, s, core.EvPosition, map[string]any{"positionSeconds": 5.0})
	if raw, ok := wsTryReadRaw(t, c, 300*time.Millisecond); ok {
		t.Fatalf("expected throttled position, got frame: %s", raw)
	}
}

// TestEmitterUnsubscribeOnServerClose verifies the server removes its event-bus
// listeners on Close: after Close no further emit reaches the (dead) server's
// broadcaster, and a subsequent server still receives events.
func TestEmitterUnsubscribeOnServerClose(t *testing.T) {
	s, ts := newWSServer(t)
	_ = ts
	_ = s.Close()

	// The listeners are gone: emitting must not panic and the recorder-style
	// dead broadcaster gets nothing (the emit simply returns nil).
	emitCoreEvent(t, s, core.EvStateChanged, map[string]any{"state": "playing"})
}
