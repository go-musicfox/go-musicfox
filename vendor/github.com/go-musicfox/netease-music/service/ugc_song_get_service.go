package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UgcSongGetService struct {
	Id string `json:"id" form:"id"`
}

func (service *UgcSongGetService) UgcSongGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/song/get",
	}
	data := make(map[string]string)
	data["songId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/song/get", data, options)

	return code, reBody
}
