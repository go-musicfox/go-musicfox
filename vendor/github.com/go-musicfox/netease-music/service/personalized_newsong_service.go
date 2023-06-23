package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PersonalizedNewsongService struct {
}

func (service *PersonalizedNewsongService) PersonalizedNewsong() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	data["type"] = "recommend"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/personalized/newsong`, data, options)

	return code, reBody
}
