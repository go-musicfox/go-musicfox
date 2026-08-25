package core

import "time"

// mprisPositionThrottle rate-limits MPRIS Position property broadcasts. The
// player emits a position every render tick; without throttling, raising the
// UI frame rate floods the DBus session bus (e.g. 60 updates/sec at 60 FPS).
//
// An update is emitted when at least mprisPositionMinInterval has passed since
// the last one, or immediately when the position jumped by more than
// mprisPositionJumpMin (e.g. after a seek). Not safe for concurrent use —
// callers use it from a single goroutine.
type mprisPositionThrottle struct {
	lastPos time.Duration
	lastAt  time.Time
}

const (
	mprisPositionMinInterval = 500 * time.Millisecond
	mprisPositionJumpMin     = 10 * time.Second
)

// shouldEmit reports whether the new position should be broadcast, updating
// the throttle state on a hit.
func (t *mprisPositionThrottle) shouldEmit(pos time.Duration, now time.Time) bool {
	delta := pos - t.lastPos
	if delta < 0 {
		delta = -delta
	}
	if now.Sub(t.lastAt) >= mprisPositionMinInterval || delta > mprisPositionJumpMin {
		t.lastPos = pos
		t.lastAt = now
		return true
	}
	return false
}
