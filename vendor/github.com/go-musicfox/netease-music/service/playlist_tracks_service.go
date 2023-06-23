package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type PlaylistTracksService struct {
	Op       string   `json:"op" form:"op"`
	Pid      string   `json:"pid" form:"pid"`
	TrackIds []string `json:"trackIds" form:"trackIds"`
}

func (service *PlaylistTracksService) PlaylistTracks() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["op"] = service.Op
	data["pid"] = service.Pid

	service.TrackIds = append(service.TrackIds, service.TrackIds...)
	if d, err := json.Marshal(service.TrackIds); err == nil {
		data["trackIds"] = string(d)
	}

	data["imme"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/playlist/manipulate/tracks`, data, options)

	return code, reBody
}
