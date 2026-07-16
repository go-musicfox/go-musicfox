//go:build darwin

package desktop_lyrics

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/storage"
)

func TestDesktopLyricsWindowOriginUsesStoredPosition(t *testing.T) {
	defaultOrigin := cocoa.CGPoint{X: 100, Y: 100}
	screenFrame := cocoa.NSRect{
		Origin: cocoa.CGPoint{X: 1920, Y: 0},
		Size:   cocoa.CGSize{Width: 2560, Height: 1440},
	}
	position := storage.DesktopLyricsPosition{ScreenID: 42, X: 320, Y: 240}

	got := desktopLyricsWindowOrigin(defaultOrigin, screenFrame, &position)
	want := cocoa.CGPoint{X: 2240, Y: 240}
	if got != want {
		t.Fatalf("desktopLyricsWindowOrigin() = %#v, want %#v", got, want)
	}
}

func TestDesktopLyricsPositionFromFrame(t *testing.T) {
	screenFrame := cocoa.NSRect{
		Origin: cocoa.CGPoint{X: -2560, Y: 0},
		Size:   cocoa.CGSize{Width: 2560, Height: 1440},
	}
	windowFrame := cocoa.NSRect{
		Origin: cocoa.CGPoint{X: -2240, Y: 240},
		Size:   cocoa.CGSize{Width: 400, Height: 80},
	}

	got := desktopLyricsPositionFromFrame(windowFrame, screenFrame, 42)
	want := storage.DesktopLyricsPosition{ScreenID: 42, X: 320, Y: 240}
	if got != want {
		t.Fatalf("desktopLyricsPositionFromFrame() = %#v, want %#v", got, want)
	}
}
