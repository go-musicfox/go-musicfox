package kitty

import (
	"strings"
	"testing"
	"time"
)

func TestParseTmuxPaneGeometry(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		wantTop  int
		wantLeft int
		wantOK   bool
	}{
		{"normal", "3,74", 3, 74, true},
		{"zeroes", "0,0", 0, 0, true},
		{"with spaces", " 3 , 74 ", 3, 74, true},
		{"trailing newline", "3,74\n", 3, 74, true},
		{"crlf", "3,74\r\n", 3, 74, true},
		{"non numeric", "abc,74", 0, 0, false},
		{"missing column", "3", 0, 0, false},
		{"extra column", "3,74,1", 0, 0, false},
		{"negative", "-1,74", 0, 0, false},
		{"empty", "", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			top, left, ok := parseTmuxPaneGeometry(tt.output)
			if ok != tt.wantOK {
				t.Fatalf("parseTmuxPaneGeometry(%q) ok = %v, want %v", tt.output, ok, tt.wantOK)
			}
			if ok && (top != tt.wantTop || left != tt.wantLeft) {
				t.Errorf("parseTmuxPaneGeometry(%q) = (%d, %d), want (%d, %d)", tt.output, top, left, tt.wantTop, tt.wantLeft)
			}
		})
	}
}

func TestBuildTmuxPositionedPayload(t *testing.T) {
	// pane_top=0, pane_left=74, in-pane start row 5 / col 1 (1-based) must map
	// to outer-terminal absolute row 5 / col 75.
	payload := BuildTmuxPositionedPayload(0, 74, 5, 1, "IMGSEQ", false)
	want := "\x1b7\x1b[5;75HIMGSEQ\x1b8"
	if payload != want {
		t.Errorf("payload = %q, want %q", payload, want)
	}

	// deleteOld prefixes the delete-all APC command.
	withDelete := BuildTmuxPositionedPayload(2, 0, 1, 1, "IMGSEQ", true)
	wantDelete := "\x1b_Ga=d,d=a,q=2\x1b\\\x1b7\x1b[3;1HIMGSEQ\x1b8"
	if withDelete != wantDelete {
		t.Errorf("payload with delete = %q, want %q", withDelete, wantDelete)
	}

	// The payload is bare (no DCS tmux; wrapper); callers wrap it exactly once.
	setTmuxPassthrough(t, true)
	if strings.Contains(payload, "\x1bPtmux;") {
		t.Errorf("payload must be bare, got %q", payload)
	}
	wrapped := Wrap(payload)
	if !strings.HasPrefix(wrapped, "\x1bPtmux;") || !strings.HasSuffix(wrapped, "\x1b\\") {
		t.Errorf("wrapped payload = %q, want DCS tmux; passthrough", wrapped)
	}
}

func TestPaneOffsetCacheValidity(t *testing.T) {
	now := time.Now()
	if !successCacheValid(true, now.Add(-time.Second), now) {
		t.Error("fresh successful cache should be valid")
	}
	if successCacheValid(true, now.Add(-2*time.Second), now) {
		t.Error("expired successful cache should be invalid")
	}
	if successCacheValid(false, now, now) {
		t.Error("uncached result should be invalid")
	}
	if !failureCacheValid(now.Add(-500*time.Millisecond), now) {
		t.Error("recent failure should keep the negative cache valid")
	}
	if failureCacheValid(now.Add(-time.Second), now) {
		t.Error("failure older than the TTL should invalidate the negative cache")
	}
	if failureCacheValid(time.Time{}, now) {
		t.Error("zero failure timestamp should not count as a cached failure")
	}
}

func TestInvalidateTmuxPaneOffset(t *testing.T) {
	// Seed both caches, then verify invalidation clears them.
	tmuxPaneOffsetMu.Lock()
	tmuxPaneOffsetCached = true
	tmuxPaneOffsetAt = time.Now()
	tmuxPaneOffsetTop, tmuxPaneOffsetLeft = 3, 7
	tmuxPaneOffsetFailAt = time.Now()
	tmuxPaneOffsetMu.Unlock()
	t.Cleanup(InvalidateTmuxPaneOffset)

	InvalidateTmuxPaneOffset()

	tmuxPaneOffsetMu.Lock()
	defer tmuxPaneOffsetMu.Unlock()
	if tmuxPaneOffsetCached {
		t.Error("expected success cache to be reset")
	}
	if !tmuxPaneOffsetFailAt.IsZero() {
		t.Error("expected failure cache to be reset")
	}
}

func TestTmuxPaneOffsetSuccessCacheHit(t *testing.T) {
	t.Setenv("TMUX", "/tmp/go-musicfox-kitty-test-nonexistent-socket,123,0")
	t.Setenv("TMUX_PANE", "%5")

	// Seed a fresh successful result: TmuxPaneOffset must return it without
	// executing tmux (the seeded socket does not exist, so any exec would
	// fail the test's expectations).
	tmuxPaneOffsetMu.Lock()
	tmuxPaneOffsetCached = true
	tmuxPaneOffsetAt = time.Now()
	tmuxPaneOffsetTop, tmuxPaneOffsetLeft = 3, 7
	tmuxPaneOffsetMu.Unlock()
	t.Cleanup(InvalidateTmuxPaneOffset)

	top, left, ok := TmuxPaneOffset()
	if !ok || top != 3 || left != 7 {
		t.Errorf("expected cached offset (3, 7), got (%d, %d, %v)", top, left, ok)
	}
}

func TestTmuxPaneOffsetNegativeCache(t *testing.T) {
	t.Setenv("TMUX", "/tmp/go-musicfox-kitty-test-nonexistent-socket,123,0")
	t.Setenv("TMUX_PANE", "%5")

	// Seed a fresh failure: the negative cache must report not-ok without
	// re-executing tmux (the socket does not exist, so any exec would fail
	// anyway; the seed only exercises the skip path).
	tmuxPaneOffsetMu.Lock()
	tmuxPaneOffsetFailAt = time.Now()
	tmuxPaneOffsetMu.Unlock()
	t.Cleanup(InvalidateTmuxPaneOffset)

	if _, _, ok := TmuxPaneOffset(); ok {
		t.Error("expected negative cache to report failure without re-query")
	}
}
