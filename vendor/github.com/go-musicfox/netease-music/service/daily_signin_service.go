package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DailySigninService struct {
	Type string `json:"type" form:"type"`
}

func (service *DailySigninService) DailySignin() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	if service.Type == "" {
		data["type"] = "0"
	} else {
		data["type"] = service.Type
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/point/dailyTask`, data, options)

	return code, reBody
}
