package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjPersonalizeRecommendService struct {
	Limit string `json:"limit" form:"limit"`
}

func (service *DjPersonalizeRecommendService) DjPersonalizeRecommend() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "6"
	} else {
		data["limit"] = service.Limit
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/djradio/personalize/rcmd", data, options)

	return code, reBody
}
