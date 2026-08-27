package webui

import (
	"encoding/json"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// webuiNoopObserver implements core.Observer with no-ops. The WebUI frontend
// consumes playback/startup events through the core event bus (see
// subscribeEmitter) instead of the observer seam; the no-op observer keeps the
// engine Startup signature intact (mirrors headless.HeadlessObserver).
type webuiNoopObserver struct{}

// Compile-time assertion that webuiNoopObserver satisfies core.Observer.
var _ core.Observer = webuiNoopObserver{}

func (webuiNoopObserver) OnSongChanged(structs.Song) {}
func (webuiNoopObserver) OnStateChanged(types.State) {}
func (webuiNoopObserver) OnPosition(time.Duration)   {}

// positionMinInterval is the WebUI position emit interval. The engine already
// throttles position ticks; this second layer keeps a stable 4Hz event rate
// regardless of the UI frame rate.
const positionMinInterval = 250 * time.Millisecond

// positionThrottle drops position events closer than positionMinInterval to the
// previous emit. Not safe for concurrent use — the event-bus listener runs on
// the emitting player goroutine (a single goroutine per loop).
type positionThrottle struct {
	lastAt time.Time
}

// shouldEmit reports whether now is far enough from the last emit, updating the
// state on a hit.
func (t *positionThrottle) shouldEmit(now time.Time) bool {
	if now.Sub(t.lastAt) >= positionMinInterval {
		t.lastAt = now
		return true
	}
	return false
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

// eventWireToFrame maps the core event-bus wire names to the WebUI frame names
// (the subset the WebUI frontend consumes; the frame names keep the browser
// protocol unchanged). The payload emitted by core already carries the frame
// data shape, so a listener only renames the event and forwards it.
var eventWireToFrame = map[string]string{
	core.EvSongChanged:  "song_changed",
	core.EvStateChanged: "state_changed",
	core.EvPosition:     "position",
	core.EvStartupPhase: "startup_phase",
	core.EvLogin:        "login",
}

// subscribeEmitter registers the core event-bus listeners that forward the
// frontend-relevant events to the broadcaster (event frame format unchanged).
// The listeners are enqueue-only: broadcast snapshots the connection set under
// the lock and writes each socket from its own goroutine, so the emitting
// player goroutine is never blocked (docs/plugin_ecosystem.md §四 并发契约).
// The returned func unsubscribes the whole set on teardown (per-wire coarse
// EventEmitter.Unregister — a server owns its subscription for its lifetime).
func subscribeEmitter(emitter *framework.EventEmitter, b *broadcaster) func() {
	posThrottle := new(positionThrottle)
	for wire, frame := range eventWireToFrame {
		wire := wire
		frame := frame
		emitter.Listener(wire, func(_ *framework.Context, payload any) error {
			if frame == "position" && !posThrottle.shouldEmit(time.Now()) {
				return nil
			}
			b.broadcast(eventFrame(frame, payload))
			return nil
		})
	}
	return func() {
		for wire := range eventWireToFrame {
			emitter.Unregister(wire)
		}
	}
}
