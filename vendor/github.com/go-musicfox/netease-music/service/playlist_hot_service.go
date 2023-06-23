package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistHotService struct{}

func (service *PlaylistHotService) PlaylistHot() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/playlist/hottags`, data, options)

	return code, reBody
}
