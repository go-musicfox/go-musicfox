//go:build darwin

package mediaplayer

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver/core"
)

func init() {
	importFramework()
	class_MPFeedbackCommand = objc.GetClass("MPFeedbackCommand")
}

var (
	//nolint:golint,unused
	class_MPFeedbackCommand objc.Class
)

var (
	sel_isActive               = objc.RegisterName("isActive")
	sel_setActive              = objc.RegisterName("setActive:")
	sel_localizedTitle         = objc.RegisterName("localizedTitle")
	sel_setLocalizedTitle      = objc.RegisterName("setLocalizedTitle:")
	sel_localizedShortTitle    = objc.RegisterName("localizedShortTitle")
	sel_setLocalizedShortTitle = objc.RegisterName("setLocalizedShortTitle:")
)

type MPFeedbackCommand struct {
	MPRemoteCommand
}

func (cmd MPFeedbackCommand) IsActive() bool {
	return objc.Send[bool](cmd.ID, sel_isActive)
}
func (cmd MPFeedbackCommand) SetActive(active bool) {
	cmd.Send(sel_setActive, active)
}

func (cmd MPFeedbackCommand) SetLocalizedTitle(title core.NSString) {
	cmd.Send(sel_setLocalizedTitle, title.ID)
}

func (cmd MPFeedbackCommand) LocalizedTitle() (str core.NSString) {
	str.SetObjcID(cmd.Send(sel_localizedTitle))
	return
}

func (cmd MPFeedbackCommand) SetLocalizedShortTitle(title core.NSString) {
	cmd.Send(sel_setLocalizedShortTitle, title.ID)
}

func (cmd MPFeedbackCommand) LocalizedShortTitle() (str core.NSString) {
	str.SetObjcID(cmd.Send(sel_localizedShortTitle))
	return
}
