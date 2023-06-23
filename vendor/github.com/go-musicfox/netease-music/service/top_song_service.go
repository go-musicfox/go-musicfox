package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type TopSongService struct {
	AreaId string `json:"type" form:"type"`
}

func (service *TopSongService) TopSong() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.AreaId == "" {
		service.AreaId = "0"
	}
	data["areaId"] = service.AreaId
	data["total"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/v1/discovery/new/songs`, data, options)

	return code, reBody
}
