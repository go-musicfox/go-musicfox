package service

import (
	"encoding/json"
	"strconv"

	"github.com/go-musicfox/netease-music/util"
)

type SearchMatchService struct {
	Title    string `json:"title" form:"title"`
	Album    string `json:"album" form:"album"`
	Artist   string `json:"artist" form:"artist"`
	Duration string `json:"duration" form:"duration"`
	Md5      string `json:"md5" form:"md5"`
}

func (service *SearchMatchService) SearchMatch() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/search/match/new",
	}
	duration, _ := strconv.Atoi(service.Duration)
	songs, _ := json.Marshal([]map[string]any{{
		"title":     service.Title,
		"album":     service.Album,
		"artist":    service.Artist,
		"duration":  duration,
		"persistId": service.Md5,
	}})

	data := make(map[string]string)
	data["songs"] = string(songs)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/search/match/new", data, options)

	return code, reBody
}
