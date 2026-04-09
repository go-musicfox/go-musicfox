//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	class_NSUserNotificationCenter = objc.GetClass("NSUserNotificationCenter")
	class_NSUserNotification = objc.GetClass("NSUserNotification")
}

var (
	class_NSUserNotificationCenter objc.Class
	class_NSUserNotification       objc.Class
)

var (
	sel_defaultUserNotificationCenter = objc.RegisterName("defaultUserNotificationCenter")
	sel_scheduleNotification          = objc.RegisterName("scheduleNotification:")
	sel_deliverNotification           = objc.RegisterName("deliverNotification:")
	sel_deliveredNotifications        = objc.RegisterName("deliveredNotifications")
	sel_removeDeliveredNotification   = objc.RegisterName("removeDeliveredNotification:")
	sel_setUserNotificationDelegate   = objc.RegisterName("setDelegate:")
	sel_setTitle                      = objc.RegisterName("setTitle:")
	sel_setSubtitle                   = objc.RegisterName("setSubtitle:")
	sel_setInformativeText            = objc.RegisterName("setInformativeText:")
	sel_setSoundName                  = objc.RegisterName("setSoundName:")
	sel_setUserInfo                   = objc.RegisterName("setUserInfo:")
	sel_setContentImage               = objc.RegisterName("setContentImage:")
	sel_userInfo                      = objc.RegisterName("userInfo")
)

type NSUserNotificationCenter struct {
	core.NSObject
}

type NSUserNotification struct {
	core.NSObject
}

func NSUserNotificationCenter_defaultCenter() NSUserNotificationCenter {
	return NSUserNotificationCenter{
		NSObject: core.NSObject{
			ID: objc.ID(class_NSUserNotificationCenter).Send(sel_defaultUserNotificationCenter),
		},
	}
}

func NewNSUserNotification() NSUserNotification {
	return NSUserNotification{
		NSObject: core.NSObject{
			ID: objc.ID(class_NSUserNotification).Send(sel_alloc).Send(sel_init),
		},
	}
}

func (c NSUserNotificationCenter) SetDelegate(delegate objc.ID) {
	c.Send(sel_setUserNotificationDelegate, delegate)
}

func (c NSUserNotificationCenter) ScheduleNotification(notification NSUserNotification) {
	c.Send(sel_scheduleNotification, notification.ID)
}

func (c NSUserNotificationCenter) DeliverNotification(notification NSUserNotification) {
	c.Send(sel_deliverNotification, notification.ID)
}

func (n NSUserNotification) SetTitle(title string) {
	titleNS := core.String(title)
	n.Send(sel_setTitle, titleNS.ID)
}

func (n NSUserNotification) SetSubtitle(subtitle string) {
	subtitleNS := core.String(subtitle)
	n.Send(sel_setSubtitle, subtitleNS.ID)
}

func (n NSUserNotification) SetInformativeText(text string) {
	textNS := core.String(text)
	n.Send(sel_setInformativeText, textNS.ID)
}

func (n NSUserNotification) SetSoundName(soundName string) {
	soundNS := core.String(soundName)
	n.Send(sel_setSoundName, soundNS.ID)
}

func (n NSUserNotification) SetSoundNameNil() {
	n.Send(sel_setSoundName, 0)
}

func (n NSUserNotification) SetUserInfo(userInfoID objc.ID) {
	n.Send(sel_setUserInfo, userInfoID)
}

func (n NSUserNotification) SetContentImage(imagePath string) {
	imagePathNS := core.String(imagePath)
	imageURL := core.NSURL_fileURLWithPath(imagePathNS)
	nsImage := NSImage_alloc().InitWithContentsOfURL(imageURL)
	n.Send(sel_setContentImage, nsImage.ID)
}
