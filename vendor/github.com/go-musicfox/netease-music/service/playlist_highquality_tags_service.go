package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistHighqualityTagsService struct {
}

func (service *PlaylistHighqualityTagsService) PlaylistHighqualityTags() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/playlist/highquality/tags", data, options)

	return code, reBody
}
