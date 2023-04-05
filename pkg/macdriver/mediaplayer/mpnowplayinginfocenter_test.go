//go:build darwin

package mediaplayer

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func TestMPNowPlayingInfoCenter(t *testing.T) {
	center := MPNowPlayingInfoCenter_defaultCenter()
	if center.ID == 0 {
		panic("get default center failed")
	}

	var (
		setID    int32 = 123
		setTitle       = "title"
	)

	dic := core.NSMutableDictionary_init()
	defer dic.Release()

	dic.SetValueForKey(core.String(MPMediaItemPropertyPersistentID), core.NSNumber_numberWithInt(setID).NSObject)
	dic.SetValueForKey(core.String(MPMediaItemPropertyTitle), core.String(setTitle).NSObject)

	center.SetNowPlayingInfo(dic.NSDictionary)
	center.SetPlaybackState(MPNowPlayingPlaybackStatePlaying)

	var (
		gotID    core.NSNumber
		gotTitle core.NSString
	)

	nowPlayingInfo := center.NowPlayingInfo()
	gotID.SetObjcID(nowPlayingInfo.ValueForKey(core.String(MPMediaItemPropertyPersistentID)))
	if gotID.IntValue() != setID {
		panic("got id error")
	}

	gotTitle.SetObjcID(nowPlayingInfo.ValueForKey(core.String(MPMediaItemPropertyTitle)))
	if gotTitle.String() != setTitle {
		panic("got title error")
	}

	state := center.PlaybackState()
	if state != MPNowPlayingPlaybackStatePlaying {
		panic("get playback state failed")
	}
}
