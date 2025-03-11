package storage

import (
	"encoding/json"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type Scrobble struct {
	Artist     []string      `json:"artist"`
	Track      string        `json:"track"`
	Album      string        `json:"album"`
	Duration   time.Duration `json:"duration"`
	Timestamp  int64         `json:"timestamp,omitempty"`
	PlayedTime time.Duration `json:"playedtime"`
}

func NewScrobble(song structs.Song, playedTime time.Duration) *Scrobble {
	return &Scrobble{
		Artist:     ArtistNames(song.Artists),
		Track:      song.Name,
		Album:      song.Album.Name,
		Timestamp:  time.Now().Unix(),
		Duration:   song.Duration,
		PlayedTime: playedTime,
	}
}

type ScrobbleList struct {
	Scrobbles []Scrobble
}

// 添加一个 Scrobble
func (sl *ScrobbleList) Add(scrobble Scrobble) {
	sl.Scrobbles = append(sl.Scrobbles, scrobble)
}

func (sl *ScrobbleList) GetDbName() string {
	return types.AppDBName
}

func (sl *ScrobbleList) GetTableName() string {
	return "default_bucket"
}

func (sl *ScrobbleList) GetKey() string {
	return "lastfm_scrobble_list"
}

func (sl *ScrobbleList) Store() {
	t := NewTable()
	_ = t.SetByKVModel(sl, sl.Scrobbles)
}

func (sl *ScrobbleList) Clear() {
	t := NewTable()
	_ = t.DeleteByKVModel(sl)
}

func (sl *ScrobbleList) InitFromStorage() {
	t := NewTable()
	if jsonStr, err := t.GetByKVModel(sl); err == nil {
		_ = json.Unmarshal(jsonStr, &sl.Scrobbles)
	}
}

func ArtistNames(artists []structs.Artist) []string {
	names := make([]string, len(artists))
	for i, artist := range artists {
		names[i] = artist.Name
	}
	return names
}

func (s *Scrobble) FilterArtist() {
	s.Artist = s.Artist[:1]
}
