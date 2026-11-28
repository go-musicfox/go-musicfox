package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VoicelistListService struct {
	Limit       string `json:"limit" form:"limit"`
	Offset      string `json:"offset" form:"offset"`
	VoiceListId string `json:"voiceListId" form:"voiceListId"`
}

func (service *VoicelistListService) VoicelistList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/workbench/voices/by/voicelist",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "200"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	data["voiceListId"] = service.VoiceListId

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/workbench/voices/by/voicelist", data, options)

	return code, reBody
}
