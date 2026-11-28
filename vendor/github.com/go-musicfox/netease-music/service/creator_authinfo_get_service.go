package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CreatorAuthinfoGetService struct {
}

func (service *CreatorAuthinfoGetService) CreatorAuthinfoGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/user/creator/authinfo/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/user/creator/authinfo/get", data, options)

	return code, reBody
}
