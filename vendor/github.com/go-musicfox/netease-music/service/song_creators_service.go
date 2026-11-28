package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongCreatorsService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongCreatorsService) SongCreators() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/creators",
	}
	data := make(map[string]string)
	data["songId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/creators", data, options)

	return code, reBody
}
