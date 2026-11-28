package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongMonthdownlistService struct {
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *SongMonthdownlistService) SongMonthdownlist() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/member/song/monthdownlist",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "20"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	data["total"] = "true"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/member/song/monthdownlist", data, options)

	return code, reBody
}
