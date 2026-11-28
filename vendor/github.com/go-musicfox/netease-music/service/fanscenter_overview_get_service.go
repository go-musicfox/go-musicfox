package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type FanscenterOverviewGetService struct {
}

func (service *FanscenterOverviewGetService) FanscenterOverviewGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/fanscenter/overview/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/fanscenter/overview/get", data, options)

	return code, reBody
}
