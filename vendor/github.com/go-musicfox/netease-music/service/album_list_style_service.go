package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type AlbumListStyleService struct {
	Area   string `json:"area" form:"area"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *AlbumListStyleService) AlbumListStyle() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "10"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	if service.Area == "" {
		service.Area = "Z_H"
	}
	data["order"] = "true"
	data["area"] = service.Area
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/vipmall/appalbum/album/style`, data, options)

	return code, reBody
}
