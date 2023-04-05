//go:build darwin

package mediaplayer

import (
	"github.com/ebitengine/purego/objc"
)

func init() {
	importFramework()
	class_MPChangePlaybackPositionCommand = objc.GetClass("MPChangePlaybackPositionCommand")
}

var (
	class_MPChangePlaybackPositionCommand objc.Class
)

type MPChangePlaybackPositionCommand struct {
	MPRemoteCommand
}
