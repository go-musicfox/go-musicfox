package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DislikeListService struct {
}

func (service *DislikeListService) DislikeList() (float64, []byte) {

	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/music/dislike/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/music/dislike/get", data, options)

	return code, reBody
}
