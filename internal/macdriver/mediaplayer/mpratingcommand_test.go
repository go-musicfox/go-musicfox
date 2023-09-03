//go:build darwin

package mediaplayer

import (
	"fmt"
	"testing"
)

func TestMPRatingCommand(t *testing.T) {
	center := MPRemoteCommandCenter_sharedCommandCenter()
	cmd := center.RatingCommand()
	if cmd.ID == 0 {
		panic("get rating command failed")
	}

	v1 := cmd.MinimumRating()
	fmt.Println(v1)

	v2 := cmd.MaximumRating()
	fmt.Println(v2)
}
