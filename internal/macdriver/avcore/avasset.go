//go:build darwin

package avcore

import (
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	importFramework()
	class_AVAsset = objc.GetClass("AVAsset")
}

var (
	class_AVAsset      objc.Class
	audioMediaTypeOnce sync.Once
	avMediaTypeAudio   objc.ID
)

var (
	sel_URL                 = objc.RegisterName("URL")
	sel_tracksWithMediaType = objc.RegisterName("tracksWithMediaType:")
)

type AVAsset struct {
	core.NSObject
}

func (i AVAsset) URL() (url core.NSURL) {
	url.SetObjcID(i.Send(sel_URL))
	return
}

func (i AVAsset) AudioTracks() []AVAssetTrack {
	if i.ID == 0 {
		return nil
	}
	tracksID := i.Send(sel_tracksWithMediaType, avMediaTypeAudioID())
	tracks := core.NSArray{NSObject: core.NSObject{ID: tracksID}}
	count := tracks.Count()
	if count == 0 {
		return nil
	}
	audioTracks := make([]AVAssetTrack, 0, count)
	for index := uint(0); index < count; index++ {
		audioTracks = append(audioTracks, AVAssetTrack{NSObject: tracks.ObjectAtIndex(index)})
	}
	return audioTracks
}

func avMediaTypeAudioID() objc.ID {
	audioMediaTypeOnce.Do(func() {
		symbol, err := purego.Dlsym(avFoundationLib, "AVMediaTypeAudio")
		if err != nil {
			panic(err)
		}
		avMediaTypeAudio = *(*objc.ID)(unsafe.Pointer(symbol))
	})
	return avMediaTypeAudio
}
