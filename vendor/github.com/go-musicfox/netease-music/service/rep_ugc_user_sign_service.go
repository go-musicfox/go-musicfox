package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RepUgcUserSignService struct {
}

func (service *RepUgcUserSignService) RepUgcUserSign() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/user/sign",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/user/sign", data, options)

	return code, reBody
}
