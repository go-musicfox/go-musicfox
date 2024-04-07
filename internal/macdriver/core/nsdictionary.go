//go:build darwin

package core

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
)

func init() {
	importFramework()
	class_NSDictionary = objc.GetClass("NSDictionary")
}

var (
	class_NSDictionary objc.Class
)

var (
	sel_valueForKey = objc.RegisterName("valueForKey:")
)

type NSDictionary struct {
	NSObject
}

func NSDictionary_init() NSDictionary {
	return NSDictionary{NSObject{ID: objc.ID(class_NSDictionary).Send(macdriver.SEL_init)}}
}

func (d NSDictionary) ValueForKey(key NSString) objc.ID {
	return d.Send(sel_valueForKey, key.ID)
}
