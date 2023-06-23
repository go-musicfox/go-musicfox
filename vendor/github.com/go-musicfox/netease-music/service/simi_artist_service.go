package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SimiArtistService struct {
	ID string `json:"id" form:"id"`
}

func (service *SimiArtistService) SimiArtist() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.ID

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/discovery/simiArtist`, data, options)

	return code, reBody
}
