package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CountriesCodeListService struct {
}

func (service *CountriesCodeListService) CountriesCodeList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/lbs/countries/v1",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `http://interface3.music.163.com/eapi/lbs/countries/v1`, data, options)

	return code, reBody
}
