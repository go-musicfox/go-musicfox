package state_handler

import (
	"go-musicfox/pkg/player"
	"time"
)

type PlayingInfo struct {
	TotalDuration  time.Duration
	PassedDuration time.Duration
	State          player.State
}
