package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VoicelistSearchService struct {
	Keyword string `json:"keyword" form:"keyword"`
	Limit   string `json:"limit" form:"limit"`
	Offset  string `json:"offset" form:"offset"`
}

func (service *VoicelistSearchService) VoicelistSearch() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/search/voicelist/get",
	}
	data := make(map[string]string)
	if service.Keyword == "" {
		data["keyword"] = ""
	} else {
		data["keyword"] = service.Keyword
	}
	data["scene"] = "normal"
	if service.Limit == "" {
		data["limit"] = "10"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "30"
	} else {
		data["offset"] = service.Offset
	}
	data["e_r"] = "true"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/search/voicelist/get", data, options)

	return code, reBody
}
