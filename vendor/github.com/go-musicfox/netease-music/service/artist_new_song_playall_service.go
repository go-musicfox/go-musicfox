package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ArtistNewSongPlayallService struct {
}

func (service *ArtistNewSongPlayallService) ArtistNewSongPlayall() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/sub/artist/new/works/song/playall",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/sub/artist/new/works/song/playall", data, options)

	return code, reBody
}
