//go:build darwin

package mediaplayer

import (
	"fmt"
	"testing"
)

func TestMPChangeRepeatModeCommand(t *testing.T) {
	center := MPRemoteCommandCenter_sharedCommandCenter()
	cmd := center.ChangeRepeatModeCommand()
	if cmd.ID == 0 {
		panic("get change repeat mode command failed")
	}

	tp := cmd.CurrentRepeatType()
	fmt.Println(tp)
}
