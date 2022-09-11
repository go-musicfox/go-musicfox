package service

import (
	"github.com/anhoder/netease-music/util"
)

type PlaylistTracksService struct {
	Op       string `json:"op" form:"op"`
	Pid      string `json:"pid" form:"pid"`
	TrackIds string `json:"tracks" form:"tracks"`
}

func (service *PlaylistTracksService) PlaylistTracks() (float64, []byte) {

	options := &util.Options{
		Crypto:  "weapi",
	}
	data := make(map[string]string)
	data["op"] = service.Op
	data["pid"] = service.Pid
	data["trackIds"] = "[" + service.TrackIds + "]"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/playlist/manipulate/tracks`, data, options)

	return code, reBody
}
