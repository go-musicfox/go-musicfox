package core

import (
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// coreOnlyObserver implements only the three playback-required Observer
// methods. It must NOT implement any optional extension interface so the
// assertion + dispatch paths take the skip branch.
type coreOnlyObserver struct{}

func (coreOnlyObserver) OnSongChanged(structs.Song) {}
func (coreOnlyObserver) OnStateChanged(types.State) {}
func (coreOnlyObserver) OnPosition(time.Duration)   {}

// fullObserver embeds coreOnlyObserver (the three playback-required methods)
// and additionally implements every optional extension interface, recording
// each dispatched event on a buffered channel for assertions.
type fullObserver struct {
	coreOnlyObserver
	rerenderCh    chan struct{}
	exhaustedDirs chan PlayDirection
	startupPhases chan StartupPhase
}

func (o *fullObserver) RequestLogin(_ func()) {}

func (o *fullObserver) OnPlaylistExhausted(dir PlayDirection) {
	o.exhaustedDirs <- dir
}

func (o *fullObserver) OnRerender() {
	o.rerenderCh <- struct{}{}
}

func (o *fullObserver) OnStartupPhase(phase StartupPhase) {
	o.startupPhases <- phase
}

func newFullObserver() *fullObserver {
	return &fullObserver{
		rerenderCh:    make(chan struct{}, 4),
		exhaustedDirs: make(chan PlayDirection, 4),
		startupPhases: make(chan StartupPhase, 4),
	}
}

func recvCh[T any](ch <-chan T) (T, bool) {
	select {
	case v := <-ch:
		return v, true
	default:
		var zero T
		return zero, false
	}
}

// TestObserverOptionalDispatchSkippedForCoreOnly verifies that an observer
// implementing only the three playback-required events is dispatched without
// panic: CtrlRerender, bottom-out NextSong and emitStartupPhase all take the
// assertion-fails branch and skip the optional event.
func TestObserverOptionalDispatchSkippedForCoreOnly(t *testing.T) {
	p := NewEmptyPlayer()
	p.SetObserver(coreOnlyObserver{})

	// CtrlRerender signal: observer does not implement RerenderObserver.
	p.handleControlSignal(CtrlSignal{Type: CtrlRerender})

	// Empty playlist hits the bottom-out path in NextSong — OnPlaylistExhausted
	// would fire, but the observer does not implement PlaylistExhaustedObserver.
	p.NextSong(false)

	// emitStartupPhase takes a core.Observer; the core-only observer is one.
	testEngine().emitStartupPhase(coreOnlyObserver{}, StartupPhaseUserRestored)
}

// TestObserverOptionalDispatchReachesFullObserver verifies that an observer
// implementing every optional extension interface receives OnRerender,
// OnPlaylistExhausted and OnStartupPhase via assertion + dispatch.
func TestObserverOptionalDispatchReachesFullObserver(t *testing.T) {
	o := newFullObserver()
	p := NewEmptyPlayer()
	p.SetObserver(o)

	// CtrlRerender signal dispatches OnRerender.
	p.handleControlSignal(CtrlSignal{Type: CtrlRerender})
	if _, ok := recvCh(o.rerenderCh); !ok {
		t.Fatal("OnRerender was not dispatched for CtrlRerender")
	}

	// Empty playlist bottom-out in NextSong dispatches OnPlaylistExhausted
	// with DurationNext.
	p.NextSong(false)
	if dir, ok := recvCh(o.exhaustedDirs); !ok || dir != DurationNext {
		t.Fatalf("OnPlaylistExhausted dir = %v, ok=%v; want DurationNext", dir, ok)
	}

	// Empty playlist top-out in PreviousSong dispatches OnPlaylistExhausted
	// with DurationPrev.
	p.PreviousSong(false)
	if dir, ok := recvCh(o.exhaustedDirs); !ok || dir != DurationPrev {
		t.Fatalf("OnPlaylistExhausted dir = %v, ok=%v; want DurationPrev", dir, ok)
	}

	// emitStartupPhase dispatches OnStartupPhase.
	testEngine().emitStartupPhase(o, StartupPhasePlaylistLoaded)
	if phase, ok := recvCh(o.startupPhases); !ok || phase != StartupPhasePlaylistLoaded {
		t.Fatalf("OnStartupPhase phase = %v, ok=%v; want StartupPhasePlaylistLoaded", phase, ok)
	}

	// No unexpected extra dispatches.
	if _, ok := recvCh(o.rerenderCh); ok {
		t.Fatal("unexpected extra OnRerender dispatch")
	}
	if _, ok := recvCh(o.exhaustedDirs); ok {
		t.Fatal("unexpected extra OnPlaylistExhausted dispatch")
	}
	if _, ok := recvCh(o.startupPhases); ok {
		t.Fatal("unexpected extra OnStartupPhase dispatch")
	}
}
