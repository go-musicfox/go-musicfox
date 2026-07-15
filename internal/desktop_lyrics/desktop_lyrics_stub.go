//go:build !darwin

package desktop_lyrics

import "github.com/go-musicfox/go-musicfox/internal/configs"

// newController returns nil on non-macOS platforms.
func newController(cfg configs.DesktopLyricsConfig) Controller {
	return nil
}
