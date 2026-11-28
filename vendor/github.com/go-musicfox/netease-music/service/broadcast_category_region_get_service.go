package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type BroadcastCategoryRegionGetService struct {
}

func (service *BroadcastCategoryRegionGetService) BroadcastCategoryRegionGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/broadcast/category/region/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/broadcast/category/region/get", data, options)

	return code, reBody
}
