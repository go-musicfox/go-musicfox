package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MiddlePlayDoLotteryService struct {
	ActivityId string `json:"activityId" form:"activityId"`
	DrawCount  string `json:"drawCount" form:"drawCount"`
	Token      string `json:"token" form:"token"`
}

func (service *MiddlePlayDoLotteryService) MiddlePlayDoLottery() (float64, []byte) {

	options := &util.Options{
		Crypto:     "eapi",
		Url:        "/api/middle/play/do/lottery",
		CheckToken: "v2",
		Token:      service.Token,
	}
	data := make(map[string]string)
	if service.ActivityId == "" {
		data["activityId"] = "6501202"
	} else {
		data["activityId"] = service.ActivityId
	}
	if service.DrawCount == "" {
		data["drawCount"] = "1"
	} else {
		data["drawCount"] = service.DrawCount
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/middle/play/do/lottery", data, options)

	return code, reBody
}
