package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MiddlePlayLotteryRemainChanceService struct {
	ActivityId string `json:"activityId" form:"activityId"`
}

func (service *MiddlePlayLotteryRemainChanceService) MiddlePlayLotteryRemainChance() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/middle/play/lottery/remain/chance",
	}
	data := make(map[string]string)
	if service.ActivityId == "" {
		data["activityId"] = "6501202"
	} else {
		data["activityId"] = service.ActivityId
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/middle/play/lottery/remain/chance", data, options)

	return code, reBody
}
