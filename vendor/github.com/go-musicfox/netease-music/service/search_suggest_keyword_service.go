package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SearchSuggestKeywordService struct {
	Keyword  string `json:"keyword" form:"keyword"`
	Keywords string `json:"keywords" form:"keywords"`
}

func (service *SearchSuggestKeywordService) SearchSuggestKeyword() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Keyword == "" {
		data["keyword"] = service.Keywords
	} else {
		data["keyword"] = service.Keyword
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/search/suggest/keyword/get", data, options)

	return code, reBody
}
