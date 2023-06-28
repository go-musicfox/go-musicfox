package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SearchSuggestService struct {
	S    string `json:"keywords" form:"keywords"`
	Type string `json:"type" form:"type"`
}

func (service *SearchSuggestService) SearchSuggest() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	if service.Type == "mobile" {
		service.Type = "keyword"
	} else {
		service.Type = "web"
	}
	data["s"] = service.S
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/search/suggest/`+service.Type, data, options)

	return code, reBody
}
