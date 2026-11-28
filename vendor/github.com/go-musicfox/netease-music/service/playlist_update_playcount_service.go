package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistUpdatePlaycountService struct {
	Id string `json:"id" form:"id"`
}

func (service *PlaylistUpdatePlaycountService) PlaylistUpdatePlaycount() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/playlist/update/playcount",
	}
	data := make(map[string]string)
	data["id"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/playlist/update/playcount", data, options)

	return code, reBody
}
