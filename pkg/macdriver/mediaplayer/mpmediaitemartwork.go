package mediaplayer

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func init() {
	importFramework()
	class_MPMediaItemArtwork = objc.GetClass("MPMediaItemArtwork")
}

var (
	class_MPMediaItemArtwork objc.Class
)

var (
	sel_initWithImage = objc.RegisterName("initWithImage:")
)

type MPMediaItemArtwork struct {
	core.NSObject
}

func MPMediaItemArtwork_alloc() MPMediaItemArtwork {
	return MPMediaItemArtwork{
		NSObject: core.NSObject{
			ID: objc.ID(class_MPMediaItemArtwork).Send(macdriver.SEL_alloc),
		},
	}
}

func (art MPMediaItemArtwork) InitWithImage(image cocoa.NSImage) MPMediaItemArtwork {
	art.Send(sel_initWithImage, image.ID)
	return art
}
