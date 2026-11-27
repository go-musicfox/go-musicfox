package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserSocialStatusService struct {
	Uid string `json:"uid" form:"uid"`
}

func (service *UserSocialStatusService) UserSocialStatus() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/social/user/status",
	}
	data := make(map[string]string)
	data["visitorId"] = service.Uid

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/social/user/status", data, options)

	return code, reBody
}
