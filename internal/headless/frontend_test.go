package headless

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// TestHeadlessObserverZeroValueNoPanic ensures every core.Observer method is
// callable on a zero-value HeadlessObserver without panicking.
func TestHeadlessObserverZeroValueNoPanic(t *testing.T) {
	var o HeadlessObserver

	o.OnSongChanged(structs.Song{})
	o.OnStateChanged(types.Playing)
	o.OnPosition(time.Second)
	o.OnPlaylistExhausted(core.DurationNext)
	o.OnPlaylistExhausted(core.DurationPrev)
	o.OnRerender()
	o.OnStartupPhase(core.StartupPhaseUserRestored)
	o.OnStartupPhase(core.StartupPhasePlaylistLoaded)
	o.OnStartupPhase(core.StartupPhaseBeforeAutoplay)
}

// TestHeadlessObserverRequestLoginDoesNotInvokeCallback ensures RequestLogin
// degrades to a log line and never runs the afterLogin callback.
func TestHeadlessObserverRequestLoginDoesNotInvokeCallback(t *testing.T) {
	var o HeadlessObserver
	var called atomic.Bool
	o.RequestLogin(func() { called.Store(true) })
	if called.Load() {
		t.Fatal("RequestLogin invoked the afterLogin callback in headless mode")
	}
}
