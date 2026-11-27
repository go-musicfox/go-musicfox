package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistCategoryListService struct {
	Cat   string `json:"cat" form:"cat"`
	Limit string `json:"limit" form:"limit"`
}

func (service *PlaylistCategoryListService) PlaylistCategoryList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/playlist/category/list",
	}
	data := make(map[string]string)
	if service.Cat == "" {
		data["cat"] = "全部"
	} else {
		data["cat"] = service.Cat
	}
	if service.Limit == "" {
		data["limit"] = "24"
	} else {
		data["limit"] = service.Limit
	}
	data["newStyle"] = "true"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/playlist/category/list", data, options)

	return code, reBody
}
