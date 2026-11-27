package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongDownloadUrlService struct {
	Br string `json:"br" form:"br"`
	Id string `json:"id" form:"id"`
}

func (service *SongDownloadUrlService) SongDownloadUrl() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/enhance/download/url",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	if service.Br == "" {
		data["br"] = "999000"
	} else {
		data["br"] = service.Br
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/enhance/download/url", data, options)

	return code, reBody
}
