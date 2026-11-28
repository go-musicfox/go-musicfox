package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ToplistDetailV2Service struct {
}

func (service *ToplistDetailV2Service) ToplistDetailV2() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/toplist/detail/v2", data, options)

	return code, reBody
}
