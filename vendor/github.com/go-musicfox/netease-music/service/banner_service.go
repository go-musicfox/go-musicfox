package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type BannerService struct {
	Type string `json:"type" form:"type"`
}

func (service *BannerService) Banner() (float64, []byte) {

	options := &util.Options{
		Crypto: "linuxapi",
	}

	TYPE := make(map[string]string, 6)
	TYPE["0"] = "pc"
	TYPE["1"] = "android"
	TYPE["2"] = "iphone"
	TYPE["3"] = "ipad"

	data := make(map[string]string)
	if _, ok := TYPE[service.Type]; ok {
		service.Type = TYPE[service.Type]
	} else {
		service.Type = TYPE["0"]
	}
	data["clientType"] = service.Type

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/v2/banner/get`, data, options)

	return code, reBody
}
