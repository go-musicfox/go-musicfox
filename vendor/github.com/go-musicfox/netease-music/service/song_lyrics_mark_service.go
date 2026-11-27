package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongLyricsMarkService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongLyricsMarkService) SongLyricsMark() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/play/lyrics/mark/song",
	}
	data := make(map[string]string)
	data["songId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/play/lyrics/mark/song", data, options)

	return code, reBody
}
