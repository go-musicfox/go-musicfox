package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CloudMatchService struct {
	Asid string `json:"asid" form:"asid"`
	Sid  string `json:"sid" form:"sid"`
	Uid  string `json:"uid" form:"uid"`
}

func (service *CloudMatchService) CloudMatch() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["userId"] = service.Uid
	data["songId"] = service.Sid
	data["adjustSongId"] = service.Asid

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/cloud/user/song/match", data, options)

	return code, reBody
}
