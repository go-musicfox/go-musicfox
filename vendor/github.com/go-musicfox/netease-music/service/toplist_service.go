package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ToplistService struct {
}

func (service *ToplistService) Toplist() (float64, []byte) {

	options := &util.Options{
		Crypto: "linuxapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/toplist`, data, options)

	return code, reBody
}
