package service

import (
	"github.com/anhoder/netease-music/util"
)

type PlaylistDetailService struct {
	Id string `json:"id" form:"id"`
	S  string `json:"s" form:"s"`
}

func (service *PlaylistDetailService) PlaylistDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "api",
	}
	data := make(map[string]string)
	if service.S == "" {
		service.S = "8"
	}
	data["id"] = service.Id
	data["n"] = "100000"
	data["s"] = service.S

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/v6/playlist/detail`, data, options)

	return code, reBody
}
