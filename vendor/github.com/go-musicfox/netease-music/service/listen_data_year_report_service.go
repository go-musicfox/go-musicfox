package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ListenDataYearReportService struct {
}

func (service *ListenDataYearReportService) ListenDataYearReport() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/content/activity/listen/data/year/report",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/content/activity/listen/data/year/report", data, options)

	return code, reBody
}
