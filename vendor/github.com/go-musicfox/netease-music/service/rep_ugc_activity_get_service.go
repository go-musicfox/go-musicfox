package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RepUgcActivityGetService struct {
}

func (service *RepUgcActivityGetService) RepUgcActivityGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/activity/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/activity/get", data, options)

	return code, reBody
}
