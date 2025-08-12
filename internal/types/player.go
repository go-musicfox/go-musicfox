package types

// Mode 播放模式
type Mode uint8

const (
	PmUnknown Mode = iota
	PmListLoop
	PmOrdered
	PmSingleLoop
	PmListRandom
	PmInfRandom
	PmIntelligent
)

type State uint8

const (
	Unknown State = iota
	Playing
	Paused
	Stopped
	Interrupted
)
