package state_handler

import "time"

type Controller interface {
	Paused()
	Resume()
	Stop()
	Toggle()
	NextSong()
	PreviousSong()
	Seek(duration time.Duration)
}
