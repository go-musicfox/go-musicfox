package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SatiResourceSubService struct {
	Cancel string `json:"cancel" form:"cancel"`
	Id     string `json:"id" form:"id"`
}

func (service *SatiResourceSubService) SatiResourceSub() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/sati/resource/sub",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	if service.Cancel == "" {
		data["cancel"] = "false"
	} else {
		data["cancel"] = service.Cancel
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/sati/resource/sub", data, options)

	return code, reBody
}
