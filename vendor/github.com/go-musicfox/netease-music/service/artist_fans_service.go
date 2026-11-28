package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ArtistFansService struct {
	Id     string `json:"id" form:"id"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *ArtistFansService) ArtistFans() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	if service.Limit == "" {
		data["limit"] = "20"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/artist/fans/get", data, options)

	return code, reBody
}
