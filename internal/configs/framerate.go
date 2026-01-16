package configs

import (
	"time"
)

// FrameRate represents the playback animation frame rate.
// It provides methods to calculate the time interval between frames.
type FrameRate int

// DefaultFrameRate returns the default frame rate (30 FPS).
func DefaultFrameRate() FrameRate {
	return 30
}

// Interval returns the time duration between frames based on the frame rate.
// For example, 30 FPS returns 33ms, 60 FPS returns 16ms.
func (f FrameRate) Interval() time.Duration {
	if f <= 0 {
		return DefaultFrameRate().Interval()
	}
	return time.Duration(int(time.Second.Milliseconds())/int(f)) * time.Millisecond
}

// DurationMs returns the interval in milliseconds.
func (f FrameRate) DurationMs() int {
	if f <= 0 {
		return int(DefaultFrameRate().Interval().Milliseconds())
	}
	return int(time.Second.Milliseconds()) / int(f)
}

// String returns the string representation of the frame rate (e.g., "30 FPS").
func (f FrameRate) String() string {
	return string(rune(f+'0')) + " FPS"
}
