package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserCloudDetailService struct {
	ID string `json:"id" form:"id"`
}

func (service *UserCloudDetailService) UserCloudDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["songIds"] = "[" + service.ID + "]"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/v1/cloud/get/byids`, data, options)

	return code, reBody
}
