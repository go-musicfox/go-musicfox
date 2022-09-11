package player

type State uint8

const (
	Unknown State = iota
	Playing
	Paused
	Stopped
	Interrupted
)
