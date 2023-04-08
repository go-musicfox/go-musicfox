//go:build darwin

package core

import (
	"github.com/ebitengine/purego/objc"
)

func init() {
	importFramework()
	class_NSURL = objc.GetClass("NSURL")
}

var (
	class_NSURL objc.Class
)

var (
	sel_URLWithString   = objc.RegisterName("URLWithString:")
	sel_fileURLWithPath = objc.RegisterName("fileURLWithPath:")
	sel_host            = objc.RegisterName("host")
	sel_absoluteString  = objc.RegisterName("absoluteString")
)

type NSURL struct {
	NSObject
}

func NSURL_URLWithString(url NSString) NSURL {
	return NSURL{
		NSObject{
			ID: objc.ID(class_NSURL).Send(sel_URLWithString, url.ID),
		},
	}
}

func NSURL_fileURLWithPath(path NSString) NSURL {
	return NSURL{
		NSObject{
			ID: objc.ID(class_NSURL).Send(sel_fileURLWithPath, path.ID),
		},
	}
}

func (url NSURL) Host() NSString {
	id := url.Send(sel_host)
	return NSString{NSObject{ID: id}}
}

func (url NSURL) AbsoluteString() NSString {
	id := url.Send(sel_absoluteString)
	return NSString{NSObject{ID: id}}
}
