package player

import "github.com/go-musicfox/go-musicfox/pkg/structs"

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

type UrlMusic struct {
	Url  string
	Type SongType
	structs.Song
}
