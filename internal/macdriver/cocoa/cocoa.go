//go:build darwin

package cocoa

import (
	"sync"

	"github.com/ebitengine/purego"
)

var importOnce sync.Once

func importFramework() {
	importOnce.Do(func() {
		_, err := purego.Dlopen("/System/Library/Frameworks/AppKit.framework/AppKit", purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
	})
}

type CGFloat = float64

type CGSize struct {
	Width, Height CGFloat
}

type CGPoint struct {
	X, Y float64
}

type CGRect struct {
	Origin CGPoint
	Size   CGSize
}

type (
	NSPoint = CGPoint
	NSRect  = CGRect
	NSSize  = CGSize
)
