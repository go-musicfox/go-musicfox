// Package headless implements the no-TUI frontend: it runs the core engine
// (startup, remote control, notifications, playback) without any bubbletea /
// foxful-cli rendering. It must never import internal/ui.
package headless

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// HeadlessObserver implements core.Observer with no-ops. The headless frontend
// has no rendering: playback events and position ticks are ignored (playback
// simply stops at the playlist end), and login gating cannot reach a login
// page.
//
// It intentionally implements only the three playback-required Observer
// events and does NOT implement the optional extension interfaces
// (core.LoginRequester, core.PlaylistExhaustedObserver, core.RerenderObserver,
// core.StartupPhaseObserver): headless has no UI to rerender, flip pages, or
// refresh titles, and login gating cannot reach a login page. The engine's
// assertion + dispatch skips these events for headless observers.
type HeadlessObserver struct{}

// Compile-time assertion that HeadlessObserver satisfies core.Observer.
var _ core.Observer = HeadlessObserver{}

// OnSongChanged ignores the newly playing song (no UI to update).
func (HeadlessObserver) OnSongChanged(structs.Song) {}

// OnStateChanged ignores playback state transitions (no UI to update).
func (HeadlessObserver) OnStateChanged(types.State) {}

// OnPosition ignores the throttled position ticks (no render ticker).
func (HeadlessObserver) OnPosition(time.Duration) {}
