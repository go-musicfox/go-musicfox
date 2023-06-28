package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserCloudDelService struct {
	ID string `json:"id" form:"id"`
}

func (service *UserCloudDelService) UserCloudDel() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["songIds"] = "[" + service.ID + "]"
	code, reBody, _ := util.CreateRequest("POST", `http://music.163.com/weapi/cloud/del`, data, options)

	return code, reBody
}
