//go:build darwin

package mediaplayer

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func init() {
	importFramework()
	class_MPChangeShuffleModeCommand = objc.GetClass("MPChangeShuffleModeCommand")
}

var (
	//nolint:golint,unused
	class_MPChangeShuffleModeCommand objc.Class
)

var (
	sel_currentShuffleType = objc.RegisterName("currentShuffleType")
)

type MPShuffleType core.NSInteger

const (
	MPShuffleTypeOff MPShuffleType = iota
	MPShuffleTypeItems
	MPShuffleTypeCollections
)

type MPChangeShuffleModeCommand struct {
	MPRemoteCommand
}

func (cmd MPChangeShuffleModeCommand) CurrentShuffleType() MPShuffleType {
	return MPShuffleType(cmd.Send(sel_currentShuffleType))
}
