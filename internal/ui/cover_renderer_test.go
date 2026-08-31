package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/ui/kitty"
)

func TestPlaceBackoffTransitions(t *testing.T) {
	r := &CoverRenderer{}
	now := time.Now()

	if r.placeBackoffActive(now) {
		t.Fatal("backoff should be inactive initially")
	}

	// First failure: 1s window.
	r.recordPlaceFailure(now)
	if !r.placeBackoffActive(now) || !r.placeBackoffActive(now.Add(999*time.Millisecond)) {
		t.Fatal("expected an active 1s backoff window after the first failure")
	}
	if r.placeBackoffActive(now.Add(time.Second)) {
		t.Fatal("backoff should expire after 1s")
	}

	// Consecutive failures double the window: 2s, then 4s.
	secondFailure := now.Add(time.Second)
	r.recordPlaceFailure(secondFailure)
	if !r.placeBackoffActive(secondFailure.Add(2*time.Second - time.Millisecond)) {
		t.Fatal("expected an active 2s backoff window after the second failure")
	}
	if r.placeBackoffActive(secondFailure.Add(2 * time.Second)) {
		t.Fatal("second backoff window should expire after 2s")
	}

	thirdFailure := secondFailure.Add(2 * time.Second)
	r.recordPlaceFailure(thirdFailure)
	if !r.placeBackoffActive(thirdFailure.Add(4*time.Second - time.Millisecond)) {
		t.Fatal("expected an active 4s backoff window after the third failure")
	}
	if r.placeBackoffActive(thirdFailure.Add(4 * time.Second)) {
		t.Fatal("third backoff window should expire after 4s")
	}

	// Repeated failures cap the window at 30s.
	failureTime := thirdFailure.Add(4 * time.Second)
	for range 10 {
		r.recordPlaceFailure(failureTime)
	}
	if r.placeBackoff != placeBackoffMax {
		t.Fatalf("backoff should be capped at %v, got %v", placeBackoffMax, r.placeBackoff)
	}
	if !r.placeBackoffActive(failureTime.Add(placeBackoffMax - time.Millisecond)) {
		t.Fatal("capped backoff window should still be active just before expiry")
	}
	if r.placeBackoffActive(failureTime.Add(placeBackoffMax)) {
		t.Fatal("capped backoff window should expire")
	}

	// Success resets the backoff entirely.
	r.recordPlaceSuccess()
	if r.placeBackoffActive(time.Now()) || r.placeBackoff != 0 || !r.placeFailAt.IsZero() {
		t.Fatal("success should reset the backoff state")
	}
}

// stripTmuxWrapped removes every DCS tmux; passthrough packet from s,
// unwrapping ESC-doubled payloads with the same rules tmux uses: ESC ESC
// inside the payload is a literal ESC, a lone ESC \ terminates the packet. A
// plain regex is not enough because the doubled inner terminators would
// confuse a non-greedy match.
func stripTmuxWrapped(s string) string {
	const dcs = "\x1bPtmux;"
	var b strings.Builder
	for {
		i := strings.Index(s, dcs)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		rest := s[i+len(dcs):]
		j := 0
		for j < len(rest) {
			if rest[j] == 0x1b {
				if j+1 < len(rest) && rest[j+1] == 0x1b {
					j += 2
					continue
				}
				if j+1 < len(rest) && rest[j+1] == '\\' {
					break
				}
			}
			j++
		}
		if j >= len(rest) {
			b.WriteString(rest)
			return b.String()
		}
		s = rest[j+2:]
	}
}

// TestBuildAnimationSequenceTmuxWrapsAllAPCs asserts that in tmux passthrough
// mode every Kitty APC of the assembled animation sequence — including the
// animation control commands SetFrameGap/StartAnimation, not just the
// placement payload — travels inside the DCS passthrough packet: tmux
// consumes bare APCs as pane titles (dropping or polluting them), so the
// animation would silently never start. Tests mutate the kitty package's
// global passthrough mode and must not run in parallel.
func TestBuildAnimationSequenceTmuxWrapsAllAPCs(t *testing.T) {
	kitty.SetTmuxPassthroughForTest(true)
	t.Cleanup(func() { kitty.SetTmuxPassthroughForTest(false) })

	const (
		animID   = 7
		oldAnimD = 5
		row      = 5
		col      = 1
	)
	seq, ok := buildAnimationSequence(animID, oldAnimD, 33, row, col, 20, 1, 0, 74)
	if !ok {
		t.Fatal("expected the placement to succeed")
	}

	if n := strings.Count(seq, "\x1bPtmux;"); n != 1 {
		t.Fatalf("expected exactly one DCS passthrough packet, got %d", n)
	}
	rest := stripTmuxWrapped(seq)
	if strings.Contains(rest, "\x1b_G") {
		t.Fatalf("found a bare Kitty APC outside the passthrough packet: %q", rest)
	}
	// The wrapped payload must carry the outer-terminal absolute CUP
	// (pane top 0 + row 5, pane left 74 + col 1) and the cursor save/restore.
	if !strings.Contains(seq, "\x1b[5;75H") || !strings.Contains(seq, "\x1b7") || !strings.Contains(seq, "\x1b8") {
		t.Fatalf("expected absolute CUP (row 5, col 75) with cursor save/restore in the payload: %q", seq)
	}
}
