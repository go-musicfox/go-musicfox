//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

var (
	class_NotificationDelegate                              objc.Class
	sel_userNotificationCenterShouldPresentNotification      = objc.RegisterName("userNotificationCenter:shouldPresentNotification:")
	sel_userNotificationCenterDidDeliverNotification        = objc.RegisterName("userNotificationCenter:didDeliverNotification:")
	sel_userNotificationCenterDidActivateNotification       = objc.RegisterName("userNotificationCenter:didActivateNotification:")

	notificationDelegate *userNotificationDelegate
)

func init() {
	importFramework()

	var err error
	class_NotificationDelegate, err = objc.RegisterClass(
		"NSUserNotificationDelegate",
		objc.GetClass("NSObject"),
		[]*objc.Protocol{objc.GetProtocol("NSUserNotificationCenterDelegate")},
		[]objc.FieldDef{},
		[]objc.MethodDef{
			{
				Cmd: sel_userNotificationCenterShouldPresentNotification,
				Fn:  userNotificationCenterShouldPresentNotification,
			},
			{
				Cmd: sel_userNotificationCenterDidDeliverNotification,
				Fn:  userNotificationCenterDidDeliverNotification,
			},
			{
				Cmd: sel_userNotificationCenterDidActivateNotification,
				Fn:  userNotificationCenterDidActivateNotification,
			},
		},
	)
	if err != nil {
		panic(err)
	}
}

func userNotificationCenterShouldPresentNotification(id objc.ID, cmd objc.SEL, center, notification objc.ID) bool {
	return true
}

func userNotificationCenterDidDeliverNotification(id objc.ID, cmd objc.SEL, center, notification objc.ID) {
}

func userNotificationCenterDidActivateNotification(id objc.ID, cmd objc.SEL, center, notification objc.ID) {
	notificationObj := core.NSObject{ID: notification}
	userInfo := notificationObj.Send(core.SEL_valueForKey, core.String("userInfo").ID)
	if userInfo == 0 {
		return
	}

	userInfoObj := core.NSObject{ID: userInfo}
	openUrlObj := userInfoObj.Send(core.SEL_valueForKey, core.String("openUrl").ID)
	if openUrlObj == 0 {
		return
	}

	urlStr := core.NSString{NSObject: core.NSObject{ID: openUrlObj}}.String()
	if urlStr == "" {
		return
	}

	nsUrl := core.NSURL_URLWithString(core.String(urlStr))
	NSWorkspace_sharedWorkspace().OpenURL(nsUrl)
}

type userNotificationDelegate struct {
	core.NSObject
}

func (d *userNotificationDelegate) ID() objc.ID {
	return d.NSObject.ID
}

func NewUserNotificationDelegate() *userNotificationDelegate {
	notificationDelegate = &userNotificationDelegate{
		NSObject: core.NSObject{
			ID: objc.ID(class_NotificationDelegate).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init),
		},
	}
	return notificationDelegate
}
