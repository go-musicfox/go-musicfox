package app

import (
	"testing"
	"time"

	"github.com/mattn/go-runewidth"
)

// fakeClock returns a new XScrollBar driven by a manually advanced clock.
func fakeClock(now *time.Time) *XScrollBar {
	bar := NewXScrollBar()
	bar.nowFn = func() time.Time { return *now }
	return bar
}

const longContent = "这是一段比视口宽得多的歌词文字，用来测试横向滚动"

// renderedAt returns the scroll output for a given cell offset.
func renderedAt(offset, width int) string {
	return runewidth.Truncate(runewidth.FillRight(runewidth.TruncateLeft(longContent, offset, ""), width), width, "")
}

func TestXScrollBarShowsStartDuringDwell(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	bar := fakeClock(&now)

	// Less than 600ms (3 cells at 5 cells/s): still at the start.
	now = now.Add(100 * time.Millisecond)
	out := bar.Tick(10, longContent)
	if out != renderedAt(0, 10) {
		t.Fatalf("during start dwell got %q, want content start", out)
	}
}

func TestXScrollBarSpeedIndependentOfFrameRate(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Simulate 1.2s at 5 FPS (6 frames of 200ms) and at 60 FPS (72 frames of
	// 16.67ms). The expected scroll offset after 1.2s is 6 cells; float
	// accumulation may land one cell early or late, so allow a ±1 window.
	inWindow := func(out string) bool {
		for off := 2; off <= 5; off++ {
			if out == renderedAt(off, 10) {
				return true
			}
		}
		return false
	}

	at5 := fakeClock(&now)
	for range 6 {
		now = now.Add(200 * time.Millisecond)
		at5.Tick(10, longContent)
	}
	out5 := at5.Tick(10, longContent)

	now = now.Add(-6 * 200 * time.Millisecond) // rewind to the same origin
	at60 := fakeClock(&now)
	for range 72 {
		now = now.Add(16667 * time.Microsecond)
		at60.Tick(10, longContent)
	}
	out60 := at60.Tick(10, longContent)

	if !inWindow(out5) || !inWindow(out60) {
		t.Fatalf("scroll position out of expected window after 1.2s:\n  5 FPS: %q\n 60 FPS: %q", out5, out60)
	}
	if out5 == renderedAt(0, 10) && out60 == renderedAt(0, 10) {
		t.Fatal("scroll never advanced after 1.2s")
	}
}

func TestXScrollBarFreezesDuringLongGap(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	bar := fakeClock(&now)

	// Advance a little, then simulate a 10s gap (playback paused) — the
	// scroll position must not jump by the gap.
	now = now.Add(1 * time.Second)
	before := bar.Tick(10, longContent)

	now = now.Add(10 * time.Second)
	after := bar.Tick(10, longContent)

	if before != after {
		t.Fatalf("scroll jumped across a long gap:\n before: %q\n  after: %q", before, after)
	}
}

func TestXScrollBarResetsOnContentChange(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	bar := fakeClock(&now)

	now = now.Add(2 * time.Second)
	bar.Tick(10, longContent) // scrolled into the middle

	other := "另一句歌词"
	out := bar.Tick(10, other)
	if want := runewidth.Truncate(runewidth.FillRight(other, 10), 10, ""); out != want {
		t.Fatalf("after content change got %q, want %q", out, want)
	}
}
