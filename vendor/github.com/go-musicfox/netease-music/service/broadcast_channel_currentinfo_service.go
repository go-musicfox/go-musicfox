package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type BroadcastChannelCurrentinfoService struct {
	Id string `json:"id" form:"id"`
}

func (service *BroadcastChannelCurrentinfoService) BroadcastChannelCurrentinfo() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/broadcast/channel/currentinfo",
	}
	data := make(map[string]string)
	data["channelId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/broadcast/channel/currentinfo", data, options)

	return code, reBody
}
