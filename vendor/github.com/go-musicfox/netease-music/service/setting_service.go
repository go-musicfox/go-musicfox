package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SettingService struct {
}

func (service *SettingService) Setting() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/user/setting`, data, options)

	return code, reBody
}
