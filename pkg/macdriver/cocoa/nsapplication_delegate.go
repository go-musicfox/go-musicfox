//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func init() {
	importFramework()

	var err error
	class_DefaultAppDelegate, err = objc.RegisterClass(&defaultAppDelegateBinding{})
	if err != nil {
		panic(err)
	}
}

var (
	class_DefaultAppDelegate objc.Class
)

var (
	sel_applicationDidFinishLaunching                   = objc.RegisterName("applicationDidFinishLaunching:")
	sel_applicationShouldTerminateAfterLastWindowClosed = objc.RegisterName("applicationShouldTerminateAfterLastWindowClosed:")
)

var (
	defaultDelegate *defaultAppDelegate
)

type defaultAppDelegateBinding struct {
	isa objc.Class `objc:"NSDefaultAppDelegate : NSObject <NSApplicationDelegate>"`
}

func (defaultAppDelegateBinding) ApplicationDidFinishLaunching(_ objc.SEL, notification objc.ID) {
	if defaultDelegate != nil && defaultDelegate.didFinishLaunchingCallback != nil {
		defaultDelegate.didFinishLaunchingCallback(notification)
	}
}

func (defaultAppDelegateBinding) ApplicationShouldTerminateAfterLastWindowClosed(_ objc.SEL, _ objc.ID) bool {
	return true
}

func (defaultAppDelegateBinding) Selector(metName string) objc.SEL {
	switch metName {
	case "ApplicationDidFinishLaunching":
		return sel_applicationDidFinishLaunching
	case "ApplicationShouldTerminateAfterLastWindowClosed":
		return sel_applicationShouldTerminateAfterLastWindowClosed
	default:
		return 0
	}
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
