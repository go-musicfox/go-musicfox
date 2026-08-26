package headless

import (
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// TestHeadlessObserverZeroValueNoPanic ensures every core.Observer method is
// callable on a zero-value HeadlessObserver without panicking. Only the three
// playback-required events are asserted: HeadlessObserver deliberately does
// not implement the optional extension interfaces (LoginRequester,
// PlaylistExhaustedObserver, RerenderObserver, StartupPhaseObserver) because
// headless has no UI to rerender, flip pages, or refresh titles — the engine's
// assertion + dispatch skips those events.
func TestHeadlessObserverZeroValueNoPanic(t *testing.T) {
	var o HeadlessObserver

	o.OnSongChanged(structs.Song{})
	o.OnStateChanged(types.Playing)
	o.OnPosition(time.Second)
}
