package state_handler

import "time"

type Controller interface {
	Paused()
	Resume()
	Stop()
	Toggle()
	Next()
	Previous()
	Seek(duration time.Duration)
}
