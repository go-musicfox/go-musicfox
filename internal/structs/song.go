package structs

import (
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type Song struct {
	Id       int64         `json:"id"`
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Artists  []Artist      `json:"artists"`
	Album    `json:"album"`
}

func (s Song) ArtistName() string {
	var artistNames []string
	for _, artist := range s.Artists {
		artistNames = append(artistNames, artist.Name)
	}
	return strings.Join(artistNames, ",")
}

// NewSongFromShortNameSongsJson 从歌单获取数据
func NewSongFromShortNameSongsJson(json []byte) (Song, error) {
	var song Song
	if len(json) == 0 {
		return song, errors.New("json is empty")
	}

	id, err := jsonparser.GetInt(json, "id")
	if err != nil {
		return song, err
	}
	song.Id = id

	if name, err := jsonparser.GetString(json, "name"); err == nil {
		song.Name = name
	}
	if duration, err := jsonparser.GetInt(json, "dt"); err == nil {
		song.Duration = time.Millisecond * time.Duration(duration)
	}

	album, err := NewAlbumFromJson(json)
	if err == nil {
		song.Album = album
	}

	_, _ = jsonparser.ArrayEach(json, func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		artist, err := NewArtist(value)
		if err == nil {
			song.Artists = append(song.Artists, artist)
		}
	}, "ar")

	return song, nil
}

// NewSongFromFmJson 从私人FM获取数据
func NewSongFromFmJson(json []byte) (Song, error) {
	var song Song
	if len(json) == 0 {
		return song, errors.New("json is empty")
	}

	id, err := jsonparser.GetInt(json, "id")
	if err != nil {
		return song, err
	}
	song.Id = id

	if name, err := jsonparser.GetString(json, "name"); err == nil {
		song.Name = name
	}
	if duration, err := jsonparser.GetInt(json, "duration"); err == nil {
		song.Duration = time.Millisecond * time.Duration(duration)
	}
	if alId, err := jsonparser.GetInt(json, "album", "id"); err == nil {
		song.Album.Id = alId
	}
	if alName, err := jsonparser.GetString(json, "album", "name"); err == nil {
		song.Album.Name = alName
	}
	if alPic, err := jsonparser.GetString(json, "album", "picUrl"); err == nil {
		song.Album.PicUrl = alPic
	}

	_, _ = jsonparser.ArrayEach(json, func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		artist, err := NewArtist(value)

		if err == nil {
			song.Artists = append(song.Artists, artist)
		}
	}, "artists")

	return song, nil
}

// NewSongFromIntelligenceJson 心动模式获取数据
func NewSongFromIntelligenceJson(json []byte) (Song, error) {
	var song Song
	if len(json) == 0 {
		return song, errors.New("json is empty")
	}

	id, err := jsonparser.GetInt(json, "songInfo", "id")
	if err != nil {
		return song, err
	}
	song.Id = id

	if name, err := jsonparser.GetString(json, "songInfo", "name"); err == nil {
		song.Name = name
	}
	if duration, err := jsonparser.GetInt(json, "songInfo", "dt"); err == nil {
		song.Duration = time.Millisecond * time.Duration(duration)
	}
	if alId, err := jsonparser.GetInt(json, "songInfo", "al", "id"); err == nil {
		song.Album.Id = alId
	}
	if alName, err := jsonparser.GetString(json, "songInfo", "al", "name"); err == nil {
		song.Album.Name = alName
	}
	if alPic, err := jsonparser.GetString(json, "songInfo", "al", "picUrl"); err == nil {
		song.Album.PicUrl = alPic
	}

	_, _ = jsonparser.ArrayEach(json, func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		artist, err := NewArtist(value)

		if err == nil {
			song.Artists = append(song.Artists, artist)
		}
	}, "songInfo", "ar")

	return song, nil
}

// NewSongFromAlbumSongsJson 从专辑获取数据
func NewSongFromAlbumSongsJson(json []byte) (Song, error) {
	return NewSongFromShortNameSongsJson(json)
}

