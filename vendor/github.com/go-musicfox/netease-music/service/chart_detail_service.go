package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ChartDetailService struct {
	ChartCode  string `json:"chartCode" form:"chartCode"`
	TargetId   string `json:"targetId" form:"targetId"`
	TargetType string `json:"targetType" form:"targetType"`
}

func (service *ChartDetailService) ChartDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/chart/detail",
	}
	data := make(map[string]string)
	data["chartCode"] = service.ChartCode
	data["targetId"] = service.TargetId
	data["targetType"] = service.TargetType

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/chart/detail", data, options)

	return code, reBody
}
