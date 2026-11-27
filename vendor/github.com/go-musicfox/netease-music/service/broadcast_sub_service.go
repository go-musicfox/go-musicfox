package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type BroadcastSubService struct {
	Id string `json:"id" form:"id"`
	T  string `json:"t" form:"t"`
}

func (service *BroadcastSubService) BroadcastSub() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/content/interact/collect",
	}
	data := make(map[string]string)
	data["contentType"] = "BROADCAST"
	data["contentId"] = service.Id
	data["cancelCollect"] = service.T

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/content/interact/collect", data, options)

	return code, reBody
}
