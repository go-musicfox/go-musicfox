package state_handler

import "time"

type PlayingInfo struct {
	TotalDuration  time.Duration
	PassedDuration time.Duration
	Rate           float32
}
