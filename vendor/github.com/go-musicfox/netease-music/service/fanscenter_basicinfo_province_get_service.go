package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type FanscenterBasicinfoProvinceGetService struct {
}

func (service *FanscenterBasicinfoProvinceGetService) FanscenterBasicinfoProvinceGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/fanscenter/basicinfo/province/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/fanscenter/basicinfo/province/get", data, options)

	return code, reBody
}
