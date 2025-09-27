//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSNotificationCenter = objc.GetClass("NSNotificationCenter")
}

var (
	class_NSNotificationCenter objc.Class
)

var (
	sel_addObserverSelectorNameObject = objc.RegisterName("addObserver:selector:name:object:")
)

type NSNotificationCenter struct {
	core.NSObject
}

func NSNotificationCenter_defaultCenter() NSNotificationCenter {
	return NSNotificationCenter{
		core.NSObject{
			ID: objc.ID(class_NSNotificationCenter).Send(macdriver.SEL_defaultCenter),
		},
	}
}

func (c NSNotificationCenter) AddObserverSelectorNameObject(observer objc.ID, selector objc.SEL, name core.NSString, object core.NSObject) {
	c.Send(sel_addObserverSelectorNameObject, observer, selector, name.ID, object.ID)
}
