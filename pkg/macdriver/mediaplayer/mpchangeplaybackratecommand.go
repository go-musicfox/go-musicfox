//go:build darwin

package mediaplayer

import (
	"github.com/ebitengine/purego/objc"
)

func init() {
	importFramework()
	class_MPChangePlaybackRateCommand = objc.GetClass("MPChangePlaybackRateCommand")
}

var (
	//nolint:golint,unused
	class_MPChangePlaybackRateCommand objc.Class
)

var (
	sel_supportedPlaybackRates = objc.RegisterName("supportedPlaybackRates")
)

type MPChangePlaybackRateCommand struct {
	MPRemoteCommand
}

func (cmd MPChangePlaybackRateCommand) SupportedPlaybackRates() objc.ID {
	return cmd.Send(sel_supportedPlaybackRates)
}
