//go:build darwin

package mediaplayer

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver/core"
)

func init() {
	importFramework()
	class_MPSkipIntervalCommand = objc.GetClass("MPSkipIntervalCommand")
}

var (
	//nolint:golint,unused
	class_MPSkipIntervalCommand objc.Class
)

var (
	sel_preferredIntervals    = objc.RegisterName("preferredIntervals")
	sel_setPreferredIntervals = objc.RegisterName("setPreferredIntervals:")
)

type MPSkipIntervalCommand struct {
	MPRemoteCommand
}

func (cmd MPSkipIntervalCommand) PreferredIntervals() objc.ID {
	return cmd.Send(sel_preferredIntervals)
}

func (cmd MPSkipIntervalCommand) SetPreferredIntervals(arr core.NSArray) {
	cmd.Send(sel_setPreferredIntervals, arr.ID)
}
