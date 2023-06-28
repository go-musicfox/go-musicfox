package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ToplistDetailService struct {
}

func (service *ToplistDetailService) ToplistDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/toplist/detail`, data, options)

	return code, reBody
}
