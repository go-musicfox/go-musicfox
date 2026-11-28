package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ThresholdDetailGetService struct {
}

func (service *ThresholdDetailGetService) ThresholdDetailGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/influencer/web/apply/threshold/detail/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/influencer/web/apply/threshold/detail/get", data, options)

	return code, reBody
}
