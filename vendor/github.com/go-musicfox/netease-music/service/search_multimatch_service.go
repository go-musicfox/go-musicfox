package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SearchMultimatchService struct {
	Type string `json:"type" form:"type"`
	S    string `json:"keywords" form:"keywords"`
}

func (service *SearchMultimatchService) SearchMultimatch() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Type == "" {
		service.Type = "1"
	}
	data["type"] = service.Type
	data["s"] = service.S
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/search/suggest/multimatch`, data, options)

	return code, reBody
}
