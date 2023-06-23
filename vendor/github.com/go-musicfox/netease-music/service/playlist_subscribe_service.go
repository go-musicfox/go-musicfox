package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistSubscribeService struct {
	T  string `json:"t" form:"t"`
	ID string `json:"id" form:"id"`
}

func (service *PlaylistSubscribeService) PlaylistSubscribe() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.ID
	if service.T == "1" {
		service.T = "subscribe"
	} else {
		service.T = "unsubscribe"
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/playlist/`+service.T, data, options)

	return code, reBody
}
