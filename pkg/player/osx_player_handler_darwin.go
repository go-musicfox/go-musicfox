//go:build darwin

package player

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/avcore"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func init() {
	var err error
	class_AVPlayerHandler, err = objc.RegisterClass(&playerHandlerBinding{})
	if err != nil {
		panic(err)
	}
}

var (
	class_AVPlayerHandler objc.Class
	_player               *osxPlayer
)

var (
	sel_handleFinish = objc.RegisterName("handleFinish:")
)

type playerHandlerBinding struct {
	isa objc.Class `objc:"AVPlayerHandler : NSObject"`
}

func (playerHandlerBinding) HandleFinish(_ objc.SEL, event objc.ID) {
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

func (playerHandlerBinding) Selector(metName string) objc.SEL {
	switch metName {
	case "HandleFinish":
		return sel_handleFinish
	default:
		return 0
	}
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
