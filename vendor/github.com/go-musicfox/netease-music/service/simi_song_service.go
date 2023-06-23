package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SimiSongService struct {
	ID     string `json:"id" form:"id"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *SimiSongService) SimiSong() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["songid"] = service.ID
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/v1/discovery/simiSong`, data, options)

	return code, reBody
}
