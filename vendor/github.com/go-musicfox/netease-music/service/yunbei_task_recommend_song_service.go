package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type YunbeiTaskRecommendSongService struct {
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *YunbeiTaskRecommendSongService) YunbeiTaskRecommendSong() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	if service.Limit == "" {
		data["limit"] = "10"
	} else {
		data["limit"] = service.Limit
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/ad/power/yunbei/distribution/recommend/song", data, options)

	return code, reBody
}
