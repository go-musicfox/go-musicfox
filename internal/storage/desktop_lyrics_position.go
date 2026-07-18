package storage

import "github.com/go-musicfox/go-musicfox/internal/types"

// DesktopLyricsPosition stores the desktop lyrics window origin relative to its display.
// Position is expressed as both absolute coordinates (backward compat) and
// screen-relative center factors (0–1) following the LyricsX convention.
type DesktopLyricsPosition struct {
	ScreenID uint32  `json:"screen_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	XFactor  float64 `json:"x_factor"` // screen-relative horizontal center factor (0–1)
	YFactor  float64 `json:"y_factor"` // screen-relative vertical center factor (0–1)
}

func (p DesktopLyricsPosition) GetDbName() string {
	return types.AppDBName
}

func (p DesktopLyricsPosition) GetTableName() string {
	return "default_bucket"
}

func (p DesktopLyricsPosition) GetKey() string {
	return "desktop_lyrics_position"
}
