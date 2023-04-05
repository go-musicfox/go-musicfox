package mediaplayer

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func init() {
	importFramework()
	class_MPNowPlayingInfoCenter = objc.GetClass("MPNowPlayingInfoCenter")
}

var (
	class_MPNowPlayingInfoCenter objc.Class
)

var (
	sel_playbackState     = objc.RegisterName("playbackState")
	sel_setPlaybackState  = objc.RegisterName("setPlaybackState:")
	sel_nowPlayingInfo    = objc.RegisterName("nowPlayingInfo")
	sel_setNowPlayingInfo = objc.RegisterName("setNowPlayingInfo:")
)

type MPNowPlayingInfoCenter struct {
	core.NSObject
}

func MPNowPlayingInfoCenter_defaultCenter() MPNowPlayingInfoCenter {
	return MPNowPlayingInfoCenter{NSObject: core.NSObject{ID: objc.ID(class_MPNowPlayingInfoCenter).Send(macdriver.SEL_defaultCenter)}}
}

func (c MPNowPlayingInfoCenter) PlaybackState() MPNowPlayingPlaybackState {
	return MPNowPlayingPlaybackState(c.Send(sel_playbackState))
}

func (c MPNowPlayingInfoCenter) SetPlaybackState(state MPNowPlayingPlaybackState) {
	c.Send(sel_setPlaybackState, state)
}

func (c MPNowPlayingInfoCenter) NowPlayingInfo() core.NSDictionary {
	var dic core.NSDictionary
	dic.SetObjcID(c.Send(sel_nowPlayingInfo))
	return dic
}

func (c MPNowPlayingInfoCenter) SetNowPlayingInfo(info core.NSDictionary) {
	c.Send(sel_setNowPlayingInfo, info.ID)
}
