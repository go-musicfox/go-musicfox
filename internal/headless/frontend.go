// Package headless implements the no-TUI frontend: it runs the core engine
// (startup, remote control, notifications, playback) without any bubbletea /
// foxful-cli rendering. It must never import internal/ui.
package headless

import (
	"log/slog"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// HeadlessObserver implements core.Observer with no-ops. The headless frontend
// has no rendering: playback events, position ticks and playlist-exhausted
// callbacks are ignored (playback simply stops at the playlist end), and
// login gating degrades to a log line instead of a login page.
type HeadlessObserver struct{}

// Compile-time assertion that HeadlessObserver satisfies core.Observer.
var _ core.Observer = HeadlessObserver{}

// OnSongChanged ignores the newly playing song (no UI to update).
func (HeadlessObserver) OnSongChanged(structs.Song) {}

// OnStateChanged ignores playback state transitions (no UI to update).
func (HeadlessObserver) OnStateChanged(types.State) {}

// OnPosition ignores the throttled position ticks (no render ticker).
func (HeadlessObserver) OnPosition(time.Duration) {}

// RequestLogin degrades login gating to a log line: there is no login page in
// headless mode, so the afterLogin callback is never invoked.
func (HeadlessObserver) RequestLogin(_ func()) {
	slog.Info("headless: login required but no frontend available")
}

// OnPlaylistExhausted ignores bottom/top-of-playlist events: playback simply
// stops at the playlist end in headless mode.
func (HeadlessObserver) OnPlaylistExhausted(core.PlayDirection) {}

// OnRerender ignores the rerender request (nothing to render).
func (HeadlessObserver) OnRerender() {}

// OnStartupPhase ignores startup phase milestones (no UI to refresh).
func (HeadlessObserver) OnStartupPhase(core.StartupPhase) {}
