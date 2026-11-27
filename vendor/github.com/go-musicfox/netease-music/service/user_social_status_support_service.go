package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserSocialStatusSupportService struct {
}

func (service *UserSocialStatusSupportService) UserSocialStatusSupport() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/social/user/status/support",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/social/user/status/support", data, options)

	return code, reBody
}
