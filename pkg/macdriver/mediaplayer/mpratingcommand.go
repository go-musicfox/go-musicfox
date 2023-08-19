//go:build darwin

package mediaplayer

import (
	"github.com/ebitengine/purego/objc"
)

func init() {
	importFramework()
	class_MPRatingCommand = objc.GetClass("MPRatingCommand")
}

var (
	//nolint:golint,unused
	class_MPRatingCommand objc.Class
)

var (
	sel_maximumRating = objc.RegisterName("maximumRating")
	sel_minimumRating = objc.RegisterName("minimumRating")
)

type MPRatingCommand struct {
	MPRemoteCommand
}

func (cmd MPRatingCommand) MaximumRating() float32 {
	return objc.Send[float32](cmd.ID, sel_maximumRating)
}

func (cmd MPRatingCommand) MinimumRating() float32 {
	return objc.Send[float32](cmd.ID, sel_minimumRating)
}
