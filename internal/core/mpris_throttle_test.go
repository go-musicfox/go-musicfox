package core

import (
	"testing"
	"time"
)

func TestMprisPositionThrottleRateLimits(t *testing.T) {
	var throttle mprisPositionThrottle
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// First call always emits.
	if !throttle.shouldEmit(time.Second, now) {
		t.Fatal("first call should emit")
	}
	// Same second: suppressed.
	if throttle.shouldEmit(2*time.Second, now.Add(100*time.Millisecond)) {
		t.Fatal("call within 500ms should be suppressed")
	}
	// After 500ms: emitted.
	if !throttle.shouldEmit(3*time.Second, now.Add(500*time.Millisecond)) {
		t.Fatal("call after 500ms should emit")
	}
}

func TestMprisPositionThrottleEmitsOnJump(t *testing.T) {
	var throttle mprisPositionThrottle
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	throttle.shouldEmit(time.Second, now)
	// A seek forward (> 10s jump) bypasses the interval.
	if !throttle.shouldEmit(30*time.Second, now.Add(50*time.Millisecond)) {
		t.Fatal("large forward jump should emit immediately")
	}
	// A seek backward (> 10s) also bypasses the interval.
	if !throttle.shouldEmit(5*time.Second, now.Add(60*time.Millisecond)) {
		t.Fatal("large backward jump should emit immediately")
	}
}
