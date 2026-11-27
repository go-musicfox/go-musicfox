package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RecommendCategoryConfigsService struct {
}

func (service *RecommendCategoryConfigsService) RecommendCategoryConfigs() (float64, []byte) {

	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/homepage/daily/song/config/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/homepage/daily/song/config/get", data, options)

	return code, reBody
}
