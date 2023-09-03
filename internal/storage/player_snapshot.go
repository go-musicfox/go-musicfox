package storage

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/constants"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

type PlayerSnapshot struct {
	CurSongIndex     int            `json:"cur_song_index"`
	Playlist         []structs.Song `json:"playlist"`
	PlaylistUpdateAt time.Time      `json:"playlist_update_at"`
}

func (p PlayerSnapshot) GetDbName() string {
	return constants.AppDBName
}

func (p PlayerSnapshot) GetTableName() string {
	return "default_bucket"
}

func (p PlayerSnapshot) GetKey() string {
	return "playlist_snapshot"
}
