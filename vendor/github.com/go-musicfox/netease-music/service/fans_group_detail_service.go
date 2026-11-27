package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type FansGroupDetailService struct {
	GroupId string `json:"groupId" form:"groupId"`
	Id      string `json:"id" form:"id"`
	Scene   string `json:"scene" form:"scene"`
}

func (service *FansGroupDetailService) FansGroupDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/social/fansgroup/bff/detail/get",
	}
	data := make(map[string]string)
	if service.GroupId == "" {
		data["groupId"] = service.Id
	} else {
		data["groupId"] = service.GroupId
	}
	data["scene"] = service.Scene

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/social/fansgroup/bff/detail/get", data, options)

	return code, reBody
}
