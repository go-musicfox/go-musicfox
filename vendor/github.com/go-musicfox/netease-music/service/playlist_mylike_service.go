package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistMylikeService struct {
	Limit string `json:"limit" form:"limit"`
	Time  string `json:"time" form:"time"`
}

func (service *PlaylistMylikeService) PlaylistMylike() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Time == "" {
		data["time"] = "-1"
	} else {
		data["time"] = service.Time
	}
	if service.Limit == "" {
		data["limit"] = "12"
	} else {
		data["limit"] = service.Limit
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/mlog/playlist/mylike/bytime/get", data, options)

	return code, reBody
}
