package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RecommendResourceService struct {
}

func (service *RecommendResourceService) RecommendResource() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/v1/discovery/recommend/resource`, data, options)

	return code, reBody
}
