package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type FanscenterBasicinfoGenderGetService struct {
}

func (service *FanscenterBasicinfoGenderGetService) FanscenterBasicinfoGenderGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/fanscenter/basicinfo/gender/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/fanscenter/basicinfo/gender/get", data, options)

	return code, reBody
}
