package core

import (
	"context"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

// TestControlSignalsOnFreshEngine verifies the headless control-channel
// regression: dispatching resume/toggle (and pause/stop/prev/next) on a fresh
// engine that has never loaded a song must neither panic nor flip the player
// state to Playing. Previously CtrlResume/CtrlToggle reached the engine's
// Resume()/Toggle() with a nil timer and corrupted the state to Playing, which
// then made the next Pause()/Stop() panic on a nil pointer.
func TestControlSignalsOnFreshEngine(t *testing.T) {
	engine := newTestEngine(t)
	d := NewDispatcher(engine)

	for _, cmd := range []string{"resume", "toggle", "pause", "stop", "prev", "next"} {
		if _, err := d.Dispatch(context.Background(), cmd, nil); err != nil {
			t.Fatalf("Dispatch(%s) on fresh engine error = %v", cmd, err)
		}
	}

	// Control signals are processed asynchronously; give the ctrl goroutine a
	// moment to apply any (wrong) state transition before asserting.
	time.Sleep(100 * time.Millisecond)
	if engine.Player().State() == types.Playing {
		t.Fatal("fresh engine must not report Playing after no-op control signals")
	}
}

// TestEngineResumeWithoutLoadedSongDoesNotCorruptState directly exercises the
// engine-level guard: Resume/Toggle with no song loaded (timer not yet created)
// must be a no-op instead of flipping the state to Playing.
func TestEngineResumeWithoutLoadedSongDoesNotCorruptState(t *testing.T) {
	engine := newTestEngine(t)
	ep := engine.Player().Engine()

	ep.Resume()
	ep.Toggle()
	ep.Pause()
	ep.Stop()
	ep.Seek(5 * time.Second)

	if ep.State() == types.Playing {
		t.Fatal("engine without a loaded song must not report Playing after Resume/Toggle")
	}
}
