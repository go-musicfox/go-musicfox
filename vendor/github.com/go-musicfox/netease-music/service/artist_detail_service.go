package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ArtistDetailService struct {
	Id string `json:"id" form:"id"`
}

func (service *ArtistDetailService) ArtistDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/artist/head/info/get",
	}
	data := make(map[string]string)
	data["id"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/artist/head/info/get", data, options)

	return code, reBody
}