// NewSongFromArtistSongsJson 从歌手获取数据
func NewSongFromArtistSongsJson(json []byte) (Song, error) {
	return NewSongFromShortNameSongsJson(json)
}

// NewSongFromDjRadioProgramJson 从DjRadio节目中获取数据
func NewSongFromDjRadioProgramJson(json []byte) (Song, error) {
	var song Song
	if len(json) == 0 {
		return song, errors.New("json is empty")
	}

	id, err := jsonparser.GetInt(json, "mainSong", "id")
	if err != nil {
		return song, err
	}
	song.Id = id

	if name, err := jsonparser.GetString(json, "mainSong", "name"); err == nil {
		song.Name = name
	}
	if duration, err := jsonparser.GetInt(json, "mainSong", "duration"); err == nil {
		song.Duration = time.Millisecond * time.Duration(duration)
	}
	if alId, err := jsonparser.GetInt(json, "mainSong", "album", "id"); err == nil {
		song.Album.Id = alId
	}
	if alName, err := jsonparser.GetString(json, "mainSong", "album", "name"); err == nil {
		song.Album.Name = alName
	}
	if alPic, err := jsonparser.GetString(json, "mainSong", "album", "picUrl"); err == nil {
		song.Album.PicUrl = alPic
	}

	var artist Artist
	if arName, err := jsonparser.GetString(json, "dj", "nickname"); err == nil {
		artist.Name = arName
	}
	song.Artists = append(song.Artists, artist)

	return song, nil
}

// NewSongFromCloudJson 从DjRadio节目中获取数据
func NewSongFromCloudJson(json []byte) (Song, error) {
	var song Song
	if len(json) == 0 {
		return song, errors.New("json is empty")
	}

	id, err := jsonparser.GetInt(json, "songId")
	if err != nil {
		return song, err
	}
	song.Id = id

	if name, err := jsonparser.GetString(json, "songName"); err == nil {
		song.Name = name
	}
	if duration, err := jsonparser.GetInt(json, "simpleSong", "dt"); err == nil {
		song.Duration = time.Millisecond * time.Duration(duration)
	}
	if alId, err := jsonparser.GetInt(json, "simpleSong", "al", "id"); err == nil {
		song.Album.Id = alId
	}
	if alName, err := jsonparser.GetString(json, "simpleSong", "al", "name"); err == nil {
		song.Album.Name = alName
	}
	if alPic, err := jsonparser.GetString(json, "simpleSong", "al", "picUrl"); err == nil {
		song.Album.PicUrl = alPic
	}

	_, _ = jsonparser.ArrayEach(json, func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		artist, err := NewArtist(value)

		if err == nil {
			song.Artists = append(song.Artists, artist)
		}
	}, "simpleSong", "ar")

	return song, nil
}

// NewSongFromDjRankProgramJson 从DjRadio节目中获取数据
func NewSongFromDjRankProgramJson(json []byte) (Song, error) {
	var song Song
	if len(json) == 0 {
		return song, errors.New("json is empty")
	}

	id, err := jsonparser.GetInt(json, "program", "mainSong", "id")
	if err != nil {
		return song, err
	}
	song.Id = id

	if name, err := jsonparser.GetString(json, "program", "mainSong", "name"); err == nil {
		song.Name = name
	}
	if duration, err := jsonparser.GetInt(json, "program", "mainSong", "duration"); err == nil {
		song.Duration = time.Millisecond * time.Duration(duration)
	}
	if alId, err := jsonparser.GetInt(json, "program", "mainSong", "album", "id"); err == nil {
		song.Album.Id = alId
	}
	if alName, err := jsonparser.GetString(json, "program", "mainSong", "album", "name"); err == nil {
		song.Album.Name = alName
	}
	if alPic, err := jsonparser.GetString(json, "program", "mainSong", "album", "picUrl"); err == nil {
		song.Album.PicUrl = alPic
	}

	_, _ = jsonparser.ArrayEach(json, func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		artist, err := NewArtist(value)

		if err == nil {
			song.Artists = append(song.Artists, artist)
		}
	}, "program", "mainSong", "artists")

	return song, nil
}
