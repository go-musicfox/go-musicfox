package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type BroadcastChannelListService struct {
	CategoryId string `json:"categoryId" form:"categoryId"`
	LastId     string `json:"lastId" form:"lastId"`
	Limit      string `json:"limit" form:"limit"`
	RegionId   string `json:"regionId" form:"regionId"`
	Score      string `json:"score" form:"score"`
}

func (service *BroadcastChannelListService) BroadcastChannelList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/broadcast/channel/list",
	}
	data := make(map[string]string)
	if service.CategoryId == "" {
		data["categoryId"] = "0"
	} else {
		data["categoryId"] = service.CategoryId
	}
	if service.RegionId == "" {
		data["regionId"] = "0"
	} else {
		data["regionId"] = service.RegionId
	}
	if service.Limit == "" {
		data["limit"] = "20"
	} else {
		data["limit"] = service.Limit
	}
	if service.LastId == "" {
		data["lastId"] = "0"
	} else {
		data["lastId"] = service.LastId
	}
	if service.Score == "" {
		data["score"] = "-1"
	} else {
		data["score"] = service.Score
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/broadcast/channel/list", data, options)

	return code, reBody
}
