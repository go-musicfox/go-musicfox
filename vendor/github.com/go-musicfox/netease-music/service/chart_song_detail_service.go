package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ChartSongDetailService struct {
	ChartCode  string `json:"chartCode" form:"chartCode"`
	TargetId   string `json:"targetId" form:"targetId"`
	TargetType string `json:"targetType" form:"targetType"`
}

func (service *ChartSongDetailService) ChartSongDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/chart/song/detail",
	}
	data := make(map[string]string)
	data["chartCode"] = service.ChartCode
	data["targetId"] = service.TargetId
	data["targetType"] = service.TargetType

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/chart/song/detail", data, options)

	return code, reBody
}
