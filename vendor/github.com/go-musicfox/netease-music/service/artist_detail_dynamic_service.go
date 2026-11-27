package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ArtistDetailDynamicService struct {
	Id string `json:"id" form:"id"`
}

func (service *ArtistDetailDynamicService) ArtistDetailDynamic() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/artist/detail/dynamic",
	}
	data := make(map[string]string)
	data["id"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/artist/detail/dynamic", data, options)

	return code, reBody
}
