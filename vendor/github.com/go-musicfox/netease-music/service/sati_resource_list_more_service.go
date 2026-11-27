package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SatiResourceListMoreService struct {
	Id string `json:"id" form:"id"`
}

func (service *SatiResourceListMoreService) SatiResourceListMore() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/sati/resource/list/more/v1",
	}
	data := make(map[string]string)
	data["id"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/sati/resource/list/more/v1", data, options)

	return code, reBody
}
