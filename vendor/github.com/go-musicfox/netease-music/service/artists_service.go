package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ArtistsService struct {
	ID string `json:"id" form:"id"`
}

func (service *ArtistsService) Artists() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.ID

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/v1/artist/`+service.ID, data, options)

	return code, reBody
}
