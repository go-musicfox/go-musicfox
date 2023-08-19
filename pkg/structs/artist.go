package structs

import (
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type Artist struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func NewArtist(json []byte) (Artist, error) {
	var artist Artist

	if len(json) == 0 {
		return artist, errors.New("json is empty")
	}

	arId, err := jsonparser.GetInt(json, "id")
	if err != nil {
		return artist, err
	}
	artist.Id = arId

	if arName, err := jsonparser.GetString(json, "name"); err == nil {
		artist.Name = arName
	}

	return artist, nil
}
