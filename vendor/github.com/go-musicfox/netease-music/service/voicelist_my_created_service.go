package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VoicelistMyCreatedService struct {
	Limit string `json:"limit" form:"limit"`
}

func (service *VoicelistMyCreatedService) VoicelistMyCreated() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "20"
	} else {
		data["limit"] = service.Limit
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/social/my/created/voicelist/v1", data, options)

	return code, reBody
}
