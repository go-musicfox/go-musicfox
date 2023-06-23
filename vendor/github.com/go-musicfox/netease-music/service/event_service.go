package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type EventService struct {
	PageSize string `json:"pagesize" form:"pagesize"`
	LastTime string `json:"lasttime" form:"lasttime"`
}

func (service *EventService) Event() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.PageSize == "" {
		data["pagesize"] = "20"
	} else {
		data["pagesize"] = service.PageSize
	}
	if service.LastTime == "" {
		data["lasttime"] = "-1"
	} else {
		data["lasttime"] = service.LastTime
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/v1/event/get`, data, options)

	return code, reBody
}
