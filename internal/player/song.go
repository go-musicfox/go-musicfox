package player

import "github.com/go-musicfox/go-musicfox/internal/structs"

// SongType 歌曲类型
type SongType uint8

const (
	Mp3 SongType = iota
	Wav
	Ogg
	Flac
)

var SongTypeMapping = map[string]SongType{
	"mp3":  Mp3,
	"wav":  Wav,
	"ogg":  Ogg,
	"flac": Flac,
}

type URLMusic struct {
	URL string
	structs.Song
	Type SongType
}
