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
	sel_playerItemWithURL   = objc.RegisterName("playerItemWithURL:")
	sel_playerItemWithAsset = objc.RegisterName("playerItemWithAsset:")
	sel_asset               = objc.RegisterName("asset")
	sel_duration            = objc.RegisterName("duration")
	sel_setAudioMix         = objc.RegisterName("setAudioMix:")
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

// AttachAudioTap configures an AVAudioMix for the first audio track and
// transfers the tap's creation reference to the audio mix.
func (i AVPlayerItem) AttachAudioTap(tap *AudioTap) bool {
	if i.ID == 0 || tap == nil {
		return false
	}
	tracks := i.Asset().AudioTracks()
	if len(tracks) == 0 {
		return false
	}
	parameters := AVMutableAudioMixInputParameters_audioMixInputParametersWithTrack(tracks[0])
	if parameters.ID == 0 {
		return false
	}
	parameters.SetAudioTapProcessor(tap)
	mix := AVMutableAudioMix_audioMix()
	if mix.ID == 0 {
		return false
	}
	mix.SetInputParameters(core.NSArray_arrayWithObject(parameters.NSObject))
	i.Send(sel_setAudioMix, mix.ID)
	tap.Close()
	return true
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
