package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UgcMvGetService struct {
	Id string `json:"id" form:"id"`
}

func (service *UgcMvGetService) UgcMvGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/mv/get",
	}
	data := make(map[string]string)
	data["mvId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/mv/get", data, options)

	return code, reBody
}
