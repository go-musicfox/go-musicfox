package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ArtistMvService struct {
	ID     string `json:"id" form:"id"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *ArtistMvService) ArtistMv() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["artistId"] = service.ID
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
	data["total"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/artist/mvs`, data, options)

	return code, reBody
}
