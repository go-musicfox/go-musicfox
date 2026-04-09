//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSWorkspace = objc.GetClass("NSWorkspace")
}

var (
	class_NSWorkspace objc.Class
)

var (
	sel_sharedWorkspace    = objc.RegisterName("sharedWorkspace")
	sel_notificationCenter = objc.RegisterName("notificationCenter")
	sel_openURL            = objc.RegisterName("openURL:")
)

type NSWorkspace struct {
	core.NSObject
}

func NSWorkspace_sharedWorkspace() NSWorkspace {
	return NSWorkspace{
		core.NSObject{
			ID: objc.ID(class_NSWorkspace).Send(sel_sharedWorkspace),
		},
	}
}

func (w NSWorkspace) NotificationCenter() (nc NSNotificationCenter) {
	nc.SetObjcID(w.Send(sel_notificationCenter))
	return
}

func (w NSWorkspace) OpenURL(url core.NSURL) {
	w.Send(sel_openURL, url.ID)
}
