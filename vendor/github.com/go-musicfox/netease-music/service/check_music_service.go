package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CheckMusicService struct {
	ID string `json:"id" form:"id"`
	Br string `json:"br" form:"br"`
}

func (service *CheckMusicService) CheckMusic() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["ids"] = "[" + service.ID + "]"
	if service.Br == "" {
		service.Br = "999000"
	}
	data["br"] = service.Br
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/song/enhance/player/url`, data, options)

	return code, reBody
}
