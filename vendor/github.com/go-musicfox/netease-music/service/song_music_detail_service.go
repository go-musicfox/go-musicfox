package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongMusicDetailService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongMusicDetailService) SongMusicDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/music/detail/get",
	}
	data := make(map[string]string)
	data["songId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/music/detail/get", data, options)

	return code, reBody
}
