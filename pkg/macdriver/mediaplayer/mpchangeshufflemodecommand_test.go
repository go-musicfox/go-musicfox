//go:build darwin

package mediaplayer

import (
	"fmt"
	"testing"
)

func TestMPChangeShuffleModeCommand(t *testing.T) {
	center := MPRemoteCommandCenter_sharedCommandCenter()
	cmd := center.ChangeShuffleModeCommand()
	if cmd.ID == 0 {
		panic("get change shuffle mode command failed")
	}

	tp := cmd.CurrentShuffleType()
	fmt.Println(tp)
}
