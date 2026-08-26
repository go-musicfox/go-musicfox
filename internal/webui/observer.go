package webui

import (
	"encoding/json"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// WebUIObserver implements core.Observer (the required playback trio) plus the
// optional core.StartupPhaseObserver. Playback/startup events are forwarded to
// every live WS connection via the broadcaster.
type WebUIObserver struct {
	b           *broadcaster
	posThrottle positionThrottle // second throttle: minInterval 250ms (4Hz)
}

// Compile-time assertions that WebUIObserver satisfies both observer contracts.
var _ core.Observer = (*WebUIObserver)(nil)
var _ core.StartupPhaseObserver = (*WebUIObserver)(nil)

// positionMinInterval is the WebUI OnPosition emit interval. The engine already
// throttles position ticks; this second layer keeps a stable 4Hz event rate
// regardless of the UI frame rate.
const positionMinInterval = 250 * time.Millisecond

// positionThrottle drops position events closer than positionMinInterval to the
// previous emit. Not safe for concurrent use — the player calls OnPosition from
// a single goroutine.
type positionThrottle struct {
	lastAt time.Time
}

// shouldEmit reports whether now is far enough from the last emit, updating the
// state on a hit (mirrors core.mprisPositionThrottle).
func (t *positionThrottle) shouldEmit(_ time.Duration, now time.Time) bool {
	if now.Sub(t.lastAt) >= positionMinInterval {
		t.lastAt = now
		return true
	}
	return false
}

// OnSongChanged pushes a song_changed event with only the frontend-relevant
// fields (never the whole structs.Song).
func (o *WebUIObserver) OnSongChanged(song structs.Song) {
	o.b.broadcast(eventFrame("song_changed", map[string]any{
		"id":              song.Id,
		"name":            song.Name,
		"artist":          song.ArtistName(),
		"album":           song.Album.Name,
		"picUrl":          song.PicUrl,
		"durationSeconds": song.Duration.Seconds(),
	}))
}

// OnStateChanged pushes a state_changed event with the stable wire state name.
func (o *WebUIObserver) OnStateChanged(state types.State) {
	o.b.broadcast(eventFrame("state_changed", map[string]any{"state": webuiStateName(state)}))
}

// OnPosition pushes a position event, rate-limited by posThrottle.
func (o *WebUIObserver) OnPosition(d time.Duration) {
	if !o.posThrottle.shouldEmit(d, time.Now()) {
		return
	}
	o.b.broadcast(eventFrame("position", map[string]any{"positionSeconds": d.Seconds()}))
}

// OnStartupPhase pushes a startup_phase milestone event.
func (o *WebUIObserver) OnStartupPhase(phase core.StartupPhase) {
	o.b.broadcast(eventFrame("startup_phase", map[string]any{"phase": string(phase)}))
}

// eventFrame marshals an event frame: {"type":"event","event":"<name>","data":{...}}.
func eventFrame(name string, data any) []byte {
	payload, err := json.Marshal(map[string]any{"type": "event", "event": name, "data": data})
	if err != nil {
		// data is always a map of JSON primitives; unreachable in practice.
		return nil
	}
	return payload
}

// webuiStateName maps a types.State to the stable lowercase wire string
// (mirrors core.stateName, which is unexported).
func webuiStateName(s types.State) string {
	switch s {
	case types.Playing:
		return "playing"
	case types.Paused:
		return "paused"
	case types.Stopped:
		return "stopped"
	case types.Interrupted:
		return "interrupted"
	default:
		return "unknown"
	}
}
