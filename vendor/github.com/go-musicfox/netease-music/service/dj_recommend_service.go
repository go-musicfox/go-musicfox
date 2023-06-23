package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjRecommendService struct {
}

func (service *DjRecommendService) DjRecommend() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/djradio/recommend/v1`, data, options)

	return code, reBody
}
