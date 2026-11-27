package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RepUgcUserGetService struct {
}

func (service *RepUgcUserGetService) RepUgcUserGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/user/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/user/get", data, options)

	return code, reBody
}
