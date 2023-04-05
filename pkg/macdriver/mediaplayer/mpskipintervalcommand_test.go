package mediaplayer

import (
	"fmt"
	"testing"

	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func TestMPSkipIntervalCommand(t *testing.T) {
	center := MPRemoteCommandCenter_sharedCommandCenter()
	cmd := center.SkipBackwardCommand()
	if cmd.ID == 0 {
		panic("get skip backward command failed")
	}

	cmd.SetPreferredIntervals(core.NSArray_arrayWithObject(core.NSNumber_numberWithDouble(15.0).NSObject))
	tp := cmd.PreferredIntervals()
	fmt.Println(tp)
}
