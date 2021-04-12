package ds

import "time"

type Song struct {
	Id int64
	Name string
	Duration time.Duration
	Artists []Artist
	Album
}
