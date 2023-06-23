package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserEventService struct {
	Uid   string `json:"uid" form:"uid"`
	Limit string `json:"limit" form:"limit"`
	Time  string `json:"lasttime " form:"lasttime "`
}

func (service *UserEventService) UserEvent() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["getcounts"] = "true"
	data["total"] = "false"
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Time == "" {
		data["time"] = "-1"
	} else {
		data["time"] = service.Time
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/event/get/`+service.Uid, data, options)

	return code, reBody
}
