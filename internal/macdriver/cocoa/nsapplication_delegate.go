//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

var (
	class_DefaultAppDelegate                            objc.Class
	sel_applicationDidFinishLaunching                   = objc.RegisterName("applicationDidFinishLaunching:")
	sel_applicationShouldTerminateAfterLastWindowClosed = objc.RegisterName("applicationShouldTerminateAfterLastWindowClosed:")

	defaultDelegate *defaultAppDelegate
)

func init() {
	importFramework()

	var err error
	class_DefaultAppDelegate, err = objc.RegisterClass(
		"NSDefaultAppDelegate",
		objc.GetClass("NSObject"),
		[]*objc.Protocol{objc.GetProtocol("NSApplicationDelegate")},
		[]objc.FieldDef{},
		[]objc.MethodDef{
			{
				Cmd: sel_applicationDidFinishLaunching,
				Fn:  applicationDidFinishLaunching,
			},
			{
				Cmd: sel_applicationShouldTerminateAfterLastWindowClosed,
				Fn:  applicationShouldTerminateAfterLastWindowClosed,
			},
		},
	)
	if err != nil {
		panic(err)
	}
}

func applicationDidFinishLaunching(id objc.ID, cmd objc.SEL, notification objc.ID) {
	if defaultDelegate != nil && defaultDelegate.didFinishLaunchingCallback != nil {
		defaultDelegate.didFinishLaunchingCallback(notification)
	}
}

func applicationShouldTerminateAfterLastWindowClosed(id objc.ID, cmd objc.SEL, notification objc.ID) bool {
	return true
}

type defaultAppDelegate struct {
	core.NSObject
	didFinishLaunchingCallback func(notification objc.ID)
}

func (d *defaultAppDelegate) RegisterDidFinishLaunchingCallback(cb func(notification objc.ID)) {
	d.didFinishLaunchingCallback = cb
}

func (d *defaultAppDelegate) Delegate() objc.ID {
	return d.ID
}

func DefaultAppDelegate() *defaultAppDelegate {
	defaultDelegate = &defaultAppDelegate{
		NSObject: core.NSObject{
			ID: objc.ID(class_DefaultAppDelegate).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init),
		},
	}
	return defaultDelegate
}
