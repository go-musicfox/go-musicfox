package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongDynamicCoverService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongDynamicCoverService) SongDynamicCover() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/songplay/dynamic-cover",
	}
	data := make(map[string]string)
	data["songId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/songplay/dynamic-cover", data, options)

	return code, reBody
}
