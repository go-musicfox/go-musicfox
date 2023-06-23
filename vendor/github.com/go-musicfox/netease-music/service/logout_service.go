package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type LogoutService struct {
}

func (service *LogoutService) Logout() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
		Ua:     "pc",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/logout`, data, options)

	return code, reBody
}
