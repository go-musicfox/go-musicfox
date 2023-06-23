package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserSubcountService struct {
}

func (service *UserSubcountService) UserSubcount() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/subcount`, data, options)

	return code, reBody
}
