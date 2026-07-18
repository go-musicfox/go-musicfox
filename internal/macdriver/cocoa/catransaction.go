//go:build darwin

package cocoa

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

func init() {
	class_CATransaction = objc.GetClass("CATransaction")
}

var class_CATransaction objc.Class

var (
	sel_begin             = objc.RegisterName("begin")
	sel_commit            = objc.RegisterName("commit")
	sel_setDisableActions = objc.RegisterName("setDisableActions:")
)

type CATransaction struct {
	core.NSObject
}

// CATransaction_Begin begins a new transaction for the current thread.
func CATransaction_Begin() {
	objc.ID(class_CATransaction).Send(sel_begin)
}

// CATransaction_Commit commits all changes made during the current transaction.
func CATransaction_Commit() {
	objc.ID(class_CATransaction).Send(sel_commit)
}

// CATransaction_SetDisableActions disables implicit animations during the transaction.
func CATransaction_SetDisableActions(disable bool) {
	objc.ID(class_CATransaction).Send(sel_setDisableActions, disable)
}
