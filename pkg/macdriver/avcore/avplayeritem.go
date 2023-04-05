//go:build darwin

package avcore

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
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
)

type AVPlayerItem struct {
	core.NSObject
}

func AVPlayerItem_playerItemWithURL(url core.NSURL) AVPlayerItem {
	return AVPlayerItem{
		NSObject: core.NSObject{
			ID: objc.ID(class_AVPlayerItem).Send(sel_playerItemWithURL, url.ID),
		},
	}
}
