package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjCategoryRecommendService struct {
}

func (service *DjCategoryRecommendService) DjCategoryRecommend() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `http://music.163.com/weapi/djradio/home/category/recommend`, data, options)

	return code, reBody
}
