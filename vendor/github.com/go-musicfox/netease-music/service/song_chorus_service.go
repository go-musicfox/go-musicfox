package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongChorusService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongChorusService) SongChorus() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/chorus",
	}
	data := make(map[string]string)
	data["ids"] = "[\"" + service.Id + "\"]"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/chorus", data, options)

	return code, reBody
}
