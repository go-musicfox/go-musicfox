//go:build darwin

package mediaplayer

import (
	"testing"
)

func TestMPChangePlaybackPositionCommand(t *testing.T) {
	center := MPRemoteCommandCenter_sharedCommandCenter()
	cmd := center.ChangePlaybackPositionCommand()
	if cmd.ID == 0 {
		panic("get change playback position command error")
	}
}
