//go:build darwin

package mediaplayer

import (
	"testing"
)

func TestMPRemoteCommand(t *testing.T) {
	center := MPRemoteCommandCenter_sharedCommandCenter()
	cmd := center.PlayCommand()
	if cmd.ID == 0 {
		panic("get play command error")
	}
}
