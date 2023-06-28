package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaymodeIntelligenceListService struct {
	SongId       string `json:"id" form:"id"`
	PlaylistId   string `json:"pid" form:"pid"`
	StartMusicId string `json:"sid" form:"sid"`
	Count        string `json:"count" form:"count"`
}

func (service *PlaymodeIntelligenceListService) PlaymodeIntelligenceList() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	data["songId"] = service.SongId
	data["type"] = "fromPlayOne"
	data["playlistId"] = service.PlaylistId
	if service.StartMusicId != "" {
		data["startMusicId"] = service.StartMusicId
	} else {
		data["startMusicId"] = service.SongId
	}
	if data["count"] == "" {
		data["count"] = "1"
	} else {
		data["count"] = service.Count
	}

	code, reBody, _ := util.CreateRequest("POST", `http://music.163.com/weapi/playmode/intelligence/list`, data, options)

	return code, reBody
}
