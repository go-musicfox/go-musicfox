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

var typeNames = map[SongType]string{
	0: "MP3",
	1: "WAV",
	2: "OGG",
	3: "FLAC",
}

func SongTypeName(t SongType) string {
	if name, ok := typeNames[t]; ok {
		return name
	}
	return "未知"
}
