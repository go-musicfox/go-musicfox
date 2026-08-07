package ui

import (
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/player"
)

func TestSpectrogramAdvanceScrollIsTimeBased(t *testing.T) {
	// speed=4 means 4 columns per 200ms → 20 columns per second regardless of
	// the frame rate. Simulate one wall-clock second at 5 FPS and at 60 FPS.
	const (
		speed = 4
		width = 200
	)

	totalAt5FPS := 0
	r5 := &SpectrogramRenderer{}
	for range 5 {
		totalAt5FPS += r5.advanceScroll(200*time.Millisecond, 200*time.Millisecond, speed, width)
	}

	totalAt60FPS := 0
	r60 := &SpectrogramRenderer{}
	for range 60 {
		totalAt60FPS += r60.advanceScroll(16667*time.Microsecond, 16*time.Millisecond, speed, width)
	}

	if totalAt5FPS != 20 {
		t.Fatalf("scrolled %d columns in 1s at 5 FPS, want 20", totalAt5FPS)
	}
	if totalAt60FPS != 20 {
		t.Fatalf("scrolled %d columns in 1s at 60 FPS, want 20", totalAt60FPS)
	}
}

func TestSpectrogramAdvanceScrollFreezesOnLongGap(t *testing.T) {
	r := &SpectrogramRenderer{}
	if got := r.advanceScroll(10*time.Second, 200*time.Millisecond, 4, 200); got != 0 {
		t.Fatalf("scrolled %d columns after a 10s pause, want 0 (history frozen)", got)
	}
}

func TestSpectrogramAdvanceScrollCapsAtWidth(t *testing.T) {
	r := &SpectrogramRenderer{}
	if got := r.advanceScroll(200*time.Millisecond, 200*time.Millisecond, 4, 3); got != 3 {
		t.Fatalf("scrolled %d columns, want capped at width 3", got)
	}
}

func TestSpectrogramScrollAndDrawMovesHistory(t *testing.T) {
	r := &SpectrogramRenderer{}
	height, width := 2, 8
	// View allocates the history buffer before scrolling; mirror that here.
	r.history = make([][]byte, height)
	for i := range r.history {
		r.history[i] = make([]byte, width)
	}

	// First frame: stamp data into the last column.
	frame := player.SpectrumFrame{}
	frame.Levels[0] = 1
	r.scrollAndDraw(frame, height, width, 1)
	if r.history[0][width-1] == 0 {
		t.Fatal("first frame was not stamped into the rightmost column")
	}

	// A full-width scroll clears everything and stamps the new frame at the right.
	r.scrollAndDraw(player.SpectrumFrame{}, height, width, width)
	for row := 0; row < height; row++ {
		for col := 0; col < width-1; col++ {
			if r.history[row][col] != 0 {
				t.Fatalf("history[%d][%d] = %d, want 0 after full-width scroll", row, col, r.history[row][col])
			}
		}
	}
}
