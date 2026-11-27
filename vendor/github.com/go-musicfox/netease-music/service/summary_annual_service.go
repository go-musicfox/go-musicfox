package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SummaryAnnualService struct {
	Year string `json:"year" form:"year"`
}

func (service *SummaryAnnualService) SummaryAnnual() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/activity/summary/annual/" + service.Year + "/${key}",
	}
	data := make(map[string]string)

	url := "https://music.163.com/eapi/activity/summary/annual/" + service.Year + "/${key}"
	code, reBody, _ := util.CreateRequest("POST", url, data, options)

	return code, reBody
}
