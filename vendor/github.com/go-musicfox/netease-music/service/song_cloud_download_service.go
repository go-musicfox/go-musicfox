package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongCloudDownloadService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongCloudDownloadService) SongCloudDownload() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/cloud/dowonload",
	}
	data := make(map[string]string)
	data["songId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/cloud/dowonload", data, options)

	return code, reBody
}
