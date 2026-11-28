package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SearchSuggestPcService struct {
	Keyword string `json:"keyword" form:"keyword"`
}

func (service *SearchSuggestPcService) SearchSuggestPc() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/search/pc/suggest/keyword/get",
	}
	data := make(map[string]string)
	data["keyword"] = service.Keyword

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/search/pc/suggest/keyword/get", data, options)

	return code, reBody
}
