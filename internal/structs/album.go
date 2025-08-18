package structs

import (
	"strings"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type Album struct {
	Id      int64    `json:"id"`
	Name    string   `json:"name"`
	PicUrl  string   `json:"pic_url"`
	Artists []Artist `json:"artists"`
}

func (a Album) ArtistName() string {
	var artistNames []string
	for _, artist := range a.Artists {
		artistNames = append(artistNames, artist.Name)
	}
	return strings.Join(artistNames, ",")
}

// NewAlbumFromJson 获取歌曲列表的专辑信息
func NewAlbumFromJson(json []byte, keys ...string) (Album, error) {
	var album Album
	if len(json) == 0 {
		return album, errors.New("json is empty")
	}

	targetData := json
	if len(keys) > 0 {
		extractedData, _, _, err := jsonparser.Get(json, keys...)
		if err != nil {
			return album, err
		}
		targetData = extractedData
	}

	alId, err := jsonparser.GetInt(targetData, "id")
	if err != nil {
		return album, err
	}
	album.Id = alId

	if alName, err := jsonparser.GetString(targetData, "name"); err == nil {
		album.Name = alName
	}

	if alPic, err := jsonparser.GetString(targetData, "picUrl"); err == nil {
		album.PicUrl = alPic
	}

	return album, nil
}

// NewAlbumFromAlbumJson 从Album列表获取专辑信息
func NewAlbumFromAlbumJson(json []byte) (Album, error) {
	var album Album
	if len(json) == 0 {
		return album, errors.New("json is empty")
	}

	album, err := NewAlbumFromJson(json)
	if err != nil {
		return album, err
	}

	_, _ = jsonparser.ArrayEach(json, func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		artist, err := NewArtist(value)

		if err == nil {
			album.Artists = append(album.Artists, artist)
		}

	}, "artists")

	return album, nil
}
