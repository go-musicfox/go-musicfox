package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SearchHotDetailService struct {
}

func (service *SearchHotDetailService) SearchHotDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/hotsearchlist/get`, data, options)

	return code, reBody
}
