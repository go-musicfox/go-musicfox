package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UgcUserDevoteService struct {
}

func (service *UgcUserDevoteService) UgcUserDevote() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/user/devote",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/user/devote", data, options)

	return code, reBody
}
