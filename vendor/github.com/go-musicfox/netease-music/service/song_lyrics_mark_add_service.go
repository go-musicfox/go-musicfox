package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongLyricsMarkAddService struct {
	Data   string `json:"data" form:"data"`
	Id     string `json:"id" form:"id"`
	MarkId string `json:"markId" form:"markId"`
}

func (service *SongLyricsMarkAddService) SongLyricsMarkAdd() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/play/lyrics/mark/add",
	}
	data := make(map[string]string)
	data["songId"] = service.Id
	if service.MarkId == "" {
		data["markId"] = ""
	} else {
		data["markId"] = service.MarkId
	}
	if service.Data == "" {
		data["data"] = "[]"
	} else {
		data["data"] = service.Data
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/play/lyrics/mark/add", data, options)

	return code, reBody
}
