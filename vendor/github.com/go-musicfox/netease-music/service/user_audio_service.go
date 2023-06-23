package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserAudioService struct {
	UID string `json:"uid" form:"uid"`
}

func (service *UserAudioService) UserAudio() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["userId"] = service.UID

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/djradio/get/byuser`, data, options)

	return code, reBody
}
