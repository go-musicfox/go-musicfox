package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaymodeSongVectorService struct {
	Ids string `json:"ids" form:"ids"`
}

func (service *PlaymodeSongVectorService) PlaymodeSongVector() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/playmode/song/vector/get",
	}
	data := make(map[string]string)
	data["ids"] = service.Ids

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/playmode/song/vector/get", data, options)

	return code, reBody
}
