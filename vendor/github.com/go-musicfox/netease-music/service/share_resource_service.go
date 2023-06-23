package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ShareResourceService struct {
	Id   string `json:"id" form:"id"`
	Msg  string `json:"msg" form:"msg"`
	Type string `json:"type" form:"type"`
}

func (service *ShareResourceService) ShareResource() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	data["msg"] = service.Msg

	if service.Type == "" {
		data["type"] = "song"
	} else {
		data["type"] = service.Type
	}
	code, reBody, _ := util.CreateRequest("POST", `http://music.163.com/weapi/share/friends/resource`, data, options)

	return code, reBody
}
