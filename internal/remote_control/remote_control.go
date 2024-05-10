//go:build !darwin && !linux && !windows

package remote_control

import "time"

type RemoteControl struct {
}

func NewRemoteControl(Controller, PlayingInfo) *RemoteControl {
	return &RemoteControl{}
}

func (s *RemoteControl) SetPosition(time.Duration) {
}

func (s *RemoteControl) SetPlayingInfo(PlayingInfo) {
}

func (s *RemoteControl) Release() {
}
