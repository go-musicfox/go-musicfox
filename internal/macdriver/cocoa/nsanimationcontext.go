//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_NSAnimationContext = objc.GetClass("NSAnimationContext")
	class_CAMediaTimingFunction = objc.GetClass("CAMediaTimingFunction")
}

var (
	class_NSAnimationContext    objc.Class
	class_CAMediaTimingFunction objc.Class

	sel_runAnimationGroup      = objc.RegisterName("runAnimationGroup:completionHandler:")
	sel_animator               = objc.RegisterName("animator")
	sel_setDuration            = objc.RegisterName("setDuration:")
	sel_setTimingFunction      = objc.RegisterName("setTimingFunction:")
	sel_setAllowsImplicitAnim  = objc.RegisterName("setAllowsImplicitAnimation:")
	sel_functionWithName       = objc.RegisterName("functionWithName:")
	sel_setFrameDisplayAnim   = objc.RegisterName("setFrame:display:")
)

// CAMediaTimingFunction names (Core Animation standard curves).
const (
	CAMediaTimingFunctionEaseOut    = "easeOut"
	CAMediaTimingFunctionEaseInEaseOut = "easeInEaseOut"
	CAMediaTimingFunctionLinear     = "linear"
)

// NSAnimationContext wraps an NSAnimationContext passed to an animation group.
type NSAnimationContext struct {
	core.NSObject
}

// SetDuration sets the animation duration in seconds for this context.
func (c NSAnimationContext) SetDuration(seconds CGFloat) {
	c.Send(sel_setDuration, seconds)
}

// SetTimingFunction sets the CAMediaTimingFunction easing curve.
func (c NSAnimationContext) SetTimingFunction(fn CAMediaTimingFunction) {
	c.Send(sel_setTimingFunction, fn.ID)
}

// SetAllowsImplicitAnimation enables implicit animation of layer-backed
// properties changed inside the group (e.g. sublayer geometry).
func (c NSAnimationContext) SetAllowsImplicitAnimation(allows bool) {
	c.Send(sel_setAllowsImplicitAnim, allows)
}

// CAMediaTimingFunction is a Core Animation easing curve.
type CAMediaTimingFunction struct {
	core.NSObject
}

// CAMediaTimingFunction_FunctionWithName returns a standard named timing curve.
func CAMediaTimingFunction_FunctionWithName(name string) CAMediaTimingFunction {
	nsName := core.String(name)
	return CAMediaTimingFunction{
		core.NSObject{
			ID: objc.ID(class_CAMediaTimingFunction).Send(sel_functionWithName, nsName.ID),
		},
	}
}

// NSAnimationContext_RunAnimationGroup runs animations grouped under a single
// NSAnimationContext driven by Core Animation. The animations closure receives
// the group's context so the caller can set duration/timing and mutate animator
// proxies; completion (may be nil) runs on the main thread when the animation
// settles.
//
// Both closures are invoked by AppKit on the main thread. The animations block
// executes synchronously before this call returns; the completion block fires
// later, so its backing objc.Block is copied and released by AppKit's own
// retain of the completion handler.
func NSAnimationContext_RunAnimationGroup(animations func(ctx NSAnimationContext), completion func()) {
	animBlock := objc.NewBlock(func(_ objc.Block, ctxID objc.ID) {
		animations(NSAnimationContext{core.NSObject{ID: ctxID}})
	})
	defer animBlock.Release()

	var completionArg any = objc.ID(0)
	var compBlock objc.Block
	if completion != nil {
		compBlock = objc.NewBlock(func(_ objc.Block) {
			completion()
		})
		completionArg = compBlock
	}

	objc.ID(class_NSAnimationContext).Send(sel_runAnimationGroup, animBlock, completionArg)

	if compBlock != 0 {
		compBlock.Release()
	}
}

// Animator returns the view's animation proxy. Setting frame properties on the
// proxy inside an NSAnimationContext group animates them via Core Animation.
func (v NSView) Animator() NSView {
	return NSView{
		core.NSObject{ID: v.Send(sel_animator)},
	}
}

// Animator returns the window's animation proxy.
func (w NSWindow) Animator() NSWindow {
	return NSWindow{
		core.NSObject{ID: w.Send(sel_animator)},
	}
}

// SetFrameDisplayAnimated sets the window frame through the animator proxy.
// Call on w.Animator() inside an animation group.
func (w NSWindow) SetFrameDisplayAnimated(frameRect NSRect, display bool) {
	w.Send(sel_setFrameDisplayAnim,
		frameRect.Origin.X, frameRect.Origin.Y,
		frameRect.Size.Width, frameRect.Size.Height,
		display,
	)
}
