package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PersonalizedMvService struct {
	ID     string `json:"id" form:"id"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *PersonalizedMvService) PersonalizedMv() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/personalized/mv`, data, options)

	return code, reBody
}
