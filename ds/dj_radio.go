package ds

import (
    "errors"
    "github.com/buger/jsonparser"
)

type DjRadio struct {
    Id     int64  `json:"id"`
    Name   string `json:"name"`
    PicUrl string `json:"pic_url"`
    Dj     User   `json:"dj"`
}

// NewDjRadioFromJson 从Json中初始化 DjRadio
func NewDjRadioFromJson(json []byte) (DjRadio, error) {
    var radio DjRadio
    if len(json) == 0 {
        return radio, errors.New("json is empty")
    }

    radioId, err := jsonparser.GetInt(json, "id")
    if err != nil {
        return radio, err
    }
    radio.Id = radioId

    if name, err := jsonparser.GetString(json, "name"); err == nil {
        radio.Name = name
    }

    if picUrl, err := jsonparser.GetString(json, "picUrl"); err == nil {
        radio.PicUrl = picUrl
    }

    if djUserId, err := jsonparser.GetInt(json, "dj", "userId"); err == nil {
        radio.Dj.UserId = djUserId
    }

    if djName, err := jsonparser.GetString(json, "dj", "nickname"); err == nil {
        radio.Dj.Nickname = djName
    }

    if djAvatar, err := jsonparser.GetString(json, "dj", "avatarUrl"); err == nil {
        radio.Dj.AvatarUrl = djAvatar
    }

    return radio, nil
}