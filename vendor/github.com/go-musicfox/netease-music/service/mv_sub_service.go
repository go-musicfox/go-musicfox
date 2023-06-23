package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MvSubService struct {
	T    string `json:"t" form:"t"`
	MvId string `json:"mvid" form:"mvid"`
}

func (service *MvSubService) MvSub() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	if service.T == "1" {
		service.T = "sub"
	} else {
		service.T = "unsub"
	}

	data["mvId"] = service.MvId
	data["mvIds"] = "[" + service.MvId + "]"

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/mv/`+service.T, data, options)

	return code, reBody
}
