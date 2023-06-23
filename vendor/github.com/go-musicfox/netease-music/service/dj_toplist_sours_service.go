package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjToplistHoursService struct {
	Limit string `json:"limit" form:"limit"`
}

func (service *DjToplistHoursService) DjToplistHours() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "100"
	} else {
		data["limit"] = service.Limit
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/dj/toplist/hours`, data, options)

	return code, reBody
}
