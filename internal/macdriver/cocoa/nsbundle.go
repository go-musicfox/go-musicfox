//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSBundle = objc.GetClass("NSBundle")
}

var (
	class_NSBundle objc.Class
)

var (
	sel_mainBundle = objc.RegisterName("mainBundle")
	sel_bundleURL  = objc.RegisterName("bundleURL")
)

type NSBundle struct {
	core.NSObject
}

func NSBundle_mainBundle() NSBundle {
	return NSBundle{
		NSObject: core.NSObject{
			ID: objc.ID(class_NSBundle).Send(sel_mainBundle),
		},
	}
}

func (b NSBundle) BundleURL() core.NSURL {
	return core.NSURL{
		NSObject: core.NSObject{
			ID: b.Send(sel_bundleURL),
		},
	}
}
