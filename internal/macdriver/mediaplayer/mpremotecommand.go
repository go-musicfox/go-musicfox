//go:build darwin

package mediaplayer

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_MPRemoteCommand = objc.GetClass("MPRemoteCommand")
}

var (
	//nolint:golint,unused
	class_MPRemoteCommand objc.Class
)

var (
	sel_addTargetAction    = objc.RegisterName("addTarget:action:")
	sel_removeTargetAction = objc.RegisterName("removeTarget:action:")
)

type MPRemoteCommand struct {
	core.NSObject
}

func (c MPRemoteCommand) AddTargetAction(target objc.ID, action objc.SEL) {
	c.Send(sel_addTargetAction, target, action)
}

// RemoveTargetAction 移除 addTarget:action: 注册的处理目标。addTarget 是
// 追加语义，重复注册会累积（每调一次 SetPlayingInfo 都翻倍触发），幂等
// 注册必须先 remove 再 add。
func (c MPRemoteCommand) RemoveTargetAction(target objc.ID, action objc.SEL) {
	c.Send(sel_removeTargetAction, target, action)
}
