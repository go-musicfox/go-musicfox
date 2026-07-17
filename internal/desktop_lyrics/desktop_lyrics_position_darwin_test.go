//go:build darwin

package desktop_lyrics

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
)

func TestConstrainWindowOrigin(t *testing.T) {
	ctrl := &darwinController{
		screenFrame: cocoa.NSRect{
			Origin: cocoa.CGPoint{X: 0, Y: 0},
			Size:   cocoa.CGSize{Width: 2560, Height: 1440},
		},
	}

	// Inside bounds – no change.
	x, y := ctrl.constrainWindowOrigin(100, 100, 400, 80)
	if x != 100 || y != 100 {
		t.Fatalf("want (100,100), got (%v,%v)", x, y)
	}

	// Clamp left edge.
	x, y = ctrl.constrainWindowOrigin(-10, 100, 400, 80)
	if x != 0 || y != 100 {
		t.Fatalf("left clamp: want (0,100), got (%v,%v)", x, y)
	}

	// Clamp right edge.
	x, y = ctrl.constrainWindowOrigin(2200, 100, 400, 80)
	if x != 2160 || y != 100 {
		t.Fatalf("right clamp: want (2160,100), got (%v,%v)", x, y)
	}

	// Clamp top edge.
	x, y = ctrl.constrainWindowOrigin(100, 1400, 400, 80)
	if y != 1360 {
		t.Fatalf("top clamp: want y=1360, got y=%v", y)
	}

	// Clamp bottom edge (minimum 4px).
	x, y = ctrl.constrainWindowOrigin(100, 1, 400, 80)
	if y != 4 {
		t.Fatalf("bottom clamp: want y=4, got y=%v", y)
	}
}
