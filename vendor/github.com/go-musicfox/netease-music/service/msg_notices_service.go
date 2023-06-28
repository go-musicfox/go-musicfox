package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MsgNoticesService struct {
	Limit    string `json:"limit" form:"limit"`
	LastTime string `json:"lasttime" form:"lasttime"`
}

func (service *MsgNoticesService) MsgNotices() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.LastTime == "" {
		data["time"] = "-1"
	} else {
		data["time"] = service.LastTime
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/msg/notices`, data, options)

	return code, reBody
}
