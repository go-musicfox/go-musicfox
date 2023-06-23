package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongOrderUpdateService struct {
	Pid string `json:"pid" form:"pid"`
	Ids string `json:"ids" form:"ids"`
}

func (service *SongOrderUpdateService) SongOrderUpdate() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
		Url:    "/api/playlist/desc/update",
	}
	data := make(map[string]string)
	data["pid"] = service.Pid
	data["trackIds"] = service.Ids
	data["op"] = "update"
	code, reBody, _ := util.CreateRequest("POST", `http://interface.music.163.com/api/playlist/manipulate/tracks`, data, options)

	return code, reBody
}
