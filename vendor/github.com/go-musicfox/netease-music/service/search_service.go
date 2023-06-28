package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SearchService struct {
	S      string `json:"keywords" form:"keywords"`
	Type   string `json:"type" form:"type"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *SearchService) Search() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}

	if service.Type == "" {
		service.Type = "1"
	}
	if service.Limit == "" {
		service.Limit = "30"
	}
	if service.Offset == "" {
		service.Offset = "0"
	}
	data := make(map[string]string)
	data["limit"] = service.Limit
	data["offset"] = service.Offset

	if service.Type == "2000" {
		data["keyword"] = service.S
		data["scene"] = "normal"
		code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/search/voice/get`, data, options)
		return code, reBody
	}

	data["type"] = service.Type
	data["s"] = service.S

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/cloudsearch/pc`, data, options)

	return code, reBody
}
