package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type AlbumListService struct {
	Area   string `json:"area" form:"area"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
	Type   string `json:"type" form:"type"`
}

func (service *AlbumListService) AlbumList() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	if service.Area == "" {
		data["area"] = "ALL"
	} else {
		data["area"] = service.Offset
	}
	data["order"] = "true"
	data["type"] = service.Type
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/vipmall/albumproduct/list`, data, options)

	return code, reBody
}
