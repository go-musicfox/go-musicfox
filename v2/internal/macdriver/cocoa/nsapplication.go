//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSApplication = objc.GetClass("NSApplication")
}

var (
	class_NSApplication objc.Class
)

var (
	sel_sharedApplication         = objc.RegisterName("sharedApplication")
	sel_setActivationPolicy       = objc.RegisterName("setActivationPolicy:")
	sel_activateIgnoringOtherApps = objc.RegisterName("activateIgnoringOtherApps:")
	sel_setDelegate               = objc.RegisterName("setDelegate:")
	sel_run                       = objc.RegisterName("run")
	sel_stop                      = objc.RegisterName("stop:")
	sel_terminate                 = objc.RegisterName("terminate:")
)

type NSApplicationActivationPolicy core.NSInteger

const (
	NSApplicationActivationPolicyRegular NSApplicationActivationPolicy = iota
	NSApplicationActivationPolicyAccessory
	NSApplicationActivationPolicyProhibited
)

type Delegate interface {
	Delegate() objc.ID
}

type NSApplication struct {
	core.NSObject
}

func NSApplication_sharedApplication() NSApplication {
	return NSApplication{
		core.NSObject{
			ID: objc.ID(class_NSApplication).Send(sel_sharedApplication),
		},
	}
}

func NSApp() NSApplication {
	return NSApplication_sharedApplication()
}

func (a NSApplication) SetActivationPolicy(policy NSApplicationActivationPolicy) {
	a.Send(sel_setActivationPolicy, policy)
}

func (a NSApplication) ActivateIgnoringOtherApps(flag bool) {
	a.Send(sel_activateIgnoringOtherApps, flag)
}

func (a NSApplication) SetDelegate(delegate Delegate) {
	a.Send(sel_setDelegate, delegate.Delegate())
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
