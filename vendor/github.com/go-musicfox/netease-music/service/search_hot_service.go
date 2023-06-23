package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SearchHotService struct {
}

func (service *SearchHotService) SearchHot() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
		Ua:     "mobile",
	}
	data := make(map[string]string)
	data["type"] = "1111"

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/search/hot`, data, options)

	return code, reBody
}
