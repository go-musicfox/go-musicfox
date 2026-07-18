//go:build darwin

package cocoa

import (
	"sync"

	"github.com/ebitengine/purego"
)

var cgOnce sync.Once
var (
	CGMainDisplayID     func() uint32
	CGDisplayPixelsWide func(displayID uint32) uint
	CGDisplayPixelsHigh func(displayID uint32) uint
)

func ImportCoreGraphics() {
	cgOnce.Do(func() {
		lib, err := purego.Dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics", purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
		purego.RegisterLibFunc(&CGMainDisplayID, lib, "CGMainDisplayID")
		purego.RegisterLibFunc(&CGDisplayPixelsWide, lib, "CGDisplayPixelsWide")
		purego.RegisterLibFunc(&CGDisplayPixelsHigh, lib, "CGDisplayPixelsHigh")
	})
}
