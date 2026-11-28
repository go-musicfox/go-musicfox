package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SearchLyricService struct {
	Keyword  string `json:"keyword" form:"keyword"`
	Keywords string `json:"keywords" form:"keywords"`
	Limit    string `json:"limit" form:"limit"`
	Offset   string `json:"offset" form:"offset"`
}

func (service *SearchLyricService) SearchLyric() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Keyword == "" {
		data["keyword"] = service.Keywords
	} else {
		data["keyword"] = service.Keyword
	}
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/search/resource/lyric", data, options)

	return code, reBody
}
