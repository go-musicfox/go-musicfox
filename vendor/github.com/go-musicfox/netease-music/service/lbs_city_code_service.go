package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type LbsCityCodeService struct {
	BizCode string `json:"bizCode" form:"bizCode"`
}

func (service *LbsCityCodeService) LbsCityCode() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/lbs/city/code",
	}
	data := make(map[string]string)
	if service.BizCode == "" {
		data["bizCode"] = ""
	} else {
		data["bizCode"] = service.BizCode
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/lbs/city/code", data, options)

	return code, reBody
}
