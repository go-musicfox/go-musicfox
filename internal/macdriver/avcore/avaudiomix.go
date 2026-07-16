//go:build darwin

package avcore

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_AVMutableAudioMix = objc.GetClass("AVMutableAudioMix")
	class_AVMutableAudioMixInputParameters = objc.GetClass("AVMutableAudioMixInputParameters")
}

var (
	class_AVMutableAudioMix                objc.Class
	class_AVMutableAudioMixInputParameters objc.Class

	sel_audioMix                         = objc.RegisterName("audioMix")
	sel_setInputParameters               = objc.RegisterName("setInputParameters:")
	sel_audioMixInputParametersWithTrack = objc.RegisterName("audioMixInputParametersWithTrack:")
	sel_setAudioTapProcessor             = objc.RegisterName("setAudioTapProcessor:")
)

type AVMutableAudioMix struct {
	core.NSObject
}

func AVMutableAudioMix_audioMix() AVMutableAudioMix {
	return AVMutableAudioMix{
		NSObject: core.NSObject{ID: objc.ID(class_AVMutableAudioMix).Send(sel_audioMix)},
	}
}

func (m AVMutableAudioMix) SetInputParameters(parameters core.NSArray) {
	m.Send(sel_setInputParameters, parameters.ID)
}

type AVMutableAudioMixInputParameters struct {
	core.NSObject
}

func AVMutableAudioMixInputParameters_audioMixInputParametersWithTrack(track AVAssetTrack) AVMutableAudioMixInputParameters {
	return AVMutableAudioMixInputParameters{
		NSObject: core.NSObject{
			ID: objc.ID(class_AVMutableAudioMixInputParameters).Send(sel_audioMixInputParametersWithTrack, track.ID),
		},
	}
}

func (p AVMutableAudioMixInputParameters) SetAudioTapProcessor(tap *AudioTap) {
	p.Send(sel_setAudioTapProcessor, tap.ref)
}
