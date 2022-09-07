package state

import "time"

type Player interface {
	Paused()
	Resume()
}

const (
	Unknown uint8 = iota
	Playing
	Paused
	Stopped
	Interrupted
)

type PlayingInfo struct {
	TotalDuration  time.Duration
	PassedDuration time.Duration
	Rate           float32
}
