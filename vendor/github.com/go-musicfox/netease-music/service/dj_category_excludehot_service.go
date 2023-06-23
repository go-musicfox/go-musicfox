package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjCategoryExcludehotService struct {
}

func (service *DjCategoryExcludehotService) DjCategoryExcludehot() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `http://music.163.com/weapi/djradio/category/excludehot`, data, options)

	return code, reBody
}
