package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type StyleAlbumService struct {
	Cursor string `json:"cursor" form:"cursor"`
	Size   string `json:"size" form:"size"`
	Sort   string `json:"sort" form:"sort"`
	TagId  string `json:"tagId" form:"tagId"`
}

func (service *StyleAlbumService) StyleAlbum() (float64, []byte) {

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
	if service.Sort == "" {
		data["sort"] = "0"
	} else {
		data["sort"] = service.Sort
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/style-tag/home/album", data, options)

	return code, reBody
}
