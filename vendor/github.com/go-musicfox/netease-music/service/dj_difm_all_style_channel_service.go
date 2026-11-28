package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjDifmAllStyleChannelService struct {
	Sources string `json:"sources" form:"sources"`
}

func (service *DjDifmAllStyleChannelService) DjDifmAllStyleChannel() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/dj/difm/all/style/channel/v2",
	}
	data := make(map[string]string)
	if service.Sources == "" {
		data["sources"] = "[0]"
	} else {
		data["sources"] = service.Sources
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/dj/difm/all/style/channel/v2", data, options)

	return code, reBody
}
