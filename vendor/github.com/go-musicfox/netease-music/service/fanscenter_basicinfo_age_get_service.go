package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type FanscenterBasicinfoAgeGetService struct {
}

func (service *FanscenterBasicinfoAgeGetService) FanscenterBasicinfoAgeGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/fanscenter/basicinfo/age/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/fanscenter/basicinfo/age/get", data, options)

	return code, reBody
}
