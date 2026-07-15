package desktop_lyrics

import "github.com/go-musicfox/go-musicfox/internal/configs"

// Controller is the cross-platform interface for desktop lyrics.
type Controller interface {
	Show()
	Hide()
	IsVisible() bool
	// Update refreshes the desktop lyrics display.
	// currentLine: the line currently being sung.
	// nextLine: the upcoming line.
	// currentIndex: the zero-based index of the currentLine in the lyrics.
	//   Even indices (0,2,4,...) place the active line at the first position (top-left);
	//   odd indices (1,3,5,...) place the active line at the second position (bottom-right).
	Update(currentLine, nextLine string, currentIndex int)
	Close()
}

// NewController creates a platform-specific desktop lyrics controller.
// Returns nil on unsupported platforms or when disabled.
func NewController(cfg configs.DesktopLyricsConfig) Controller {
	return newController(cfg)
}
