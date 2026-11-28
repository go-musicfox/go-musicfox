package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjDifmChannelSubscribeService struct {
	Id string `json:"id" form:"id"`
}

func (service *DjDifmChannelSubscribeService) DjDifmChannelSubscribe() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/dj/difm/channel/subscribe",
	}
	data := make(map[string]string)
	data["id"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/dj/difm/channel/subscribe", data, options)

	return code, reBody
}
