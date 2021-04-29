package ds

import (
	"errors"
	"github.com/buger/jsonparser"
)

type Album struct {
	Id     int64
	Name   string
	PicUrl string
}

func NewAlbumFromJson(json []byte) (Album, error) {
	var album Album
	if len(json) == 0 {
		return album, errors.New("json is empty")
	}

	alId, err := jsonparser.GetInt(json, "al", "id")
	if err != nil {
		return album, err
	}
	album.Id = alId

	if alName, err := jsonparser.GetString(json, "al", "name"); err == nil {
		album.Name = alName
	}

	if alPic, err := jsonparser.GetString(json, "al", "picUrl"); err == nil {
		album.PicUrl = alPic
	}

	return album, nil
}