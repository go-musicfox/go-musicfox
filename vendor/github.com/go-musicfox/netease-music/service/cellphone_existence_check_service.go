package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CellphoneExistenceCheckService struct {
	Cellphone   string `json:"phone" form:"phone"`
	Countrycode string `json:"countrycode" form:"countrycode"`
}

func (service *CellphoneExistenceCheckService) CellphoneExistenceCheck() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/cellphone/existence/check",
	}
	data := make(map[string]string)
	if service.Countrycode != "" {
		data["countrycode"] = service.Countrycode
	}
	data["cellphone"] = service.Cellphone

	code, reBody, _ := util.CreateRequest("POST", `http://music.163.com/eapi/cellphone/existence/check`, data, options)

	return code, reBody
}
