package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MvFirstService struct {
	Area  string `json:"area" form:"area"`
	Limit string `json:"limit" form:"limit"`
}

func (service *MvFirstService) MvFirst() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["area"] = service.Area
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}

	data["order"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://interface.music.163.com/weapi/mv/first`, data, options)

	return code, reBody
}
