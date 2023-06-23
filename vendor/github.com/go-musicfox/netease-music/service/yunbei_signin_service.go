package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type YunbeiSigninService struct {
}

func (service *YunbeiSigninService) Signin() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := map[string]string{
		"type": "0",
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/point/dailyTask`, data, options)

	return code, reBody
}
