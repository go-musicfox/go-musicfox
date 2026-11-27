package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type BroadcastChannelCollectListService struct {
	Limit string `json:"limit" form:"limit"`
}

func (service *BroadcastChannelCollectListService) BroadcastChannelCollectList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/content/channel/collect/list",
	}
	data := make(map[string]string)
	data["contentType"] = "BROADCAST"
	if service.Limit == "" {
		data["limit"] = "99999"
	} else {
		data["limit"] = service.Limit
	}
	data["timeReverseOrder"] = "true"
	data["startDate"] = "4762584922000"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/content/channel/collect/list", data, options)

	return code, reBody
}
