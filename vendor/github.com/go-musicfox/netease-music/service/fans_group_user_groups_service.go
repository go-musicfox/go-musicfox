package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type FansGroupUserGroupsService struct {
}

func (service *FansGroupUserGroupsService) FansGroupUserGroups() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/social/fansgroup/bff/user/groups/get",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/social/fansgroup/bff/user/groups/get", data, options)

	return code, reBody
}
