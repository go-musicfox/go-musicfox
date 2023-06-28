package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SearchDefaultService struct {
}

func (service *SearchDefaultService) SearchDefault() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/search/defaultkeyword/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", `http://interface3.music.163.com/eapi/search/defaultkeyword/get`, data, options)

	return code, reBody
}
