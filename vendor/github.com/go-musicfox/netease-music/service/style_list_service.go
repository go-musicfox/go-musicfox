package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type StyleListService struct {
}

func (service *StyleListService) StyleList() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/tag/list/get", data, options)

	return code, reBody
}
