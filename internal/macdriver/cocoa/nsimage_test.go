//go:build darwin

package cocoa

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func TestNSImage(t *testing.T) {
	url := core.NSURL_URLWithString(core.NSString_alloc().InitWithUTF8String("https://avatars.githubusercontent.com/u/27605589"))
	defer url.Release()

	image := NSImage_alloc().InitWithContentsOfURL(url)
	defer image.Release()

	if image.ID == 0 {
		panic("init image failed")
	}
}
