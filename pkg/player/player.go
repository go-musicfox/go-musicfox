package player

import (
	"time"
)

type Player interface {
	Play(songType SongType, url string, duration time.Duration)
	CurMusic() UrlMusic
	Paused()
	Resume()
	Stop()
	PassedTime() time.Duration
	TimeChan() <-chan time.Duration
	State() State
	StateChan() <-chan State
	UpVolume()
	DownVolume()
	Close()
}
