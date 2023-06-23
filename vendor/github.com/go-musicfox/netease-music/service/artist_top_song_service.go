package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ArtistTopSongService struct {
	Id string `json:"id" form:"id"`
}

func (service *ArtistTopSongService) ArtistTopSong() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	data["id"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/artist/top/song`, data, options)

	return code, reBody
}
