package structs

import (
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type Playlist struct {
	Id      int64
	Name    string
	Creator User
	Privacy bool
}

// NewPlaylistFromJson 获取歌单信息
func NewPlaylistFromJson(json []byte) (Playlist, error) {
	var playlist Playlist
	if len(json) == 0 {
		return playlist, errors.New("json is empty")
	}

	id, err := jsonparser.GetInt(json, "id")
	if err != nil {
		return playlist, err
	}
	playlist.Id = id

	if name, err := jsonparser.GetString(json, "name"); err == nil {
		playlist.Name = name
	}

	// privacy as int
	if privacy, err := jsonparser.GetInt(json, "privacy"); err == nil {
		playlist.Privacy = (privacy != 0)
	}

	if dj, err := NewUserFromJson(json, "creator"); err == nil {
		playlist.Creator = dj
	}

	return playlist, nil
}
