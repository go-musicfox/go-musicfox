package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type FansGroupUserGroupDetailService struct {
	GroupId string `json:"groupId" form:"groupId"`
	Id      string `json:"id" form:"id"`
}

func (service *FansGroupUserGroupDetailService) FansGroupUserGroupDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/social/fansgroup/bff/user/group/detail/get",
	}
	data := make(map[string]string)
	if service.GroupId == "" {
		data["groupId"] = service.Id
	} else {
		data["groupId"] = service.GroupId
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/social/fansgroup/bff/user/group/detail/get", data, options)

	return code, reBody
}
