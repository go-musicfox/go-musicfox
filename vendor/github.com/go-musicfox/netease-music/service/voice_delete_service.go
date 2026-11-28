package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VoiceDeleteService struct {
	Ids string `json:"ids" form:"ids"`
}

func (service *VoiceDeleteService) VoiceDelete() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/content/voice/delete",
	}
	data := make(map[string]string)
	data["ids"] = service.Ids

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/content/voice/delete", data, options)

	return code, reBody
}
