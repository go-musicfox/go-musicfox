package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SatiResourceListService struct {
	Tag string `json:"tag" form:"tag"`
}

func (service *SatiResourceListService) SatiResourceList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/sati/resource/list",
	}
	data := make(map[string]string)
	data["tag"] = service.Tag
	data["firstQuery"] = "false"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/sati/resource/list", data, options)

	return code, reBody
}
