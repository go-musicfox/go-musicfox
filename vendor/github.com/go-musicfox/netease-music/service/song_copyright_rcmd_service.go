package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongCopyrightRcmdService struct {
	Id     string `json:"id" form:"id"`
	Songid string `json:"songid" form:"songid"`
}

func (service *SongCopyrightRcmdService) SongCopyrightRcmd() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/copyright/rcmd",
	}
	data := make(map[string]string)
	data["songid"] = service.Songid
	if service.Songid == "" {
		data["songid"] = service.Id
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/copyright/rcmd", data, options)

	return code, reBody
}
