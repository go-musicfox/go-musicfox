package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SatiTimesceneResourcesGetService struct {
}

func (service *SatiTimesceneResourcesGetService) SatiTimesceneResourcesGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/sati/timescene/resources/get",
	}
	data := make(map[string]string)
	data["firstQuery"] = "false"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/sati/timescene/resources/get", data, options)

	return code, reBody
}
