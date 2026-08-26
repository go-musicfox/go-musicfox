package webui

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// TestPositionThrottle exercises the 250ms (4Hz) position throttle: the first
// emit always passes, emits closer than 250ms are dropped, and the next emit
// after 250ms passes again.
func TestPositionThrottle(t *testing.T) {
	var th positionThrottle
	now := time.Now()
	if !th.shouldEmit(0, now) {
		t.Fatal("first emit should pass")
	}
	if th.shouldEmit(time.Second, now.Add(100*time.Millisecond)) {
		t.Fatal("emit 100ms after the last must be dropped (< 250ms)")
	}
	if th.shouldEmit(time.Second, now.Add(200*time.Millisecond)) {
		t.Fatal("emit 200ms after the last must be dropped (< 250ms)")
	}
	if !th.shouldEmit(time.Second, now.Add(300*time.Millisecond)) {
		t.Fatal("emit 300ms after the last must pass")
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

// TestWSObserverSnapshot verifies a fresh connection first receives the
// snapshot frame carrying the status fields and the trimmed playlist.
func TestWSObserverSnapshot(t *testing.T) {
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

// TestWSObserverEvents drives the observer directly (the same instance NewServer
// attaches to the engine player) and verifies each event reaches a connected
// WS client. Frames are collected until all expected events arrive because
// broadcasts write each connection from its own goroutine (order not strict).
func TestWSObserverEvents(t *testing.T) {
	s, ts := newWSServer(t)
	c, _ := wsDialAuthed(t, s, ts)

	song := structs.Song{
		Id:       42,
		Name:     "测试歌曲",
		Artists:  []structs.Artist{{Id: 1, Name: "歌手A"}},
		Album:    structs.Album{Id: 2, Name: "专辑X", PicUrl: "https://p1.music.126.net/cover.jpg"},
		Duration: 3 * time.Minute,
	}
	s.observer.OnSongChanged(song)
	s.observer.OnStateChanged(types.Paused)
	s.observer.OnPosition(0)
	s.observer.OnStartupPhase(core.StartupPhasePlaylistLoaded)

	seen := map[string]map[string]any{}
	for len(seen) < 4 {
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
}

// TestWSObserverPositionThrottledOverWS verifies the second throttle also drops
// consecutive position events on the wire.
func TestWSObserverPositionThrottledOverWS(t *testing.T) {
	s, ts := newWSServer(t)
	c, _ := wsDialAuthed(t, s, ts)

	s.observer.OnPosition(0)
	if event, data := readEventFrame(t, c); event != "position" {
		t.Fatalf("event = %q, want position (data=%v)", event, data)
	}

	// An immediate second position must be dropped by the 250ms throttle.
	s.observer.OnPosition(5 * time.Second)
	if raw, ok := wsTryReadRaw(t, c, 300*time.Millisecond); ok {
		t.Fatalf("expected throttled position, got frame: %s", raw)
	}
}
