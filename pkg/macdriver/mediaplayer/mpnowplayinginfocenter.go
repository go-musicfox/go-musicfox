package mediaplayer

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
)

var (
	class_MPNowPlayingInfoCenter = objc.GetClass("MPNowPlayingInfoCenter")
)

var (
	sel_setPlaybackState  = objc.RegisterName("setPlaybackState:")
	sel_setNowPlayingInfo = objc.RegisterName("setNowPlayingInfo:")
)

type MPNowPlayingPlaybackState uint

const (
	MPNowPlayingPlaybackStateUnknown MPNowPlayingPlaybackState = iota
	MPNowPlayingPlaybackStatePlaying
	MPNowPlayingPlaybackStatePaused
	MPNowPlayingPlaybackStateStopped
	MPNowPlayingPlaybackStateInterrupted
)

type MPNowPlayingInfoCenter struct {
	objc.ID
}

func MPNowPlayingInfoCenter_defaultCenter() MPNowPlayingInfoCenter {
	return MPNowPlayingInfoCenter{objc.ID(class_MPNowPlayingInfoCenter).Send(macdriver.SEL_defaultCenter)}
}

func (c MPNowPlayingInfoCenter) SetPlaybackState(state MPNowPlayingPlaybackState) {
	c.Send(sel_setPlaybackState, state)
}

func (c MPNowPlayingInfoCenter) SetNowPlayingInfo(info objc.ID) {
	c.Send(sel_setPlaybackState, info)
}

func (c MPNowPlayingInfoCenter) Release() {
	c.Send(macdriver.SEL_release)
}
