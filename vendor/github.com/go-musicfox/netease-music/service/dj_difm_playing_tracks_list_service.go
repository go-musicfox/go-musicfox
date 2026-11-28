package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjDifmPlayingTracksListService struct {
	ChannelId string `json:"channelId" form:"channelId"`
	Limit     string `json:"limit" form:"limit"`
	Source    string `json:"source" form:"source"`
}

func (service *DjDifmPlayingTracksListService) DjDifmPlayingTracksList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/dj/difm/playing/tracks/list",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "5"
	} else {
		data["limit"] = service.Limit
	}
	if service.Source == "" {
		data["source"] = "0"
	} else {
		data["source"] = service.Source
	}
	data["channelId"] = service.ChannelId

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/dj/difm/playing/tracks/list", data, options)

	return code, reBody
}
