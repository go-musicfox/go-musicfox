//go:build darwin

package avcore

import (
	"unsafe"

	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_AVPlayerItem = objc.GetClass("AVPlayerItem")
}

var (
	class_AVPlayerItem objc.Class
)

var (
	sel_playerItemWithURL = objc.RegisterName("playerItemWithURL:")
	sel_asset             = objc.RegisterName("asset")
	sel_duration          = objc.RegisterName("duration")
)

type AVPlayerItem struct {
	core.NSObject
}

func AVPlayerItem_playerItemWithURL(url core.NSURL) AVPlayerItem {
	return AVPlayerItem{
		core.NSObject{
			ID: objc.ID(class_AVPlayerItem).Send(sel_playerItemWithURL, url.ID),
		},
	}
}

func (i AVPlayerItem) Asset() (asset AVAsset) {
	asset.SetObjcID(i.Send(sel_asset))
	return
}

func (i AVPlayerItem) Duration() (time CMTime) {
	sig := core.NSMethodSignature_instanceMethodSignatureForSelector(objc.ID(class_AVPlayerItem), sel_duration)
	inv := core.NSInvocation_invocationWithMethodSignature(sig)
	inv.SetSelector(sel_duration)
	inv.InvokeWithTarget(i.ID)
	inv.GetReturnValue(unsafe.Pointer(&time))
	return
}
