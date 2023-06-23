package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserDjService struct {
	Uid    string `json:"uid" form:"uid"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *UserDjService) UserDj() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["uid"] = service.Uid
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/dj/program/`+service.Uid, data, options)

	return code, reBody
}
