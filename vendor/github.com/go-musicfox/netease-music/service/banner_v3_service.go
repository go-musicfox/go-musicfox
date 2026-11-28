package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type BannerV3Service struct {
	ClientType string `json:"clientType" form:"clientType"`
}

func (service *BannerV3Service) BannerV3() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.ClientType == "" {
		data["clientType"] = "android"
	} else {
		data["clientType"] = service.ClientType
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/banner/get/v3", data, options)

	return code, reBody
}
