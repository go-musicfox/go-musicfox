package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongDownloadUrlV1Service struct {
	Id    string `json:"id" form:"id"`
	Level string `json:"level" form:"level"`
}

func (service *SongDownloadUrlV1Service) SongDownloadUrlV1() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/enhance/download/url/v1",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	data["immerseType"] = "c51"
	data["level"] = service.Level

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/enhance/download/url/v1", data, options)

	return code, reBody
}
