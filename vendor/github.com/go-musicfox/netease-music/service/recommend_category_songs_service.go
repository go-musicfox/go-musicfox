package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RecommendCategorySongsService struct {
}

func (service *RecommendCategorySongsService) RecommendCategorySongs() (float64, []byte) {

	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/homepage/category/daily/song/list",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/homepage/category/daily/song/list", data, options)

	return code, reBody
}
