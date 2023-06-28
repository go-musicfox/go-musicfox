package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SimiMvService struct {
	ID     string `json:"mvid" form:"mvid"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *SimiMvService) SimiMv() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["mvid"] = service.ID
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/discovery/simiMV`, data, options)

	return code, reBody
}
