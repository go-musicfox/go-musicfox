package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjToplistService struct {
	Type   string `json:"type" form:"type"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *DjToplistService) DjToplist() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "100"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	if service.Type == "hot" {
		data["type"] = "1"
	} else {
		data["type"] = "0"
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/djradio/toplist`, data, options)

	return code, reBody
}
