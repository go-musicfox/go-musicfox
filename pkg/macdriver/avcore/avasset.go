//go:build darwin

package avcore

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func init() {
	importFramework()
	class_AVAsset = objc.GetClass("AVAsset")
}

var (
	//nolint:golint,unused
	class_AVAsset objc.Class
)

var (
	sel_URL = objc.RegisterName("URL")
)

type AVAsset struct {
	core.NSObject
}

func (i AVAsset) URL() (url core.NSURL) {
	url.SetObjcID(i.Send(sel_URL))
	return
}
