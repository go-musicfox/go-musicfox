package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SatiResourceSubListService struct {
}

func (service *SatiResourceSubListService) SatiResourceSubList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/sati/resource/sub/list",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/sati/resource/sub/list", data, options)

	return code, reBody
}
