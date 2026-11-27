package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UgcArtistGetService struct {
	Id string `json:"id" form:"id"`
}

func (service *UgcArtistGetService) UgcArtistGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/artist/get",
	}
	data := make(map[string]string)
	data["artistId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/artist/get", data, options)

	return code, reBody
}
