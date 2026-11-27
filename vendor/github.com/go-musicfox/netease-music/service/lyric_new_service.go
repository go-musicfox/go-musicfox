package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type LyricNewService struct {
	Id string `json:"id" form:"id"`
}

func (service *LyricNewService) LyricNew() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/lyric/v1",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	data["cp"] = "false"
	data["tv"] = "0"
	data["lv"] = "0"
	data["rv"] = "0"
	data["kv"] = "0"
	data["yv"] = "0"
	data["ytv"] = "0"
	data["yrv"] = "0"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/lyric/v1", data, options)

	return code, reBody
}
