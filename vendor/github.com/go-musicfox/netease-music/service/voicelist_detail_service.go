package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VoicelistDetailService struct {
	Id string `json:"id" form:"id"`
}

func (service *VoicelistDetailService) VoicelistDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/workbench/voicelist/detail",
	}
	data := make(map[string]string)
	data["id"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/workbench/voicelist/detail", data, options)

	return code, reBody
}
