package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UgcArtistSearchService struct {
	Keyword string `json:"keyword" form:"keyword"`
	Limit   string `json:"limit" form:"limit"`
}

func (service *UgcArtistSearchService) UgcArtistSearch() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/artist/search",
	}
	data := make(map[string]string)
	data["keyword"] = service.Keyword
	if service.Limit == "" {
		data["limit"] = "40"
	} else {
		data["limit"] = service.Limit
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/artist/search", data, options)

	return code, reBody
}
