package structs

import (
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type DjRadio struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	PicUrl  string `json:"pic_url"`
	Dj      User   `json:"dj"`
	Privacy bool   `json:"privacy"`
}

// NewDjRadioFromJson 从Json中初始化 DjRadio
func NewDjRadioFromJson(json []byte, keys ...string) (DjRadio, error) {
	var radio DjRadio
	if len(json) == 0 {
		return radio, errors.New("json is empty")
	}

	targetData := json
	if len(keys) > 0 {
		extractedData, _, _, err := jsonparser.Get(json, keys...)
		if err != nil {
			return radio, err
		}
		targetData = extractedData
	}

	radioId, err := jsonparser.GetInt(targetData, "id")
	if err != nil {
		return radio, err
	}
	radio.Id = radioId

	if name, err := jsonparser.GetString(targetData, "name"); err == nil {
		radio.Name = name
	}

	if picUrl, err := jsonparser.GetString(targetData, "picUrl"); err == nil {
		radio.PicUrl = picUrl
	}

	if dj, err := NewUserFromJson(json, "dj"); err == nil {
		radio.Dj = dj
	}

	// privacy as bool
	if privacy, err := jsonparser.GetBoolean(json, "privacy"); err == nil {
		radio.Privacy = privacy
	}

	return radio, nil
}
