package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ListenDataSongPlayRankService struct {
	Type    string `json:"type" form:"type"`
	EndTime string `json:"endTime" form:"endTime"`
}

func (service *ListenDataSongPlayRankService) ListenDataSongPlayRank() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/content/activity/listen/data/song/play/rank",
	}
	data := make(map[string]string)
	if service.Type == "" {
		data["type"] = "month"
	} else {
		data["type"] = service.Type
	}
	if service.EndTime != "" {
		data["endTime"] = service.EndTime
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/content/activity/listen/data/song/play/rank", data, options)

	return code, reBody
}
