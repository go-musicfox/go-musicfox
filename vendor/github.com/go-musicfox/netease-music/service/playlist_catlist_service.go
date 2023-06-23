package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistCatlistService struct{}

func (service *PlaylistCatlistService) PlaylistCatlist() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/playlist/catalogue`, data, options)

	return code, reBody
}
