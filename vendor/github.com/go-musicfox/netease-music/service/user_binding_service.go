package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserBindingService struct {
	Uid string `json:"uid" form:"uid"`
}

func (service *UserBindingService) UserBinding() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	url := "https://music.163.com/api/v1/user/bindings/" + service.Uid
	code, reBody, _ := util.CreateRequest("POST", url, data, options)

	return code, reBody
}
