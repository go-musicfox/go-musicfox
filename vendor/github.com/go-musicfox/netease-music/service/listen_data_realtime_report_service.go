package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ListenDataRealtimeReportService struct {
	Type string `json:"type" form:"type"`
}

func (service *ListenDataRealtimeReportService) ListenDataRealtimeReport() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/content/activity/listen/data/realtime/report",
	}
	data := make(map[string]string)
	if service.Type == "" {
		data["type"] = "week"
	} else {
		data["type"] = service.Type
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/content/activity/listen/data/realtime/report", data, options)

	return code, reBody
}
