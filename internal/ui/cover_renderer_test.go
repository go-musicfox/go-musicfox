package ui

import (
	"testing"
	"time"
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
