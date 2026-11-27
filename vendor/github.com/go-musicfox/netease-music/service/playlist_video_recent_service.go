package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistVideoRecentService struct {
}

func (service *PlaylistVideoRecentService) PlaylistVideoRecent() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/playlist/video/recent", data, options)

	return code, reBody
}
