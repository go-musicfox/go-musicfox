package core

import (
	"sync"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// emitterRecorder collects (name, payload) pairs from the event bus for
// double-write assertions.
type emitterRecorder struct {
	mu  sync.Mutex
	seq []string
	got map[string]any
}

func newEmitterRecorder() *emitterRecorder {
	return &emitterRecorder{got: map[string]any{}}
}

func (r *emitterRecorder) listen(emitter *framework.EventEmitter, names ...string) {
	for _, name := range names {
		name := name
		emitter.Listener(name, func(_ *framework.Context, payload any) error {
			r.mu.Lock()
			defer r.mu.Unlock()
			r.seq = append(r.seq, name)
			r.got[name] = payload
			return nil
		})
	}
}

func (r *emitterRecorder) received(name string) (any, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.got[name]
	return v, ok
}

// TestPlayerEventsDoubleWriteRerenderAndPlaylistEnd asserts the observer-first
// then emitter double-write on the two playback call sites reachable on an
// empty player: CtrlRerender and the empty-playlist bottom/top-out paths.
func TestPlayerEventsDoubleWriteRerenderAndPlaylistEnd(t *testing.T) {
	emitter := framework.NewEventEmitter()
	rec := newEmitterRecorder()
	rec.listen(emitter, EvRerender, EvPlaylistEnd)

	o := newFullObserver()
	p := NewEmptyPlayer()
	p.SetObserver(o)
	p.emitter = emitter

	// CtrlRerender → observer.OnRerender + EvRerender on the bus.
	p.handleControlSignal(CtrlSignal{Type: CtrlRerender})
	if _, ok := recvCh(o.rerenderCh); !ok {
		t.Fatal("observer OnRerender was not dispatched for CtrlRerender")
	}
	if _, ok := rec.received(EvRerender); !ok {
		t.Fatal("EvRerender was not emitted on the event bus")
	}

	// Empty-playlist NextSong bottom-out → observer.OnPlaylistExhausted +
	// EvPlaylistEnd (direction next).
	p.NextSong(false)
	if dir, ok := recvCh(o.exhaustedDirs); !ok || dir != DurationNext {
		t.Fatalf("observer OnPlaylistExhausted dir = %v, ok=%v; want DurationNext", dir, ok)
	}
	if payload, ok := rec.received(EvPlaylistEnd); !ok {
		t.Fatal("EvPlaylistEnd was not emitted on the event bus")
	} else if m := payload.(map[string]any); m["direction"] != "next" {
		t.Fatalf("EvPlaylistEnd direction = %v, want next", m["direction"])
	}

	// Empty-playlist PreviousSong top-out → observer.OnPlaylistExhausted +
	// EvPlaylistEnd (direction prev).
	p.PreviousSong(false)
	if dir, ok := recvCh(o.exhaustedDirs); !ok || dir != DurationPrev {
		t.Fatalf("observer OnPlaylistExhausted dir = %v, ok=%v; want DurationPrev", dir, ok)
	}
	if payload, ok := rec.received(EvPlaylistEnd); !ok {
		t.Fatal("EvPlaylistEnd was not emitted on the event bus")
	} else if m := payload.(map[string]any); m["direction"] != "prev" {
		t.Fatalf("EvPlaylistEnd direction = %v, want prev", m["direction"])
	}

	// No unexpected extra dispatches (observer side must stay single-fire).
	if _, ok := recvCh(o.rerenderCh); ok {
		t.Fatal("unexpected extra OnRerender dispatch")
	}
	if _, ok := recvCh(o.exhaustedDirs); ok {
		t.Fatal("unexpected extra OnPlaylistExhausted dispatch")
	}
}

// TestPlayerEmitHelperForwardsEventPayloads verifies the Player.emit helper
// forwards the state/song/position events with the wire payload shapes the
// subscribers expect (the observer side of those three sites lives in the
// player loops and is covered by the double-write pattern above).
func TestPlayerEmitHelperForwardsEventPayloads(t *testing.T) {
	emitter := framework.NewEventEmitter()
	rec := newEmitterRecorder()
	rec.listen(emitter, EvStateChanged, EvSongChanged, EvPosition)

	p := NewEmptyPlayer()
	p.emitter = emitter

	p.emit(EvStateChanged, map[string]any{"state": stateName(types.Paused)})
	p.emit(EvSongChanged, songEventPayload(structs.Song{
		Id:       42,
		Name:     "测试歌曲",
		Artists:  []structs.Artist{{Id: 1, Name: "歌手A"}},
		Album:    structs.Album{Id: 2, Name: "专辑X", PicUrl: "https://p1.music.126.net/cover.jpg"},
		Duration: 3 * time.Minute,
	}))
	p.emit(EvPosition, map[string]any{"positionSeconds": 30.0})

	payload, ok := rec.received(EvStateChanged)
	if !ok {
		t.Fatal("EvStateChanged not received")
	}
	if m := payload.(map[string]any); m["state"] != "paused" {
		t.Fatalf("EvStateChanged state = %v, want paused", m["state"])
	}

	payload, ok = rec.received(EvSongChanged)
	if !ok {
		t.Fatal("EvSongChanged not received")
	}
	m := payload.(map[string]any)
	if m["id"] != int64(42) || m["name"] != "测试歌曲" || m["artist"] != "歌手A" ||
		m["album"] != "专辑X" || m["picUrl"] != "https://p1.music.126.net/cover.jpg" {
		t.Fatalf("EvSongChanged payload = %v", m)
	}
	if secs := m["durationSeconds"].(float64); secs != 180 {
		t.Fatalf("EvSongChanged durationSeconds = %v, want 180", secs)
	}

	payload, ok = rec.received(EvPosition)
	if !ok {
		t.Fatal("EvPosition not received")
	}
	if m := payload.(map[string]any); m["positionSeconds"] != 30.0 {
		t.Fatalf("EvPosition positionSeconds = %v, want 30", m["positionSeconds"])
	}
}

// TestEngineEmitStartupPhaseDoubleWrite asserts the Engine startup-phase emit
// reaches both the observer and the event bus.
func TestEngineEmitStartupPhaseDoubleWrite(t *testing.T) {
	emitter := framework.NewEventEmitter()
	rec := newEmitterRecorder()
	rec.listen(emitter, EvStartupPhase)

	o := newFullObserver()
	e := &Engine{ctx: &framework.Context{}}
	e.ctx.Provide(ServiceEventBus, emitter)
	e.emitStartupPhase(o, StartupPhasePlaylistLoaded)

	if phase, ok := recvCh(o.startupPhases); !ok || phase != StartupPhasePlaylistLoaded {
		t.Fatalf("observer OnStartupPhase phase = %v, ok=%v; want %v", phase, ok, StartupPhasePlaylistLoaded)
	}
	payload, ok := rec.received(EvStartupPhase)
	if !ok {
		t.Fatal("EvStartupPhase not emitted on the event bus")
	}
	if m := payload.(map[string]any); m["phase"] != string(StartupPhasePlaylistLoaded) {
		t.Fatalf("EvStartupPhase phase = %v, want %q", m["phase"], StartupPhasePlaylistLoaded)
	}
}

// TestEngineEmitNilSafe verifies the Engine emit helper is a no-op when no bus
// is registered (nil-safe for bare engines and error paths).
func TestEngineEmitNilSafe(t *testing.T) {
	e := &Engine{}                                       // nil ctx, no bus
	e.emit(EvStartupPhase, map[string]any{"phase": "x"}) // must not panic
	e.emitStartupPhase(coreOnlyObserver{}, StartupPhaseUserRestored)
}

// TestLoginEventPayloadShape verifies the login event data shape (and its
// nil-user fallback).
func TestLoginEventPayloadShape(t *testing.T) {
	user := &structs.User{UserId: 42, Nickname: "tester", AvatarUrl: "https://p1.music.126.net/a.png"}
	m := loginEventPayload(user)["user"].(map[string]any)
	if m["userId"] != int64(42) || m["nickname"] != "tester" || m["avatarUrl"] != "https://p1.music.126.net/a.png" {
		t.Fatalf("login payload user = %v", m)
	}
	if _, ok := loginEventPayload(nil)["user"]; !ok {
		t.Fatal("nil user must still yield a user key")
	}
}
