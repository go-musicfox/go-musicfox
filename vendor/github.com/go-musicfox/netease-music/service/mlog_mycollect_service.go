package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MlogMycollectService struct {
	Time  string `json:"time" form:"time"`
	Limit string `json:"limit" form:"limit"`
	Type  string `json:"type" form:"type"`
}

func (service *MlogMycollectService) MlogMycollect() (float64, []byte) {

	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/mlog/mycollect/bytime/get",
	}
	data := make(map[string]string)
	if service.Time == "" {
		data["time"] = "-1"
	} else {
		data["time"] = service.Time
	}
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Type == "" {
		data["mlogType"] = "3"
	} else {
		data["mlogType"] = service.Type
	}

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/mlog/mycollect/bytime/get", data, options)

	return code, reBody
}
