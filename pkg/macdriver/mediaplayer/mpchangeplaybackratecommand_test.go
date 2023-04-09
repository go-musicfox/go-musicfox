//go:build darwin

package mediaplayer

import (
	"fmt"
	"testing"
)

func TestMPChangePlaybackRateCommand(t *testing.T) {
	center := MPRemoteCommandCenter_sharedCommandCenter()
	cmd := center.ChangePlaybackRateCommand()
	if cmd.ID == 0 {
		panic("get change playback rate command failed")
	}

	id := cmd.SupportedPlaybackRates()
	fmt.Println(id)
}
