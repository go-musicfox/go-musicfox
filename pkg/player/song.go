package player

import "time"

// SongType 歌曲类型
type SongType uint8

const (
	Mp3 SongType = iota
	Wav
	Ogg
	Flac
)

type UrlMusic struct {
	Url      string
	Type     SongType
	Duration time.Duration
}
