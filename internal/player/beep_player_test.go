package player

import (
	"testing"

	"github.com/gopxl/beep"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

func newIdleBeepPlayer() *beepPlayer {
	p := &beepPlayer{
		ctrl:      &beep.Ctrl{},
		stateChan: make(chan types.State, 10),
	}
	p.state.Store(uint32(types.Stopped))
	return p
}

func TestBeepPlayerResumeBeforeFirstSongStaysStopped(t *testing.T) {
	p := newIdleBeepPlayer()

	// Remote-control Resume/Toggle can arrive before any song is decoded, when
	// the timer is still nil. It must not report Playing in that state.
	p.Resume()
	if got := p.State(); got != types.Stopped {
		t.Fatalf("state after resume without a song = %v, want Stopped", got)
	}

	// And the subsequent Pause/Stop must not dereference the nil timer.
	p.Pause()
	p.Stop()
	if got := p.State(); got != types.Stopped {
		t.Fatalf("state after pause/stop without a song = %v, want Stopped", got)
	}
}

func TestBeepPlayerPauseAndStopWithNilTimerDoNotPanic(t *testing.T) {
	p := newIdleBeepPlayer()
	p.state.Store(uint32(types.Playing))

	p.Pause()
	if got := p.State(); got != types.Paused {
		t.Fatalf("state after pause = %v, want Paused", got)
	}

	p.Stop()
	if got := p.State(); got != types.Stopped {
		t.Fatalf("state after stop = %v, want Stopped", got)
	}
}
