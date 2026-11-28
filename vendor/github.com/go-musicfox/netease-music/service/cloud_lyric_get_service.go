package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CloudLyricGetService struct {
	Sid string `json:"sid" form:"sid"`
	Uid string `json:"uid" form:"uid"`
}

func (service *CloudLyricGetService) CloudLyricGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/cloud/lyric/get",
	}
	data := make(map[string]string)
	data["userId"] = service.Uid
	data["songId"] = service.Sid
	data["lv"] = "-1"
	data["kv"] = "-1"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/cloud/lyric/get", data, options)

	return code, reBody
}
