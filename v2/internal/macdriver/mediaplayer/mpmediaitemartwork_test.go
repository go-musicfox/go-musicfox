//go:build darwin

package mediaplayer

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver/core"
)

func TestMPMediaItemArtwork(t *testing.T) {
	url := core.NSURL_URLWithString(core.NSString_alloc().InitWithUTF8String("https://avatars.githubusercontent.com/u/27605589"))
	defer url.Release()

	image := cocoa.NSImage_alloc().InitWithContentsOfURL(url)
	defer image.Release()
	if image.ID == 0 {
		panic("init image failed")
	}

	art := MPMediaItemArtwork_alloc().InitWithImage(image)
	defer art.Release()
	if art.ID == 0 {
		panic("init artwork failed")
	}
}
