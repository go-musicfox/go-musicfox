package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjProgramToplistHoursService struct {
	Limit string `json:"limit" form:"limit"`
}

func (service *DjProgramToplistHoursService) DjProgramToplistHours() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/djprogram/toplist/hours`, data, options)

	return code, reBody
}
