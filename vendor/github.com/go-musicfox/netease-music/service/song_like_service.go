package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongLikeService struct {
	Id   string `json:"id" form:"id"`
	Uid  string `json:"uid" form:"uid"`
	Like string `json:"like" form:"like"`
}

func (service *SongLikeService) SongLike() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/like",
	}
	data := make(map[string]string)
	data["trackId"] = service.Id
	data["userid"] = service.Uid
	data["like"] = "true"
	if service.Like == "false" {
		data["like"] = "false"
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/like", data, options)

	return code, reBody
}
