// +build darwin

package player

import (
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/mediaplayer"
)

func init() {
	go func() {
		app := cocoa.NSApp()
		app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
		app.ActivateIgnoringOtherApps(true)
		app.Run()
	}()

	playingCenter := mediaplayer.MPNowPlayingInfoCenter_defaultCenter()
	commandCenter := mediaplayer.MPRemoteCommandCenter_sharedCommandCenter()

	nowPlayingCenter = &playingCenter
	remoteCommandCenter = &commandCenter
}

type remoteCommandHandler struct {

}

func NewRemoteCommandHandler() {

}
