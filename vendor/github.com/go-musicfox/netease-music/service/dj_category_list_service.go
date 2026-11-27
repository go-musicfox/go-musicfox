package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjCategoryListService struct {
}

func (service *DjCategoryListService) DjCategoryList() (float64, []byte) {

	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/dj/radio/category/list",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/dj/radio/category/list", data, options)

	return code, reBody
}
