package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RecentListenListService struct {
}

func (service *RecentListenListService) RecentListenList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/pc/recent/listen/list",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/pc/recent/listen/list", data, options)

	return code, reBody
}
