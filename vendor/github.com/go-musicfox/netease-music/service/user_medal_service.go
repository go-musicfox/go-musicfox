package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserMedalService struct {
	Uid string `json:"uid" form:"uid"`
}

func (service *UserMedalService) UserMedal() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/medal/user/page",
	}
	data := make(map[string]string)
	data["uid"] = service.Uid

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/medal/user/page", data, options)

	return code, reBody
}
