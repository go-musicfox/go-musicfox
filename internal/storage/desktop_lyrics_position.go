package storage

import "github.com/go-musicfox/go-musicfox/internal/types"

// DesktopLyricsPosition stores the desktop lyrics window origin relative to its display.
type DesktopLyricsPosition struct {
	ScreenID uint32  `json:"screen_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
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
