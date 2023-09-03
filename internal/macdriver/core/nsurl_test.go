//go:build darwin

package core

import (
	"testing"
)

func TestNSURL(t *testing.T) {
	u := NSString_alloc().InitWithUTF8String("https://github.com/anhoder")
	defer u.Release()

	url := NSURL_URLWithString(u)
	defer url.Release()

	if url.Host().String() != "github.com" {
		panic("host is wrong")
	}

	if url.AbsoluteString().String() != "https://github.com/anhoder" {
		panic("absoluteString is wrong")
	}
}
