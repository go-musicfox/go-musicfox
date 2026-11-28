package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongLikeCheckService struct {
	Ids string `json:"ids" form:"ids"`
}

func (service *SongLikeCheckService) SongLikeCheck() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/like/check",
	}
	data := make(map[string]string)
	data["trackIds"] = service.Ids

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/like/check", data, options)

	return code, reBody
}
