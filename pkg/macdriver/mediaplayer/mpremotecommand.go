//go:build darwin

package mediaplayer

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func init() {
	importFramework()
	class_MPRemoteCommand = objc.GetClass("MPRemoteCommand")
}

var (
	class_MPRemoteCommand objc.Class
)

var (
	sel_addTargetAction = objc.RegisterName("addTarget:action:")
)

type MPRemoteCommand struct {
	core.NSObject
}

func (c MPRemoteCommand) AddTargetAction(target objc.ID, action objc.SEL) {
	c.Send(sel_addTargetAction, target, action)
}
