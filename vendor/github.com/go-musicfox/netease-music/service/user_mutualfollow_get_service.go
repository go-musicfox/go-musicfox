package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserMutualfollowGetService struct {
	Uid string `json:"uid" form:"uid"`
}

func (service *UserMutualfollowGetService) UserMutualfollowGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/user/mutualfollow/get",
	}
	data := make(map[string]string)
	data["friendid"] = service.Uid

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/user/mutualfollow/get", data, options)

	return code, reBody
}
