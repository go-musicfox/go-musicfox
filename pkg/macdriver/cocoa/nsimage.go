//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func init() {
	importFramework()
	class_NSImage = objc.GetClass("NSImage")
}

var (
	class_NSImage objc.Class
)

var (
	sel_initWithContentsOfURL = objc.RegisterName("initWithContentsOfURL:")
)

type NSImage struct {
	core.NSObject
}

func NSImage_alloc() NSImage {
	return NSImage{
		core.NSObject{
			ID: objc.ID(class_NSImage).Send(macdriver.SEL_alloc),
		},
	}
}

func (i NSImage) InitWithContentsOfURL(url core.NSURL) NSImage {
	i.ID = i.Send(sel_initWithContentsOfURL, url.ID)
	return i
}
