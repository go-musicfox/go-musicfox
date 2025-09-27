//go:build darwin

package mediaplayer

import (
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver/core"
)

func init() {
	importFramework()
	class_MPChangeRepeatModeCommand = objc.GetClass("MPChangeRepeatModeCommand")
}

var (
	//nolint:golint,unused
	class_MPChangeRepeatModeCommand objc.Class
)

var (
	sel_currentRepeatType = objc.RegisterName("currentRepeatType")
)

type MPRepeatType core.NSInteger

const (
	MPRepeatTypeOff MPRepeatType = iota
	MPRepeatTypeOne
	MPRepeatTypeAll
)

type MPChangeRepeatModeCommand struct {
	MPRemoteCommand
}

func (cmd MPChangeRepeatModeCommand) CurrentRepeatType() MPRepeatType {
	return MPRepeatType(cmd.Send(sel_currentRepeatType))
}
