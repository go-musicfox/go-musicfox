//go:build darwin

package player

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/avcore"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

var (
	class_AVPlayerHandler objc.Class
	sel_handleFinish      = objc.RegisterName("handleFinish:")
	sel_handleFailed      = objc.RegisterName("handleFailed:")
	_player               *osxPlayer
)

func init() {
	var err error
	class_AVPlayerHandler, err = objc.RegisterClass(
		"AVPlayerHandler",
		objc.GetClass("NSObject"),
		[]*objc.Protocol{},
		[]objc.FieldDef{},
		[]objc.MethodDef{
			{
				Cmd: sel_handleFinish,
				Fn:  handleFinish,
			},
			{
				Cmd: sel_handleFailed,
				Fn:  handleFailed,
			},
		},
	)
	if err != nil {
		panic(err)
	}
}

func handleFinish(id objc.ID, cmd objc.SEL, event objc.ID) {
	// 这里会出现两次通知
	core.Autorelease(func() {
		var item avcore.AVPlayerItem
		item.SetObjcID(event.Send(macdriver.SEL_object))

		url := item.Asset().URL()
		curUrl := _player.player.CurrentItem().Asset().URL()
		if url.AbsoluteString().String() == curUrl.AbsoluteString().String() {
			_player.Stop()
		}
	})
}

func handleFailed(id objc.ID, cmd objc.SEL, event objc.ID) {
	// 播放失败处理
	core.Autorelease(func() {
		var item avcore.AVPlayerItem
		item.SetObjcID(event.Send(macdriver.SEL_object))

		url := item.Asset().URL()
		curUrl := _player.player.CurrentItem().Asset().URL()
		if url.AbsoluteString().String() == curUrl.AbsoluteString().String() {
			_player.Stop()
		}
	})
}

type playerHandler struct {
	core.NSObject
}

func playerHandler_new(p *osxPlayer) playerHandler {
	_player = p
	return playerHandler{
		core.NSObject{
			ID: objc.ID(class_AVPlayerHandler).Send(macdriver.SEL_new),
		},
	}
}
