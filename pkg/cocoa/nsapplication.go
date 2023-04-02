//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/utils"
)

var (
	class_DefaultAppDelegate objc.Class
	class_NSApplication      = objc.GetClass("NSApplication")
)

var (
	sel_applicationDidFinishLaunching                   = objc.RegisterName("applicationDidFinishLaunching:")
	sel_applicationWillTerminate                        = objc.RegisterName("applicationWillTerminate:")
	sel_applicationShouldTerminateAfterLastWindowClosed = objc.RegisterName("applicationShouldTerminateAfterLastWindowClosed:")
	sel_sharedApplication                               = objc.RegisterName("sharedApplication")
	sel_setActivationPolicy                             = objc.RegisterName("setActivationPolicy:")
	sel_activateIgnoringOtherApps                       = objc.RegisterName("activateIgnoringOtherApps:")
	sel_setDelegate                                     = objc.RegisterName("setDelegate:")
	sel_run                                             = objc.RegisterName("run")
	sel_stop                                            = objc.RegisterName("stop:")
	sel_terminate                                       = objc.RegisterName("terminate:")
)

const (
	NSApplicationActivationPolicyRegular    = 0
	NSApplicationActivationPolicyAccessory  = 1
	NSApplicationActivationPolicyProhibited = 2
)

type defaultAppDelegate struct {
	isa objc.Class `objc:"NSDefaultAppDelegate : NSObject <NSApplicationDelegate>"`
}

func (d *defaultAppDelegate) ApplicationDidFinishLaunching(_ objc.SEL, notification objc.ID) {
	utils.Logger().Println("11111111111111111111111")
}

func (d *defaultAppDelegate) ApplicationWillTerminate(_ objc.SEL, notification objc.ID) {
	utils.Logger().Println("222222222222222222222222")
}

func (*defaultAppDelegate) ApplicationShouldTerminateAfterLastWindowClosed(_ objc.SEL, _ objc.ID) bool {
	return true
}

func (*defaultAppDelegate) Selector(metName string) objc.SEL {
	switch metName {
	case "ApplicationDidFinishLaunching":
		return sel_applicationDidFinishLaunching
	case "ApplicationWillTerminate":
		return sel_applicationWillTerminate
	case "ApplicationShouldTerminateAfterLastWindowClosed":
		return sel_applicationShouldTerminateAfterLastWindowClosed
	default:
		return 0
	}
}

func DefaultAppDelegate() objc.ID {
	return objc.ID(class_DefaultAppDelegate).Send(sel_alloc).Send(sel_init)
}

type NSApplication struct {
	objc.ID
}

func NSApplication_sharedApplication() NSApplication {
	return NSApplication{objc.ID(class_NSApplication).Send(sel_sharedApplication)}
}

func NSApp() NSApplication {
	return NSApplication_sharedApplication()
}

func (a NSApplication) SetActivationPolicy(policy int) {
	a.Send(sel_setActivationPolicy, policy)
}

func (a NSApplication) ActivateIgnoringOtherApps(flag bool) {
	a.Send(sel_activateIgnoringOtherApps, flag)
}

func (a NSApplication) SetDelegate(delegate objc.ID) {
	a.Send(sel_setDelegate, delegate)
}

func (a NSApplication) Stop(sender objc.ID) {
	a.Send(sel_stop, sender)
}

func (a NSApplication) Terminate(sender objc.ID) {
	a.Send(sel_terminate, sender)
}

func (a NSApplication) Run() {
	a.Send(sel_run)
}

func init() {
	var err error
	class_DefaultAppDelegate, err = objc.RegisterClass(&defaultAppDelegate{})
	if err != nil {
		panic(err)
	}

}
