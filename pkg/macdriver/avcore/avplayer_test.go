//go:build darwin

package avcore

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func TestAVPlayer(t *testing.T) {
	player := AVPlayer_alloc().InitWithPlayerItem(AVPlayerItem_playerItemWithURL(core.NSURL_URLWithString(core.String("http://m801.music.126.net/20230405213504/57bd814bc9ce001962f988d48145b250/jdymusic/obj/wo3DlMOGwrbDjj7DisKw/9444694535/c8bb/65da/db54/9e257ff0610f84e4242cf0127dac6005.mp3"))))
	if player.ID == 0 {
		panic("init player failed")
	}

	player.SetActionAtItemEnd(2)
	player.SetVolume(100)
	item := player.CurrentItem()
	if item.ID == 0 {
		panic("get player current item failed")
	}
}
