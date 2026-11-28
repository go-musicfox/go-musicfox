package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SatiTagListService struct {
}

func (service *SatiTagListService) SatiTagList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/sati/tag/list",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/sati/tag/list", data, options)

	return code, reBody
}
