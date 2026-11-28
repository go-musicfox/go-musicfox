package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RepUgcUserCollectVipService struct {
	ActivityId string `json:"activityId" form:"activityId"`
}

func (service *RepUgcUserCollectVipService) RepUgcUserCollectVip() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/user/collect-vip",
	}
	data := make(map[string]string)
	if service.ActivityId == "" {
		data["activityId"] = "5001"
	} else {
		data["activityId"] = service.ActivityId
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/user/collect-vip", data, options)

	return code, reBody
}
