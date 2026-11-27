package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RecordRecentAlbumService struct {
	Limit string `json:"limit" form:"limit"`
}

func (service *RecordRecentAlbumService) RecordRecentAlbum() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "100"
	} else {
		data["limit"] = service.Limit
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/play-record/album/list", data, options)

	return code, reBody
}
