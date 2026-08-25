package core

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// Observer is the frontend-facing callback seam. The engine emits playback
// events; each frontend implements the methods it cares about (TUI rerenders,
// headless no-ops). Nil-safe: the engine checks for nil observers.
type Observer interface {
	// OnSongChanged reports a song that started playing.
	OnSongChanged(song structs.Song)
	// OnStateChanged reports a playback state transition (Playing/Paused/...).
	OnStateChanged(state types.State)
	// OnPosition is the already-throttled per-tick position (feeds the TUI
	// render ticker).
	OnPosition(d time.Duration)
	// RequestLogin is reserved for login gating.
	RequestLogin(afterLogin func())
	// OnPlaylistExhausted is called when the playlist bottom/top is reached:
	// the TUI flips the page / runs menu hooks; headless stops.
	OnPlaylistExhausted(dir PlayDirection)
	// OnRerender handles CtrlRerender.
	OnRerender()
	// OnStartupPhase reports the engine startup sequence reaching a phase
	// milestone (user restored / playlist state loaded / right before
	// autoplay). The TUI refreshes titles and rerenders at these points.
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
