package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VoiceDetailService struct {
	Id string `json:"id" form:"id"`
}

func (service *VoiceDetailService) VoiceDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/workbench/voice/detail",
	}
	data := make(map[string]string)
	data["id"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/workbench/voice/detail", data, options)

	return code, reBody
}
