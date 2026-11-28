package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RepUgcUserVipService struct {
}

func (service *RepUgcUserVipService) RepUgcUserVip() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/user/vip",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/user/vip", data, options)

	return code, reBody
}
