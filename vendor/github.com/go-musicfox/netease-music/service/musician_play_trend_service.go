package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MusicianPlayTrendService struct {
	EndTime   string `json:"endTime" form:"endTime"`
	StartTime string `json:"startTime" form:"startTime"`
}

func (service *MusicianPlayTrendService) MusicianPlayTrend() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["startTime"] = service.StartTime
	data["endTime"] = service.EndTime

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/creator/musician/play/count/statistic/data/trend/get", data, options)

	return code, reBody
}
