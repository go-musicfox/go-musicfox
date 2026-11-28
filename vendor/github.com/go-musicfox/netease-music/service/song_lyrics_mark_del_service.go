package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongLyricsMarkDelService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongLyricsMarkDelService) SongLyricsMarkDel() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/play/lyrics/mark/del",
	}
	data := make(map[string]string)
	data["markIds"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/play/lyrics/mark/del", data, options)

	return code, reBody
}
