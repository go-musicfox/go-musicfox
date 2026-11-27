package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjSubscriberService struct {
	Id    string `json:"id" form:"id"`
	Limit string `json:"limit" form:"limit"`
	Time  string `json:"time" form:"time"`
}

func (service *DjSubscriberService) DjSubscriber() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Time == "" {
		data["time"] = "-1"
	} else {
		data["time"] = service.Time
	}
	data["id"] = service.Id
	if service.Limit == "" {
		data["limit"] = "20"
	} else {
		data["limit"] = service.Limit
	}
	data["total"] = "true"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/djradio/subscriber", data, options)

	return code, reBody
}
