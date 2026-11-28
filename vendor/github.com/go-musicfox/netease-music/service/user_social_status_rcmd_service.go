package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserSocialStatusRcmdService struct {
}

func (service *UserSocialStatusRcmdService) UserSocialStatusRcmd() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/social/user/status/rcmd",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/social/user/status/rcmd", data, options)

	return code, reBody
}
