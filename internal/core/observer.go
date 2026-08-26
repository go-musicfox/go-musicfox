package core

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// Observer is the frontend-facing callback seam. The engine emits playback
// events; each frontend implements the methods it cares about (TUI rerenders,
// headless no-ops). Nil-safe: the engine checks for nil observers.
//
// Only the three playback-required methods below are mandatory; frontends that
// need the optional extension events implement the corresponding optional
// interfaces defined next to this type (LoginRequester,
// PlaylistExhaustedObserver, RerenderObserver, StartupPhaseObserver). The
// engine dispatches the optional events via assertion + dispatch.
type Observer interface {
	// OnSongChanged reports a song that started playing.
	OnSongChanged(song structs.Song)
	// OnStateChanged reports a playback state transition (Playing/Paused/...).
	OnStateChanged(state types.State)
	// OnPosition is the already-throttled per-tick position (feeds the TUI
	// render ticker).
	OnPosition(d time.Duration)
}

// LoginRequester is the optional login-gating observer: frontends with a login
// page implement it to gate playback on authentication.
type LoginRequester interface {
	RequestLogin(afterLogin func())
}

// PlaylistExhaustedObserver is the optional bottom/top-of-playlist observer:
// the TUI flips the page / runs menu hooks when the playlist end is reached;
// headless simply stops and does not implement it.
type PlaylistExhaustedObserver interface {
	OnPlaylistExhausted(dir PlayDirection)
}

// RerenderObserver is the optional CtrlRerender handler.
type RerenderObserver interface {
	OnRerender()
}

// StartupPhaseObserver is the optional startup-sequence milestone observer.
// The TUI refreshes titles and rerenders at these points.
type StartupPhaseObserver interface {
	OnStartupPhase(phase StartupPhase)
}

// LoadingIndicator is a nil-safe playback-loading indicator (the TUI shows a
// loading tip, headless ignores it).
type LoadingIndicator interface {
	Start()
	Complete()
}

// SongLocator is a nil-safe hook to locate the playing song in the frontend
// view (the TUI scrolls the menu to the playing row).
type SongLocator interface {
	LocatePlayingSong()
}
