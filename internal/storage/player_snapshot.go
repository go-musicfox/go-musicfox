package storage

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type PlayerSnapshot struct {
	CurSongIndex     int            `json:"cur_song_index"`
	Playlist         []structs.Song `json:"playlist"`
	PlaylistUpdateAt time.Time      `json:"playlist_update_at"`
}

func (p PlayerSnapshot) GetDbName() string {
	return types.AppDBName
}

func (p PlayerSnapshot) GetTableName() string {
	return "default_bucket"
}

func (p PlayerSnapshot) GetKey() string {
	return "playlist_snapshot"
}
