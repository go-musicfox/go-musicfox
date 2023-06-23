package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ArtistSubService struct {
	T  string `json:"t" form:"t"`
	Id string `json:"id" form:"id"`
}

func (service *ArtistSubService) ArtistSub() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	if service.T == "1" {
		service.T = "sub"
	} else {
		service.T = "unsub"
	}

	data["artistId"] = service.Id
	data["artistIds"] = "[" + service.Id + "]"

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/artist/`+service.T, data, options)

	return code, reBody
}
