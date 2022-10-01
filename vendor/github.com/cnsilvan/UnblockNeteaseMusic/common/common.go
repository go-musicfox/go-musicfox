package common

import (
	"math/rand"
	"time"
)

type MapType = map[string]interface{}
type SliceType = []interface{}
type Song struct {
	Id                string
	Size              int64
	Br                int
	Url               string
	Md5               string
	Name              string
	Artist            string
	AlbumName         string
	MatchScore        float32
	Source            string
	PlatformUniqueKey MapType `json:"-"`
}
type SongSlice []*Song

func (a SongSlice) Len() int {
	return len(a)
}
func (a SongSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a SongSlice) Less(i, j int) bool {
	return a[j].MatchScore < a[i].MatchScore
}

type SearchSong struct {
	Keyword     string
	Name        string
	ArtistsName string
	Quality     MusicQuality
	OrderBy     SearchOrderBy
	Limit       int
	ArtistList  []string
}
type PlatformIdTag string

const (
	StartTag PlatformIdTag = "9000"
	KuWoTag  PlatformIdTag = "90000"
	MiGuTag  PlatformIdTag = "90001"
	KuGouTag PlatformIdTag = "90002"
	QQTag    PlatformIdTag = "90003"
)

type SearchOrderBy int32

const (
	MatchedScoreDesc SearchOrderBy = iota
	PlatformDefault
)

type MusicQuality int32

const (
	Standard MusicQuality = iota
	Higher
	ExHigh
	Lossless
)

type SearchMusic struct {
	Quality MusicQuality
	Id      string
}

func (m MusicQuality) String() string {
	switch m {
	case Standard:
		return "Standard(0)"
	case Higher:
		return "Higher(1)"
	case ExHigh:
		return "ExHigh(2)"
	case Lossless:
		return "Lossless(3)"
	default:
		return "UNKNOWN"
	}
}

var (
	ProxyIp     = "127.0.0.1"
	ProxyDomain = map[string]string{
		"music.163.com":            "59.111.181.35",
		"interface.music.163.com":  "59.111.181.35",
		"interface3.music.163.com": "59.111.181.35",
		"apm.music.163.com":        "59.111.181.35",
		"apm3.music.163.com":       "59.111.181.35",
	}
	HostDomain = map[string]string{
		"music.163.com":           "59.111.181.35",
		"interface.music.163.com": "59.111.181.35",
	}
	Source []string
	Rand   = rand.New(
		rand.NewSource(time.Now().UnixNano()))
)
