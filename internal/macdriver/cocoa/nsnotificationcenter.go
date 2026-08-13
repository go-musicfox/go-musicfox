//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSNotificationCenter = objc.GetClass("NSNotificationCenter")
}

var (
	class_NSNotificationCenter objc.Class
)

var (
	sel_addObserverSelectorNameObject    = objc.RegisterName("addObserver:selector:name:object:")
	sel_removeObserverSelectorNameObject = objc.RegisterName("removeObserver:name:object:")
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

// RemoveObserverNameObject 按 observer/name/object 移除观察者。
// NSNotificationCenter 没有带 selector 的移除方法（removeObserver:name:object:
// 仅有 3 个参数），与 AddObserverSelectorNameObject 的 5 参数消息不同。
func (c NSNotificationCenter) RemoveObserverNameObject(observer objc.ID, name core.NSString, object core.NSObject) {
	c.Send(sel_removeObserverSelectorNameObject, observer, name.ID, object.ID)
}
