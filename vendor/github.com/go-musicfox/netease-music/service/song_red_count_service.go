package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongRedCountService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongRedCountService) SongRedCount() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/red/count",
	}
	data := make(map[string]string)
	data["songId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/red/count", data, options)

	return code, reBody
}
