package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjTodayPerferedService struct {
	Page string `json:"page" form:"page"`
}

func (service *DjTodayPerferedService) DjTodayPerfered() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Page == "" {
		data["page"] = "0"
	} else {
		data["page"] = service.Page
	}
	code, reBody, _ := util.CreateRequest("POST", `http://music.163.com/weapi/djradio/home/today/perfered`, data, options)

	return code, reBody
}
