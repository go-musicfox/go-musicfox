//go:build darwin

package cocoa

import (
	"sync"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

var importUserNotificationsOnce sync.Once

func importUserNotificationsFramework() {
	importUserNotificationsOnce.Do(func() {
		_, err := purego.Dlopen("/System/Library/Frameworks/UserNotifications.framework/UserNotifications", purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
	})
}

var (
	class_UNUserNotificationCenter          objc.Class
	class_UNMutableNotificationContent      objc.Class
	class_UNNotificationRequest             objc.Class
	class_UNTimeIntervalNotificationTrigger objc.Class
	class_UNNotificationAttachment          objc.Class
)

var (
	sel_currentNotificationCenter              = objc.RegisterName("currentNotificationCenter")
	sel_addNotificationRequest                 = objc.RegisterName("addNotificationRequest:withCompletionHandler:")
	sel_triggerWithTimeIntervalRepeats         = objc.RegisterName("triggerWithTimeInterval:repeats:")
	sel_requestWithIdentifierContentTrigger    = objc.RegisterName("requestWithIdentifier:content:trigger:")
	sel_setValueForKey                         = objc.RegisterName("setValue:forKey:")
	sel_attachmentWithIdentifierURLOptions     = objc.RegisterName("attachmentWithIdentifier:URL:options:error:")
	sel_dictionary                             = objc.RegisterName("dictionary")
	sel_alloc                                  = objc.RegisterName("alloc")
	sel_init                                   = objc.RegisterName("init")
	sel_requestAuthorizationWithOptionsHandler = objc.RegisterName("requestAuthorizationWithOptions:completionHandler:")
	sel_authorizationStatus                    = objc.RegisterName("authorizationStatus")
)

const (
	UNAuthorizationOptionBadge = 1 << 0
	UNAuthorizationOptionSound = 1 << 1
	UNAuthorizationOptionAlert = 1 << 2
)

const (
	UNAuthorizationStatusNotDetermined = 0
	UNAuthorizationStatusDenied        = 1
	UNAuthorizationStatusAuthorized    = 2
)

const UNNotificationSoundDefault = 0

func init() {
	importUserNotificationsFramework()
	class_UNUserNotificationCenter = objc.GetClass("UNUserNotificationCenter")
	class_UNMutableNotificationContent = objc.GetClass("UNMutableNotificationContent")
	class_UNNotificationRequest = objc.GetClass("UNNotificationRequest")
	class_UNTimeIntervalNotificationTrigger = objc.GetClass("UNTimeIntervalNotificationTrigger")
	class_UNNotificationAttachment = objc.GetClass("UNNotificationAttachment")
}

func UNUserNotificationCenter_currentNotificationCenter() UNUserNotificationCenter {
	return UNUserNotificationCenter{
		NSObject: core.NSObject{
			ID: objc.ID(class_UNUserNotificationCenter).Send(sel_currentNotificationCenter),
		},
	}
}

type UNUserNotificationCenter struct {
	core.NSObject
}

func (c UNUserNotificationCenter) AddNotificationRequest(request UNNotificationRequest) {
	c.Send(sel_addNotificationRequest, request.ID, 0)
}

func (c UNUserNotificationCenter) RequestAuthorization(options int) {
	c.Send(sel_requestAuthorizationWithOptionsHandler, options, 0)
}

func UNMutableNotificationContent_new() UNMutableNotificationContent {
	return UNMutableNotificationContent{
		NSObject: core.NSObject{
			ID: objc.ID(class_UNMutableNotificationContent).Send(sel_alloc),
		},
	}
}

type UNMutableNotificationContent struct {
	core.NSObject
}

func (c UNMutableNotificationContent) SetTitle(title string) {
	nsTitle := core.String(title)
	c.Send(sel_setValueForKey, nsTitle.ID, core.String("title").ID)
}

func (c UNMutableNotificationContent) SetBody(body string) {
	nsBody := core.String(body)
	c.Send(sel_setValueForKey, nsBody.ID, core.String("body").ID)
}

func (c UNMutableNotificationContent) SetSound(soundName string) {
	nsSound := core.String(soundName)
	c.Send(sel_setValueForKey, nsSound.ID, core.String("sound").ID)
}

func (c UNMutableNotificationContent) SetValueForKey(value core.NSObject, key core.NSObject) {
	c.Send(sel_setValueForKey, value.ID, key.ID)
}

type UNNotificationRequest struct {
	core.NSObject
}

func UNNotificationRequest_requestWithIdentifierContentTrigger(identifier string, content, trigger core.NSObject) UNNotificationRequest {
	nsIdentifier := core.String(identifier)
	return UNNotificationRequest{
		NSObject: core.NSObject{
			ID: objc.ID(class_UNNotificationRequest).Send(sel_requestWithIdentifierContentTrigger, nsIdentifier.ID, content.ID, trigger.ID),
		},
	}
}

type UNTimeIntervalNotificationTrigger struct {
	core.NSObject
}

func UNTimeIntervalNotificationTrigger_triggerWithTimeInterval(timeInterval float64, repeats bool) UNTimeIntervalNotificationTrigger {
	return UNTimeIntervalNotificationTrigger{
		NSObject: core.NSObject{
			ID: objc.ID(class_UNTimeIntervalNotificationTrigger).Send(sel_triggerWithTimeIntervalRepeats, timeInterval, repeats),
		},
	}
}

type UNNotificationAttachment struct {
	core.NSObject
}

func UNNotificationAttachment_attachmentWithIdentifierURL(identifier core.NSString, url core.NSURL, options core.NSDictionary) UNNotificationAttachment {
	return UNNotificationAttachment{
		NSObject: core.NSObject{
			ID: objc.ID(class_UNNotificationAttachment).Send(sel_attachmentWithIdentifierURLOptions, identifier.ID, url.ID, options.ID, 0),
		},
	}
}

func NSDictionary_empty() core.NSDictionary {
	class_NSMutableDictionary := objc.GetClass("NSMutableDictionary")
	return core.NSDictionary{NSObject: core.NSObject{ID: objc.ID(class_NSMutableDictionary).Send(sel_alloc).Send(sel_init)}}
}
