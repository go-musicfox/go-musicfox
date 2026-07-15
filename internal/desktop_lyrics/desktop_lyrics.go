package desktop_lyrics

import "github.com/go-musicfox/go-musicfox/internal/configs"

// LyricWord represents a single word with its timing in milliseconds.
type LyricWord struct {
	Word      string
	StartTime int64 // ms
	EndTime   int64 // ms
}

// LyricLine represents a line of lyrics for desktop display.
// When Words is nil or empty, the line is rendered as plain Text.
// When Words is populated (YRC mode), each word is colored individually
// based on currentTimeMs: played words get the progress color, unplayed
// words are dimmed, and the currently-playing word is interpolated.
type LyricLine struct {
	Text  string
	Words []LyricWord
}

// Controller is the cross-platform interface for desktop lyrics.
type Controller interface {
	Show()
	Hide()
	IsVisible() bool
	// Update refreshes the desktop lyrics display.
	// currentTimeMs is the current playback position in milliseconds,
	// used for word-by-word coloring in YRC mode.
	Update(curLine, nextLine LyricLine, currentIndex int, currentTimeMs int64)
	Close()
}

// NewController creates a platform-specific desktop lyrics controller.
// Returns nil on unsupported platforms or when disabled.
func NewController(cfg configs.DesktopLyricsConfig) Controller {
	return newController(cfg)
}
