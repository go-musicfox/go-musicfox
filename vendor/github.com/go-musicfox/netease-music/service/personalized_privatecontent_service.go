package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PersonalizedPrivatecontentService struct {
}

func (service *PersonalizedPrivatecontentService) PersonalizedPrivatecontent() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/personalized/privatecontent`, data, options)

	return code, reBody
}
