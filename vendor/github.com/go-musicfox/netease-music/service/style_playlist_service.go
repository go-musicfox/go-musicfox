package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type StylePlaylistService struct {
	Cursor string `json:"cursor" form:"cursor"`
	Size   string `json:"size" form:"size"`
	TagId  string `json:"tagId" form:"tagId"`
}

func (service *StylePlaylistService) StylePlaylist() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Cursor == "" {
		data["cursor"] = "0"
	} else {
		data["cursor"] = service.Cursor
	}
	if service.Size == "" {
		data["size"] = "20"
	} else {
		data["size"] = service.Size
	}
	data["tagId"] = service.TagId
	data["sort"] = "0"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/style-tag/home/playlist", data, options)

	return code, reBody
}
