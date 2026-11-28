package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RepUgcActivityCollectService struct {
	ActivityId string `json:"activityId" form:"activityId"`
}

func (service *RepUgcActivityCollectService) RepUgcActivityCollect() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/activity/collect",
	}
	data := make(map[string]string)
	if service.ActivityId == "" {
		data["activityId"] = "5001"
	} else {
		data["activityId"] = service.ActivityId
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/activity/collect", data, options)

	return code, reBody
}
