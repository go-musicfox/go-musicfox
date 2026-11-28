package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RecommendSongsDislikeService struct {
	Id string `json:"id" form:"id"`
}

func (service *RecommendSongsDislikeService) RecommendSongsDislike() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["resId"] = service.Id
	data["resType"] = "4"
	data["sceneType"] = "1"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/v2/discovery/recommend/dislike", data, options)

	return code, reBody
}
